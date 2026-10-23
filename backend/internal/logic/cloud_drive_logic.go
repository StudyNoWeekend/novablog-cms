package logic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"novablog/enum"
	"novablog/internal/cache"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/pkg/aliyundrive"
	"novablog/pkg/baidupan"
	"novablog/pkg/clouddrive"
	"novablog/pkg/clouddrive/factory"
	"novablog/pkg/webdav"
	"novablog/utils/crypto"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CloudDriveSettings 云盘导入模块配置，由 main 启动阶段注入。
type CloudDriveSettings struct {
	ProxyURL string // 可选 HTTP 代理（云盘 API 与下载；空=直连）
	// 单文件导入大小上限（MB），0=默认 2048；单次导入文件数上限，0=默认 200
	MaxFileSizeMB    int
	ImportBatchLimit int
}

// cloudDriveSettings 云盘导入模块运行时配置。
var cloudDriveSettings = &CloudDriveSettings{MaxFileSizeMB: 2048, ImportBatchLimit: 200}

// SetCloudDriveSettings 注入云盘导入模块配置。
func SetCloudDriveSettings(s *CloudDriveSettings) {
	if s != nil {
		cloudDriveSettings = s
	}
}

// MaxCloudImportFileSize 返回单文件导入大小上限（字节）。
func MaxCloudImportFileSize() int64 {
	mb := cloudDriveSettings.MaxFileSizeMB
	if mb <= 0 {
		mb = 2048
	}
	return int64(mb) << 20
}

// CloudImportBatchLimit 返回单次导入文件数上限。
func CloudImportBatchLimit() int {
	if cloudDriveSettings.ImportBatchLimit <= 0 {
		return 200
	}
	return cloudDriveSettings.ImportBatchLimit
}

// CloudDriveLogic 云盘账号与文件浏览业务逻辑。
type CloudDriveLogic struct {
	configModel  *model.CloudDriveConfigModel
	accountModel *model.CloudDriveAccountModel
	cryptoKey    string
	logger       *zap.Logger

	refreshMu   sync.Mutex
	refreshing  map[string]bool // 正在刷新 token 的账号 ID，防止并发刷新
}

// NewCloudDriveLogic 创建 CloudDriveLogic 实例。
func NewCloudDriveLogic(cryptoKey string, logger *zap.Logger) *CloudDriveLogic {
	return &CloudDriveLogic{
		configModel:  model.NewCloudDriveConfig(),
		accountModel: model.NewCloudDriveAccount(),
		cryptoKey:    cryptoKey,
		logger:       logger,
		refreshing:   make(map[string]bool),
	}
}

// ==================== 应用凭据配置 ====================

// ListConfigs 查询云盘应用凭据（client_secret 脱敏）。
func (l *CloudDriveLogic) ListConfigs(ctx context.Context) ([]res.CloudDriveConfigRes, error) {
	configs, err := l.configModel.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询云盘配置失败: %w", err)
	}
	result := make([]res.CloudDriveConfigRes, 0, len(configs))
	for _, c := range configs {
		result = append(result, res.CloudDriveConfigRes{
			ID:              c.ID,
			Provider:        c.Provider,
			ClientID:        c.ClientID,
			ClientSecret:    maskSecret(c.ClientSecret),
			CallbackBaseURL: c.CallbackBaseURL,
			CreatedAt:       c.CreatedAt,
			UpdatedAt:       c.UpdatedAt,
		})
	}
	return result, nil
}

// UpsertConfig 保存云盘应用凭据；client_secret 留空表示不修改。
func (l *CloudDriveLogic) UpsertConfig(ctx context.Context, r *req.CloudDriveConfigReq) error {
	if !clouddrive.Supported(r.Provider) || r.Provider == clouddrive.ProviderWebDAV {
		return enum.ErrCloudDriveProviderInvalid
	}
	secret := ""
	if r.ClientSecret != "" {
		encrypted, err := crypto.Encrypt(r.ClientSecret, l.cryptoKey)
		if err != nil {
			return fmt.Errorf("加密 client_secret 失败: %w", err)
		}
		secret = encrypted
	}
	config := &model.CloudDriveConfig{
		Provider:        r.Provider,
		ClientID:        r.ClientID,
		ClientSecret:    secret,
		CallbackBaseURL: r.CallbackBaseURL,
	}
	if err := l.configModel.Upsert(ctx, config); err != nil {
		return fmt.Errorf("保存云盘配置失败: %w", err)
	}
	return nil
}

