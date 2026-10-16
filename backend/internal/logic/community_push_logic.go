package logic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"novablog/enum"
	"novablog/internal/model"
	"novablog/pkg/novapush"

	"go.uber.org/zap"
)

// 推送重试上限：超过后停止自动重试，等待博主手动处理。
const communityPushMaxRetry = 5

// CommunityPushLogic CMS 内容推送到 NovaBlog 社区的核心服务。
type CommunityPushLogic struct {
	syncModel   *model.CommunitySyncRecordModel
	configModel *model.CommunityPushConfigModel
}

func NewCommunityPushLogic() *CommunityPushLogic {
	return &CommunityPushLogic{
		syncModel:   model.NewCommunitySyncRecord(),
		configModel: model.NewCommunityPushConfig(),
	}
}

// SyncActionResult 推送/撤回动作结果。
type SyncActionResult struct {
	Result       string `json:"result"`       // created | updated | skipped | hidden | deleted
	CommunityID  uint   `json:"community_id"` // 社区侧 article_id
	SyncStatus   string `json:"sync_status"`  // synced | hidden | failed | deleted
	ContentTitle string `json:"content_title"`
	CommunityURL string `json:"community_url,omitempty"`
}

// resolvePushContext 解析推送配置并归一化地址；未配置时返回业务错误。
func (l *CommunityPushLogic) resolvePushContext() (baseURL, pushKey string, err error) {
	base := OfficialSiteBaseURL()
	key := getCommunityPushKey()
	if base == "" || strings.TrimSpace(key) == "" {
		return "", "", enum.ErrCommunityPushNotConfigured
	}
	baseURL, err = novapush.NormalizeBaseURL(base)
	if err != nil {
		return "", "", enum.ErrCommunityPushBaseURLInvalid
	}
	return baseURL, key, nil
}

// PushToCommunity 推送/更新内容到社区（opt-in 的首次推送与手动重推共用）。
// 内容与状态均无变化（hash 一致且上次成功）时跳过实际调用。
func (l *CommunityPushLogic) PushToCommunity(ctx context.Context, targetType, targetID string) (*SyncActionResult, error) {
	baseURL, pushKey, err := l.resolvePushContext()
	if err != nil {
		return nil, err
	}

	built, err := BuildCommunityPayload(ctx, targetType, targetID)
	if err != nil {
		l.markFailure(ctx, targetType, targetID, "", err)
		return nil, mapPushError(err)
	}
	payload := built.Payload

	hash, err := payloadHash(payload)
	if err != nil {
		return nil, err
	}
	existing, err := l.syncModel.GetByTarget(ctx, targetType, targetID)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.SyncStatus == model.CommunitySyncStatusSynced && existing.ContentHash == hash {
		return &SyncActionResult{
			Result:       "skipped",
			CommunityID:  existing.CommunityArticleID,
			SyncStatus:   existing.SyncStatus,
			ContentTitle: payload.Title,
		}, nil
	}

	result, err := novapush.UpsertArticle(ctx, baseURL, pushKey, payload)
	if err != nil {
		l.markFailure(ctx, targetType, targetID, payload.Title, err)
		return nil, mapPushError(err)
	}

	now := time.Now()
	record := &model.CommunitySyncRecord{
		TargetType:         targetType,
		TargetID:           targetID,
		ContentTitle:       payload.Title,
		CommunityArticleID: result.ArticleID,
		SyncStatus:         model.CommunitySyncStatusSynced,
		ContentHash:        hash,
		RetryCount:         0,
		PushedAt:           &now,
	}
	if payload.Status == "hidden" {
		record.SyncStatus = model.CommunitySyncStatusHidden
	}
	if err := l.syncModel.UpsertByTarget(ctx, record); err != nil {
		communityPushLog().Error("写入社区同步记录失败", zap.String("target", targetType+"/"+targetID), zap.Error(err))
	}
	return &SyncActionResult{
		Result:       result.Result,
		CommunityID:  result.ArticleID,
		SyncStatus:   record.SyncStatus,
		ContentTitle: payload.Title,
	}, nil
}

// HideFromCommunity 从社区下架（status=hidden，保留社区数据：评论/阅读数不丢）。
func (l *CommunityPushLogic) HideFromCommunity(ctx context.Context, targetType, targetID string) (*SyncActionResult, error) {
	baseURL, pushKey, err := l.resolvePushContext()
	if err != nil {
		return nil, err
	}

	built, err := BuildCommunityPayload(ctx, targetType, targetID)
	if err != nil {
		// 内容已不存在：直接走删除同步
		return l.RemoveFromCommunity(ctx, targetType, targetID)
	}
	built.Payload.Status = "hidden"

	result, err := novapush.UpsertArticle(ctx, baseURL, pushKey, built.Payload)
	if err != nil {
		l.markFailure(ctx, targetType, targetID, built.Payload.Title, err)
		return nil, mapPushError(err)
	}

	if err := l.markStatus(ctx, targetType, targetID, model.CommunitySyncStatusHidden, built.Payload, result.ArticleID); err != nil {
		communityPushLog().Error("写入社区同步记录失败", zap.Error(err))
	}
	return &SyncActionResult{
		Result:       "hidden",
		CommunityID:  result.ArticleID,
		SyncStatus:   model.CommunitySyncStatusHidden,
		ContentTitle: built.Payload.Title,
	}, nil
}

// RemoveFromCommunity 从社区删除（软删同步，幂等）。
func (l *CommunityPushLogic) RemoveFromCommunity(ctx context.Context, targetType, targetID string) (*SyncActionResult, error) {
	baseURL, pushKey, err := l.resolvePushContext()
	if err != nil {
		return nil, err
	}

	result, err := novapush.DeleteArticle(ctx, baseURL, pushKey, targetID)
	if err != nil {
		l.markFailure(ctx, targetType, targetID, "", err)
		return nil, mapPushError(err)
	}

	if err := l.markStatusWithArgs(ctx, targetType, targetID, model.CommunitySyncStatusDeleted, uint(0)); err != nil {
		communityPushLog().Error("写入社区同步记录失败", zap.Error(err))
	}
	return &SyncActionResult{
		Result:      result.Result,
		CommunityID: 0,
		SyncStatus:  model.CommunitySyncStatusDeleted,
	}, nil
}

// SyncOnUpdated 自动同步钩子：内容创建/更新/状态变更后调用。
// 仅对已开启推送（存在同步记录且非 deleted）的内容生效，按当前 CMS 状态映射 published/hidden。
func (l *CommunityPushLogic) SyncOnUpdated(ctx context.Context, targetType, targetID string) {
	record, err := l.syncModel.GetByTarget(ctx, targetType, targetID)
	if err != nil {
		communityPushLog().Error("读取社区同步记录失败", zap.Error(err))
		return
	}
	if record == nil || record.SyncStatus == model.CommunitySyncStatusDeleted {
		return
	}
	syncCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	if _, err := l.PushToCommunity(syncCtx, targetType, targetID); err != nil {
		communityPushLog().Warn("社区自动同步失败（已记录，等待重试）",
			zap.String("target", targetType+"/"+targetID), zap.Error(err))
	}
}

// SyncOnDeleted 自动同步钩子：内容删除后调用，向社区发起删除同步。
func (l *CommunityPushLogic) SyncOnDeleted(ctx context.Context, targetType, targetID string) {
	record, err := l.syncModel.GetByTarget(ctx, targetType, targetID)
	if err != nil {
		communityPushLog().Error("读取社区同步记录失败", zap.Error(err))
		return
	}
	if record == nil || record.SyncStatus == model.CommunitySyncStatusDeleted {
		return
	}
	syncCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	if _, err := l.RemoveFromCommunity(syncCtx, targetType, targetID); err != nil {
		communityPushLog().Warn("社区删除同步失败（已记录）",
			zap.String("target", targetType+"/"+targetID), zap.Error(err))
	}
}

// RetryFailed 重试 worker：重推失败记录（重建 payload，始终推最新内容）。
func (l *CommunityPushLogic) RetryFailed(ctx context.Context) (retried int, err error) {
	records, err := l.syncModel.GetFailed(ctx, communityPushMaxRetry, 20)
	if err != nil {
		return 0, err
	}
	for _, record := range records {
		retryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		if _, err := l.PushToCommunity(retryCtx, record.TargetType, record.TargetID); err != nil {
			cancel()
			// 单条失败继续处理下一条
			continue
		}
		cancel()
		retried++
	}
	return retried, nil
}