// DeleteConfig 删除云盘应用凭据配置。
func (l *CloudDriveLogic) DeleteConfig(ctx context.Context, provider string) error {
	if err := l.configModel.Delete(ctx, provider); err != nil {
		return fmt.Errorf("删除云盘配置失败: %w", err)
	}
	return nil
}

// ==================== 账号管理 ====================

// ListAccounts 查询全部云盘账号（token/密码脱敏）。
func (l *CloudDriveLogic) ListAccounts(ctx context.Context) ([]res.CloudDriveAccountRes, error) {
	accounts, err := l.accountModel.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询云盘账号失败: %w", err)
	}
	result := make([]res.CloudDriveAccountRes, 0, len(accounts))
	for i := range accounts {
		result = append(result, l.toAccountRes(&accounts[i]))
	}
	return result, nil
}

// toAccountRes 账号模型转 DTO（脱敏）。
func (l *CloudDriveLogic) toAccountRes(a *model.CloudDriveAccount) res.CloudDriveAccountRes {
	return res.CloudDriveAccountRes{
		ID:          a.ID,
		Provider:    a.Provider,
		OpenUserID:  a.OpenUserID,
		DisplayName: a.DisplayName,
		BaseURL:     a.BaseURL,
		Username:    a.Username,
		Status:      a.Status,
		ExpiresAt:   a.ExpiresAt,
		QuotaTotal:  a.QuotaTotal,
		QuotaUsed:   a.QuotaUsed,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

// DeleteAccount 解绑云盘账号（软删除）。
func (l *CloudDriveLogic) DeleteAccount(ctx context.Context, id string) error {
	if _, err := l.accountModel.GetByID(ctx, id); err != nil {
		return enum.ErrCloudDriveAccountNotFound
	}
	if err := l.accountModel.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("解绑云盘账号失败: %w", err)
	}
	return nil
}

// TestWebDAV 测试 WebDAV 连通性（不落库）。
func (l *CloudDriveLogic) TestWebDAV(ctx context.Context, r *req.WebDAVTestReq) error {
	client := webdav.New(r.BaseURL, r.Username, r.Password, cloudDriveSettings.ProxyURL)
	return client.Test(ctx)
}

// AddWebDAVAccount 测试并保存 WebDAV 账号（密码 AES 加密落库）。
// 同一服务器+用户名重复绑定时更新密码。
func (l *CloudDriveLogic) AddWebDAVAccount(ctx context.Context, r *req.WebDAVAccountReq) (*res.CloudDriveAccountRes, error) {
	client := webdav.New(r.BaseURL, r.Username, r.Password, cloudDriveSettings.ProxyURL)
	if err := client.Test(ctx); err != nil {
		return nil, err
	}

	encrypted, err := crypto.Encrypt(r.Password, l.cryptoKey)
	if err != nil {
		return nil, fmt.Errorf("加密 WebDAV 密码失败: %w", err)
	}

	openUserID := webDAVOpenUserID(r.BaseURL, r.Username)
	account, err := l.accountModel.GetByProviderAndOpenUser(ctx, clouddrive.ProviderWebDAV, openUserID)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("查询 WebDAV 账号失败: %w", err)
		}
		account = &model.CloudDriveAccount{
			ID:          "",
			Provider:    clouddrive.ProviderWebDAV,
			OpenUserID:  openUserID,
			DisplayName: r.Username,
		}
	}
	account.AccessToken = encrypted
	account.BaseURL = r.BaseURL
	account.Username = r.Username
	account.Status = model.CloudAccountStatusActive

	if account.CreatedAt.IsZero() && account.ID == "" {
		if err := l.accountModel.Create(ctx, account); err != nil {
			return nil, fmt.Errorf("保存 WebDAV 账号失败: %w", err)
		}
	} else {
		account.DeletedAt = gorm.DeletedAt{}
		if err := l.accountModel.Update(ctx, account); err != nil {
			return nil, fmt.Errorf("更新 WebDAV 账号失败: %w", err)
		}
	}
	result := l.toAccountRes(account)
	return &result, nil
}