// GetCommunityTypes 代理社区内容类型字典（推送配置页展示与校验用）。
func (l *CommunityPushLogic) GetCommunityTypes(ctx context.Context) ([]novapush.ArticleTypeInfo, error) {
	baseURL, pushKey, err := l.resolvePushContext()
	if err != nil {
		return nil, err
	}
	types, err := novapush.Types(ctx, baseURL, pushKey)
	if err != nil {
		return types, mapPushError(err)
	}
	return types, nil
}

// —— 状态查询 ——

// CommunitySyncState 单个内容的推送状态（供按钮渲染）。
type CommunitySyncState struct {
	TargetType         string `json:"target_type"`
	TargetID           string `json:"target_id"`
	SyncStatus         string `json:"sync_status"` // none | synced | hidden | failed | deleted
	CommunityArticleID uint   `json:"community_article_id"`
	LastError          string `json:"last_error,omitempty"`
	PushedAt           string `json:"pushed_at,omitempty"`
}

func (l *CommunityPushLogic) GetState(ctx context.Context, targetType, targetID string) (*CommunitySyncState, error) {
	record, err := l.syncModel.GetByTarget(ctx, targetType, targetID)
	if err != nil {
		return nil, err
	}
	return recordToState(targetType, targetID, record), nil
}

func (l *CommunityPushLogic) GetStates(ctx context.Context, targetType string, targetIDs []string) (map[string]*CommunitySyncState, error) {
	records, err := l.syncModel.GetListByTargets(ctx, targetType, targetIDs)
	if err != nil {
		return nil, err
	}
	states := make(map[string]*CommunitySyncState, len(records))
	for i := range records {
		state := recordToState(targetType, records[i].TargetID, &records[i])
		states[records[i].TargetID] = state
	}
	return states, nil
}

// ListRecords 同步记录分页。
type CommunitySyncRecordItem struct {
	CommunitySyncState
	ContentTitle string `json:"content_title"`
	RetryCount   int    `json:"retry_count"`
	UpdatedAt    string `json:"updated_at"`
}

func (l *CommunityPushLogic) ListRecords(ctx context.Context, page, size int, targetType, syncStatus string) ([]CommunitySyncRecordItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	records, total, err := l.syncModel.GetList(ctx, page, size, targetType, syncStatus)
	if err != nil {
		return nil, 0, err
	}
	items := make([]CommunitySyncRecordItem, 0, len(records))
	for i := range records {
		state := recordToState(records[i].TargetType, records[i].TargetID, &records[i])
		items = append(items, CommunitySyncRecordItem{
			CommunitySyncState: *state,
			ContentTitle:       records[i].ContentTitle,
			RetryCount:         records[i].RetryCount,
			UpdatedAt:          records[i].UpdatedAt.Format("2006-01-02 15:04"),
		})
	}
	return items, total, nil
}

func recordToState(targetType, targetID string, record *model.CommunitySyncRecord) *CommunitySyncState {
	if record == nil {
		return &CommunitySyncState{TargetType: targetType, TargetID: targetID, SyncStatus: "none"}
	}
	pushedAt := ""
	if record.PushedAt != nil {
		pushedAt = record.PushedAt.Format("2006-01-02 15:04")
	}
	return &CommunitySyncState{
		TargetType:         targetType,
		TargetID:           targetID,
		SyncStatus:         record.SyncStatus,
		CommunityArticleID: record.CommunityArticleID,
		LastError:          record.LastError,
		PushedAt:           pushedAt,
	}
}

// —— 配置 ——

// CommunityPushConfigData 配置数据（push_key 脱敏为前缀）。
type CommunityPushConfigData struct {
	BaseURL       string `json:"base_url"`
	SiteURL       string `json:"site_url"`
	PushKeyPrefix string `json:"push_key_prefix"` // nbk_xxxxxxxx，完整密钥不回显
	HasPushKey    bool   `json:"has_push_key"`
	VerifyCode    string `json:"verify_code"` // 社区归属验证码（对外可见，直接回显）
}