// webDAVOpenUserID WebDAV 账号的唯一标识。
func webDAVOpenUserID(baseURL, username string) string {
	return baseURL + "|" + username
}

// maskSecret 脱敏已加密凭据：非空返回 "******"。
func maskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	return "******"
}

// ==================== OAuth 授权 ====================

// OAuthCallbackResult 授权回调处理结果（控制器据此渲染提示页）。
type OAuthCallbackResult struct {
	Success bool
	Message string
	Account string
}

// oauthStateKey Redis 中 OAuth state 的 key 前缀。
const oauthStateKey = "clouddrive:state:"

// oauthStateTTL state 有效期。
const oauthStateTTL = 5 * time.Minute

// GetAuthorizeURL 生成 OAuth 授权页地址。
func (l *CloudDriveLogic) GetAuthorizeURL(ctx context.Context, provider string) (*res.CloudAuthorizeRes, error) {
	if provider != clouddrive.ProviderBaidu && provider != clouddrive.ProviderAliyun {
		return nil, enum.ErrCloudDriveProviderInvalid
	}
	config, err := l.configModel.GetByProvider(ctx, provider)
	if err != nil {
		return nil, enum.ErrCloudDriveConfigMissing
	}
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, enum.ErrCloudDriveConfigMissing
	}
	if config.CallbackBaseURL == "" {
		return nil, fmt.Errorf("请先在应用凭据中配置回调地址（本实例外部可达的域名）")
	}

	state, err := randomState()
	if err != nil {
		return nil, fmt.Errorf("生成授权状态失败: %w", err)
	}
	if cache.RedisClient == nil {
		return nil, fmt.Errorf("Redis 不可用，无法发起授权")
	}
	if err := cache.RedisClient.Set(ctx, oauthStateKey+state, provider, oauthStateTTL).Err(); err != nil {
		return nil, fmt.Errorf("存储授权状态失败: %w", err)
	}

	redirectURI := callbackURI(config.CallbackBaseURL, provider)
	var authURL string
	switch provider {
	case clouddrive.ProviderBaidu:
		authURL = baidupan.BuildAuthorizeURL(config.ClientID, redirectURI, state)
	case clouddrive.ProviderAliyun:
		authURL = aliyundrive.BuildAuthorizeURL(config.ClientID, redirectURI, state)
	}
	return &res.CloudAuthorizeRes{AuthURL: authURL, State: state}, nil
}

// callbackURI 拼接 OAuth 回调地址。
func callbackURI(callbackBaseURL, provider string) string {
	return fmt.Sprintf("%s/api/v1/clouddrive/callback/%s", trimSlash(callbackBaseURL), provider)
}

func trimSlash(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '/') {
		s = s[:len(s)-1]
	}
	return s
}

// randomState 生成随机 state。
func randomState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// HandleCallback 处理 OAuth 回调：校验 state → 换取凭据 → 拉取用户信息 → 落库。
func (l *CloudDriveLogic) HandleCallback(ctx context.Context, provider, code, state string) (*OAuthCallbackResult, error) {
	if provider != clouddrive.ProviderBaidu && provider != clouddrive.ProviderAliyun {
		return nil, enum.ErrCloudDriveProviderInvalid
	}
	// 校验并消费 state（一次性）
	if cache.RedisClient == nil {
		return nil, fmt.Errorf("Redis 不可用")
	}
	if state == "" {
		return nil, enum.ErrCloudDriveStateInvalid
	}
	saved, err := cache.RedisClient.Get(ctx, oauthStateKey+state).Result()
	if err != nil || saved != provider {
		return nil, enum.ErrCloudDriveStateInvalid
	}
	cache.RedisClient.Del(ctx, oauthStateKey+state)

	config, err := l.configModel.GetByProvider(ctx, provider)
	if err != nil {
		return nil, enum.ErrCloudDriveConfigMissing
	}
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, enum.ErrCloudDriveConfigMissing
	}
	clientSecret, err := crypto.Decrypt(config.ClientSecret, l.cryptoKey)
	if err != nil {
		return nil, fmt.Errorf("解密 client_secret 失败: %w", err)
	}
	redirectURI := callbackURI(config.CallbackBaseURL, provider)

	// 换取凭据
	proxy := cloudDriveSettings.ProxyURL
	var token *clouddrive.TokenResult
	switch provider {
	case clouddrive.ProviderBaidu:
		token, err = baidupan.ExchangeToken(ctx, config.ClientID, clientSecret, redirectURI, code, proxy)
	case clouddrive.ProviderAliyun:
		token, err = aliyundrive.ExchangeToken(ctx, config.ClientID, clientSecret, redirectURI, code, proxy)
	}
	if err != nil {
		l.logger.Warn("云盘授权换取凭据失败", zap.String("provider", provider), zap.Error(err))
		return &OAuthCallbackResult{Success: false, Message: err.Error()}, nil
	}

	// 拉取用户信息并落库
	openUserID, displayName, quotaTotal, quotaUsed, driveID, uerr := l.fetchUserInfo(ctx, provider, token.AccessToken)
	if uerr != nil {
		l.logger.Warn("云盘授权拉取用户信息失败", zap.String("provider", provider), zap.Error(uerr))
		return &OAuthCallbackResult{Success: false, Message: uerr.Error()}, nil
	}

	if err := l.upsertOAuthAccount(ctx, provider, openUserID, displayName, token, quotaTotal, quotaUsed, driveID); err != nil {
		return &OAuthCallbackResult{Success: false, Message: err.Error()}, nil
	}
	return &OAuthCallbackResult{Success: true, Message: "授权成功", Account: displayName}, nil
}

// fetchUserInfo 用 access token 拉取用户信息（含阿里云盘 drive_id）。
func (l *CloudDriveLogic) fetchUserInfo(ctx context.Context, provider, accessToken string) (openUserID, displayName string, quotaTotal, quotaUsed int64, driveID string, err error) {
	switch provider {
	case clouddrive.ProviderBaidu:
		info, uerr := baidupan.New(accessToken, cloudDriveSettings.ProxyURL).UserInfo(ctx)
		if uerr != nil {
			return "", "", 0, 0, "", uerr
		}
		return info.OpenUserID, info.DisplayName, info.QuotaTotal, info.QuotaUsed, "", nil
	case clouddrive.ProviderAliyun:
		info, uerr := aliyundrive.New(accessToken, "", cloudDriveSettings.ProxyURL).UserInfo(ctx)
		if uerr != nil {
			return "", "", 0, 0, "", uerr
		}
		return info.OpenUserID, info.DisplayName, info.QuotaTotal, info.QuotaUsed, info.DriveID, nil
	}
	return "", "", 0, 0, "", enum.ErrCloudDriveProviderInvalid
}