// GetPushConfig 读取推送配置（yaml 出厂值兜底，DB 单行表优先）。
func (l *CommunityPushLogic) GetPushConfig(ctx context.Context) (*CommunityPushConfigData, error) {
	row, err := l.configModel.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	baseURL := row.BaseURL
	if baseURL == "" {
		baseURL = OfficialSiteBaseURL()
	}
	siteURL := row.SiteURL
	if siteURL == "" {
		siteURL = getCommunityPushSettings().SiteURL
	}
	prefix := row.PushKey
	if len(prefix) > 12 {
		prefix = prefix[:12]
	}
	return &CommunityPushConfigData{
		BaseURL:       baseURL,
		SiteURL:       siteURL,
		PushKeyPrefix: prefix,
		HasPushKey:    strings.TrimSpace(row.PushKey) != "",
		VerifyCode:    row.VerifyCode,
	}, nil
}

// UpdatePushConfigRequest 更新推送配置请求。
type UpdatePushConfigRequest struct {
	BaseURL    string `json:"base_url"`
	SiteURL    string `json:"site_url"`
	PushKey    string `json:"push_key"`    // 留空表示不修改已有密钥
	VerifyCode string `json:"verify_code"` // 社区归属验证码（允许清空）
}

// UpdatePushConfigResult 保存推送配置结果（附验证码提交结果）。
type UpdatePushConfigResult struct {
	CommunityPushConfigData
	VerifyStatus  string `json:"verify_status"`  // verified | failed | skipped
	VerifyMessage string `json:"verify_message"` // failed 时的原因
}

// UpdatePushConfig 保存推送配置并热更新运行时值；携带验证码时顺带向社区提交归属验证。
func (l *CommunityPushLogic) UpdatePushConfig(ctx context.Context, r *UpdatePushConfigRequest) (*UpdatePushConfigResult, error) {
	row, err := l.configModel.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	normalized, err := novapush.NormalizeBaseURL(r.BaseURL)
	if err != nil {
		return nil, enum.ErrCommunityPushBaseURLInvalid
	}
	row.BaseURL = normalized
	row.SiteURL = strings.TrimRight(strings.TrimSpace(r.SiteURL), "/")
	row.VerifyCode = strings.TrimSpace(r.VerifyCode)
	if strings.TrimSpace(r.PushKey) != "" {
		row.PushKey = strings.TrimSpace(r.PushKey)
	}
	if err := l.configModel.UpdateConfig(ctx, row); err != nil {
		return nil, err
	}

	// 热更新运行时值
	SetOfficialSiteBaseURL(normalized)
	SetCommunitySiteURL(row.SiteURL)
	SetCommunityPushKey(row.PushKey)

	config, err := l.GetPushConfig(ctx)
	if err != nil {
		return nil, err
	}
	result := &UpdatePushConfigResult{CommunityPushConfigData: *config}
	result.VerifyStatus, result.VerifyMessage = l.submitVerifyCode(ctx, row.BaseURL, row.VerifyCode)
	return result, nil
}

// submitVerifyCode 向社区提交站点归属验证码。
// skipped：未填写验证码；failed：社区拒绝或网络异常（已保存验证码，修正后重新保存即可）。
func (l *CommunityPushLogic) submitVerifyCode(ctx context.Context, baseURL, code string) (status, message string) {
	if code == "" {
		return "skipped", ""
	}
	normalized, err := novapush.NormalizeBaseURL(baseURL)
	if err != nil {
		return "failed", "社区地址无效，无法提交验证码"
	}
	result, err := novapush.VerifyBlog(ctx, normalized, code)
	if err != nil {
		communityPushLog().Warn("提交站点归属验证码失败", zap.Error(err))
		return "failed", mapVerifyError(err)
	}
	return "verified", "验证成功，社区账号「" + result.Username + "」已获得站长标识"
}

// mapVerifyError 验证码提交失败的用户可读原因（无推送密钥参与，不走 mapPushError 的密钥文案）。
func mapVerifyError(err error) string {
	var apiErr *novapush.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Message
	}
	var authErr *novapush.AuthError
	if errors.As(err, &authErr) {
		return authErr.Message
	}
	return "社区服务暂不可用，请稍后在社区页面刷新验证状态"
}