// upsertOAuthAccount 创建或更新 OAuth 账号（token AES 加密落库）。
func (l *CloudDriveLogic) upsertOAuthAccount(ctx context.Context, provider, openUserID, displayName string, token *clouddrive.TokenResult, quotaTotal, quotaUsed int64, driveID string) error {
	encAccess, err := crypto.Encrypt(token.AccessToken, l.cryptoKey)
	if err != nil {
		return fmt.Errorf("加密 access_token 失败: %w", err)
	}
	encRefresh, err := crypto.Encrypt(token.RefreshToken, l.cryptoKey)
	if err != nil {
		return fmt.Errorf("加密 refresh_token 失败: %w", err)
	}
	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	if token.ExpiresIn <= 0 {
		expiresAt = time.Now().Add(2 * time.Hour)
	}

	account, err := l.accountModel.GetByProviderAndOpenUser(ctx, provider, openUserID)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("查询云盘账号失败: %w", err)
		}
		account = &model.CloudDriveAccount{
			Provider:    provider,
			OpenUserID:  openUserID,
			DisplayName: displayName,
		}
	}
	account.AccessToken = encAccess
	account.RefreshToken = encRefresh
	account.ExpiresAt = &expiresAt
	account.DisplayName = displayName
	account.Status = model.CloudAccountStatusActive
	if driveID != "" {
		account.DriveID = driveID
	}
	if quotaTotal > 0 {
		account.QuotaTotal = quotaTotal
	}
	if quotaUsed > 0 {
		account.QuotaUsed = quotaUsed
	}

	if account.ID == "" {
		if err := l.accountModel.Create(ctx, account); err != nil {
			return fmt.Errorf("保存云盘账号失败: %w", err)
		}
	} else {
		account.DeletedAt = gorm.DeletedAt{}
		if err := l.accountModel.Update(ctx, account); err != nil {
			return fmt.Errorf("更新云盘账号失败: %w", err)
		}
	}
	return nil
}

// ==================== Token 管理与客户端构建 ====================

// GetValidAccessToken 获取有效的 access token：即将过期时自动刷新（防并发刷新）。
func (l *CloudDriveLogic) GetValidAccessToken(ctx context.Context, account *model.CloudDriveAccount) (string, error) {
	if account.Provider == clouddrive.ProviderWebDAV {
		// WebDAV：AccessToken 字段存的是加密密码
		return crypto.Decrypt(account.AccessToken, l.cryptoKey)
	}
	if account.AccessToken == "" {
		return "", enum.ErrCloudDriveStateInvalid
	}
	if !l.needRefresh(account) {
		return crypto.Decrypt(account.AccessToken, l.cryptoKey)
	}
	return l.refreshAccessToken(ctx, account)
}

// needRefresh 判断 token 是否需要刷新：无过期时间、已过期或剩余不足 10 分钟。
func (l *CloudDriveLogic) needRefresh(account *model.CloudDriveAccount) bool {
	if account.ExpiresAt == nil {
		return true
	}
	return time.Until(*account.ExpiresAt) < 10*time.Minute
}

// refreshAccessToken 刷新 access token（mutex 防并发，刷新后回写账号）。
func (l *CloudDriveLogic) refreshAccessToken(ctx context.Context, account *model.CloudDriveAccount) (string, error) {
	l.refreshMu.Lock()
	if l.refreshing[account.ID] {
		l.refreshMu.Unlock()
		// 等待他人刷新完成后再重试一次
		for i := 0; i < 20; i++ {
			time.Sleep(100 * time.Millisecond)
			if !l.isRefreshing(account.ID) {
				break
			}
		}
		if !l.needRefresh(account) {
			return crypto.Decrypt(account.AccessToken, l.cryptoKey)
		}
		return "", fmt.Errorf("token 刷新中，请重试")
	}
	l.refreshing[account.ID] = true
	l.refreshMu.Unlock()
	defer func() {
		l.refreshMu.Lock()
		delete(l.refreshing, account.ID)
		l.refreshMu.Unlock()
	}()

	if account.RefreshToken == "" {
		return "", enum.ErrCloudDriveStateInvalid
	}
	refreshToken, err := crypto.Decrypt(account.RefreshToken, l.cryptoKey)
	if err != nil {
		return "", fmt.Errorf("解密 refresh_token 失败: %w", err)
	}

	config, err := l.configModel.GetByProvider(ctx, account.Provider)
	if err != nil {
		return "", enum.ErrCloudDriveConfigMissing
	}
	clientSecret, err := crypto.Decrypt(config.ClientSecret, l.cryptoKey)
	if err != nil {
		return "", fmt.Errorf("解密 client_secret 失败: %w", err)
	}

	proxy := cloudDriveSettings.ProxyURL
	var token *clouddrive.TokenResult
	switch account.Provider {
	case clouddrive.ProviderBaidu:
		token, err = baidupan.RefreshToken(ctx, config.ClientID, clientSecret, refreshToken, proxy)
	case clouddrive.ProviderAliyun:
		token, err = aliyundrive.RefreshToken(ctx, config.ClientID, clientSecret, refreshToken, proxy)
	default:
		return "", enum.ErrCloudDriveProviderInvalid
	}
	if err != nil {
		// 刷新失败：标记账号过期，提示重新授权
		account.Status = model.CloudAccountStatusExpired
		if uerr := l.accountModel.Update(ctx, account); uerr != nil {
			l.logger.Warn("更新账号状态失败", zap.Error(uerr))
		}
		return "", fmt.Errorf("刷新云盘凭据失败: %w（请到云盘账号页重新授权）", err)
	}

	encAccess, encErr := crypto.Encrypt(token.AccessToken, l.cryptoKey)
	encRefresh, encErr2 := crypto.Encrypt(token.RefreshToken, l.cryptoKey)
	if encErr != nil || encErr2 != nil {
		return "", fmt.Errorf("加密刷新凭据失败")
	}
	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	account.AccessToken = encAccess
	account.RefreshToken = encRefresh
	account.ExpiresAt = &expiresAt
	account.Status = model.CloudAccountStatusActive
	if err := l.accountModel.Update(ctx, account); err != nil {
		return "", fmt.Errorf("回写刷新凭据失败: %w", err)
	}
	return token.AccessToken, nil
}

// isRefreshing 查询账号是否正在刷新。
func (l *CloudDriveLogic) isRefreshing(accountID string) bool {
	l.refreshMu.Lock()
	defer l.refreshMu.Unlock()
	return l.refreshing[accountID]
}

// BuildClientForAccount 为账号构建云盘客户端（自动刷新 token），供浏览与导入任务复用。
// WebDAV 分支：GetValidAccessToken 返回解密后的密码（AccessToken 字段复用存密码）。
func (l *CloudDriveLogic) BuildClientForAccount(ctx context.Context, account *model.CloudDriveAccount) (clouddrive.Client, error) {
	var params factory.BuildParams
	switch account.Provider {
	case clouddrive.ProviderWebDAV:
		password, err := l.GetValidAccessToken(ctx, account)
		if err != nil {
			return nil, err
		}
		params = factory.BuildParams{
			Provider: account.Provider,
			BaseURL:  account.BaseURL,
			Username: account.Username,
			Password: password,
		}
	default:
		token, err := l.GetValidAccessToken(ctx, account)
		if err != nil {
			return nil, err
		}
		params = factory.BuildParams{
			Provider:    account.Provider,
			AccessToken: token,
			DriveID:     account.DriveID,
		}
	}
	params.ProxyURL = cloudDriveSettings.ProxyURL
	return factory.BuildClient(params)
}

// ListFiles 浏览云盘目录。
func (l *CloudDriveLogic) ListFiles(ctx context.Context, accountID, parent, cursor string, limit int) (*res.CloudFileListRes, error) {
	account, err := l.accountModel.GetByID(ctx, accountID)
	if err != nil {
		return nil, enum.ErrCloudDriveAccountNotFound
	}
	client, err := l.BuildClientForAccount(ctx, account)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	files, next, err := client.List(ctx, parent, cursor, limit)
	if err != nil {
		return nil, err
	}
	items := make([]res.CloudFileRes, 0, len(files))
	for _, f := range files {
		items = append(items, res.CloudFileRes{
			FileID:   f.FileID,
			ParentID: f.ParentID,
			Path:     f.Path,
			Name:     f.Name,
			IsDir:    f.IsDir,
			Size:     f.Size,
			Ext:      f.Ext,
		})
	}
	return &res.CloudFileListRes{
		List:       items,
		NextCursor: next,
		Parent:     parent,
	}, nil
}