// TestConnection 连通性自检：ping 社区并回显博主信息。
func (l *CommunityPushLogic) TestConnection(ctx context.Context) (*novapush.PingInfo, error) {
	baseURL, pushKey, err := l.resolvePushContext()
	if err != nil {
		return nil, err
	}
	info, err := novapush.Ping(ctx, baseURL, pushKey)
	if err != nil {
		return info, mapPushError(err)
	}
	return info, nil
}

// —— 内部工具 ——

// markFailure 记录推送失败（保留原 synced 状态语义切换为 failed，等待重试）。
func (l *CommunityPushLogic) markFailure(ctx context.Context, targetType, targetID, title string, pushErr error) {
	record, err := l.syncModel.GetByTarget(ctx, targetType, targetID)
	if err != nil {
		return
	}
	if record == nil {
		record = &model.CommunitySyncRecord{TargetType: targetType, TargetID: targetID}
	}
	record.SyncStatus = model.CommunitySyncStatusFailed
	record.LastError = truncateRunes(pushErr.Error(), 1000)
	record.RetryCount++
	if title != "" {
		record.ContentTitle = title
	}
	if err := l.syncModel.UpsertByTarget(ctx, record); err != nil {
		communityPushLog().Error("写入失败状态失败", zap.Error(err))
	}
}

func (l *CommunityPushLogic) markStatus(ctx context.Context, targetType, targetID, status string, payload *novapush.UpsertArticlePayload, communityID uint) error {
	hash, hashErr := payloadHash(payload)
	if hashErr != nil {
		hash = ""
	}
	return l.markStatusWithArgs(ctx, targetType, targetID, status, communityID, hash, payload.Title)
}

func (l *CommunityPushLogic) markStatusWithArgs(ctx context.Context, targetType, targetID, status string, communityID uint, hashAndTitle ...any) error {
	record, err := l.syncModel.GetByTarget(ctx, targetType, targetID)
	if err != nil {
		return err
	}
	if record == nil {
		record = &model.CommunitySyncRecord{TargetType: targetType, TargetID: targetID}
	}
	record.SyncStatus = status
	record.CommunityArticleID = communityID
	record.LastError = ""
	record.RetryCount = 0
	now := time.Now()
	record.PushedAt = &now
	if len(hashAndTitle) > 1 {
		if h, ok := hashAndTitle[0].(string); ok {
			record.ContentHash = h
		}
		if t, ok := hashAndTitle[1].(string); ok && t != "" {
			record.ContentTitle = t
		}
	}
	return l.syncModel.UpsertByTarget(ctx, record)
}

// payloadHash 对推送 payload 做摘要（内容与状态均一致时跳过推送）。
func payloadHash(payload *novapush.UpsertArticlePayload) (string, error) {
	bs, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(bs)
	return hex.EncodeToString(sum[:]), nil
}

// mapPushError 将客户端错误映射为业务错误。
func mapPushError(err error) error {
	var authErr *novapush.AuthError
	if errors.As(err, &authErr) {
		return enum.NewBizError(401106, "推送密钥无效或已吊销，请在系统设置中更新", 400)
	}
	var apiErr *novapush.APIError
	if errors.As(err, &apiErr) {
		return enum.NewBizError(400404, fmt.Sprintf("社区推送被拒绝：%s", apiErr.Message), 400)
	}
	return enum.ErrCommunityPushUpstream
}

// SyncOnUpdatedAsync 异步触发自动同步（钩子调用入口，避免阻塞内容保存）。
func SyncOnUpdatedAsync(targetType, targetID string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				communityPushLog().Error("社区自动同步 panic", zap.Any("recover", r))
			}
		}()
		logic := NewCommunityPushLogic()
		logic.SyncOnUpdated(context.Background(), targetType, targetID)
	}()
}

// SyncOnDeletedAsync 异步触发删除同步。
func SyncOnDeletedAsync(targetType, targetID string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				communityPushLog().Error("社区删除同步 panic", zap.Any("recover", r))
			}
		}()
		logic := NewCommunityPushLogic()
		logic.SyncOnDeleted(context.Background(), targetType, targetID)
	}()
}
