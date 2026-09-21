// Package logic 定义业务逻辑层。
package logic

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/internal/storage"
	"novablog/pkg/novablogapi"
	"novablog/utils/crypto"
	"novablog/utils/hash"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SetupLogger 初始化模块日志记录器。
var SetupLogger *zap.Logger

// SetupLogic 初始化业务逻辑结构体。
type SetupLogic struct {
	bloggerModel *model.BloggerModel
	manager      *storage.Manager
	cryptoKey    string
}

// NewSetupLogic 创建 SetupLogic 实例。
func NewSetupLogic(manager *storage.Manager, cryptoKey string) *SetupLogic {
	return &SetupLogic{
		bloggerModel: model.NewBlogger(),
		manager:      manager,
		cryptoKey:    cryptoKey,
	}
}

// CheckStatus 检查系统初始化状态。
func (l *SetupLogic) CheckStatus(ctx context.Context) (*res.StatusRes, error) {
	count, err := l.bloggerModel.Count(ctx)
	if err != nil {
		SetupLogger.Error("查询博主数量失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	return &res.StatusRes{
		Initialized: count > 0,
	}, nil
}

// InitBlogger 初始化博主账号。
func (l *SetupLogic) InitBlogger(ctx context.Context, r *req.InitReq) (*res.InitRes, error) {
	// 检查是否已初始化
	count, err := l.bloggerModel.Count(ctx)
	if err != nil {
		SetupLogger.Error("查询博主数量失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}
	if count > 0 {
		SetupLogger.Warn("系统已初始化，无法重复创建")
		return nil, enum.ErrAlreadyInitialized
	}

	// 哈希密码
	passwordHash, err := hash.HashPassword(r.Password)
	if err != nil {
		SetupLogger.Error("密码哈希失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	// 设置昵称，如果未提供则使用用户名
	nickname := r.Nickname
	if nickname == "" {
		nickname = r.Username
	}

	// 创建博主记录
	blogger := &model.Blogger{
		ID:           uuid.New().String(),
		Username:     r.Username,
		PasswordHash: passwordHash,
		Nickname:     nickname,
	}

	if err := l.bloggerModel.Create(ctx, blogger); err != nil {
		SetupLogger.Error("创建博主失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	SetupLogger.Info("博主账号初始化成功", zap.String("username", r.Username))

	return &res.InitRes{
		Success: true,
		Message: "初始化成功",
	}, nil
}

// SetupStorage 首装向导保存存储配置。
// provider 为空或 "local" 时创建本地存储配置记录（IsActive=true）。
// 非 local 时创建激活状态的对象存储记录并热重载 storage.Manager。
func (l *SetupLogic) SetupStorage(ctx context.Context, r *req.SetupStorageReq) (*res.InitRes, error) {
	// provider 为空或 local → 创建本地存储配置记录并激活
	if r.Provider == "" || r.Provider == "local" {
		config := &model.StorageConfig{
			Provider: "local",
			IsActive: true,
		}
		if err := model.NewStorageConfig().Create(ctx, config); err != nil {
			SetupLogger.Error("保存本地存储配置失败", zap.Error(err))
			return nil, enum.ErrInternalServer
		}

		// 热重载存储管理器
		if err := l.manager.Reload(ctx); err != nil {
			SetupLogger.Error("存储配置热重载失败", zap.Error(err))
		}

		SetupLogger.Info("首装向导本地存储配置完成")
		return &res.InitRes{
			Success: true,
			Message: "将使用本地存储",
		}, nil
	}

	// 校验云存储必填字段
	if r.Endpoint == "" {
		return nil, enum.NewBizError(enum.ErrInvalidParam.Code, "Endpoint 不能为空", enum.ErrInvalidParam.HttpCode)
	}
	if r.AccessKey == "" {
		return nil, enum.NewBizError(enum.ErrInvalidParam.Code, "AccessKey 不能为空", enum.ErrInvalidParam.HttpCode)
	}
	if r.AccessSecret == "" {
		return nil, enum.NewBizError(enum.ErrInvalidParam.Code, "AccessSecret 不能为空", enum.ErrInvalidParam.HttpCode)
	}

	// 加密 access_secret
	encrypted, err := crypto.Encrypt(r.AccessSecret, l.cryptoKey)
	if err != nil {
		SetupLogger.Error("加密存储密钥失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	// 创建激活状态的存储配置
	config := &model.StorageConfig{
		ID:           uuid.New().String(),
		Provider:     r.Provider,
		Endpoint:     r.Endpoint,
		Region:       r.Region,
		Bucket:       r.Bucket,
		AccessKey:    r.AccessKey,
		AccessSecret: encrypted,
		PathPrefix:   r.PathPrefix,
		CustomDomain: r.CustomDomain,
		Extra:        r.Extra,
		IsActive:     true,
	}

	if err := model.NewStorageConfig().Create(ctx, config); err != nil {
		SetupLogger.Error("保存存储配置失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	// 热重载存储管理器
	if err := l.manager.Reload(ctx); err != nil {
		SetupLogger.Error("存储配置热重载失败", zap.Error(err))
		// 配置已保存，重载失败仅打日志，不阻断流程
	}

	SetupLogger.Info("首装向导存储配置完成",
		zap.String("provider", r.Provider),
		zap.String("bucket", r.Bucket),
	)

	return &res.InitRes{
		Success: true,
		Message: "存储配置保存成功",
	}, nil
}

// ThemeInstallJob 首装主题安装任务状态（内存态，仅保留最近一次，进程重启即失效）。
type ThemeInstallJob struct {
	ID           string
	Status       string // running | success | failed
	Stage        string // fetching | installing | activating | done
	Message      string
	ThemeID      string
	ThemeName    string
	ThemeVersion string
	StartedAt    time.Time
	FinishedAt   time.Time
}

// 主题安装任务存储（单任务滚动覆盖）。
var (
	themeInstallMu  sync.Mutex
	themeInstallJob *ThemeInstallJob
)

// StartThemeInstall 启动首装主题安装任务：登录官方账号后拉取默认主题并激活；
// marketBaseURL 为首装向导输入的官方地址（覆盖 config.yaml 的 themes.market_base_url），
// marketEmail/marketPassword 为官方市场账号（官方代理下载要求登录态，密码仅瞬时使用不落库）。
// 重复触发时若任务进行中则返回当前状态，已有激活主题时幂等成功。
func (l *SetupLogic) StartThemeInstall(ctx context.Context, marketBaseURL, marketEmail, marketPassword string) (*res.ThemeInstallStatusRes, error) {
	// 必须先完成博主账号初始化
	count, err := l.bloggerModel.Count(ctx)
	if err != nil {
		return nil, enum.ErrInternalServer
	}
	if count == 0 {
		return nil, enum.NewBizError(enum.ErrInvalidParam.Code, "请先完成博主账号初始化", enum.ErrInvalidParam.HttpCode)
	}

	// 用向导传入的地址覆盖配置并持久化（如非空），重启后依然生效
	if marketBaseURL != "" {
		if err := NewThemeMarketConfigLogic().SyncMarketBaseURL(ctx, marketBaseURL); err != nil {
			return nil, err
		}
	}

	themeLogic := NewThemeLogic()
	if active, err := themeLogic.GetActiveTheme(ctx); err == nil && active != nil {
		return &res.ThemeInstallStatusRes{
			Status:       "success",
			Stage:        "done",
			ThemeID:      active.ThemeID,
			ThemeName:    active.Name,
			ThemeVersion: active.Version,
			Message:      "已有使用中的主题",
		}, nil
	}

	themeInstallMu.Lock()
	if themeInstallJob != nil && themeInstallJob.Status == "running" {
		job := themeInstallJob
		themeInstallMu.Unlock()
		return themeInstallStatusRes(job), nil
	}
	job := &ThemeInstallJob{
		ID:        uuid.New().String(),
		Status:    "running",
		Stage:     "fetching",
		StartedAt: time.Now(),
	}
	themeInstallJob = job
	themeInstallMu.Unlock()

	go l.runThemeInstall(job, marketBaseURL, marketEmail, marketPassword)
	return themeInstallStatusRes(job), nil
}

// GetThemeInstallStatus 查询最近一次主题安装任务状态。
func (l *SetupLogic) GetThemeInstallStatus(ctx context.Context) (*res.ThemeInstallStatusRes, error) {
	themeInstallMu.Lock()
	job := themeInstallJob
	themeInstallMu.Unlock()
	if job == nil {
		return &res.ThemeInstallStatusRes{Status: "not_started", Stage: ""}, nil
	}
	return themeInstallStatusRes(job), nil
}

// runThemeInstall 执行安装任务：登录官方账号后从官方市场拉取默认主题并激活。
func (l *SetupLogic) runThemeInstall(job *ThemeInstallJob, marketBaseURL, marketEmail, marketPassword string) {
	// 统一的任务状态更新入口（避免跨协程数据竞争）
	setJob := func(mutate func(*ThemeInstallJob)) {
		themeInstallMu.Lock()
		mutate(job)
		themeInstallMu.Unlock()
	}

	defer func() {
		if r := recover(); r != nil {
			setJob(func(j *ThemeInstallJob) {
				j.Status = "failed"
				j.Stage = "done"
				j.Message = fmt.Sprintf("安装任务异常退出: %v", r)
				j.FinishedAt = time.Now()
			})
			SetupLogger.Error("主题安装任务 panic", zap.Any("panic", r))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 向导传入的地址覆盖配置
	if marketBaseURL != "" {
		getThemeSettings().MarketBaseURL = marketBaseURL
	}

	themeLogic := NewThemeLogic()

	// 官方代理下载要求登录态：先用向导填写的官方账号换取 Token（密码仅瞬时使用，不落库）
	marketToken := ""
	if marketEmail != "" && marketPassword != "" {
		if base, berr := novablogapi.NormalizeBaseURL(getThemeSettings().MarketBaseURL); berr == nil {
			pair, _, loginErr := novablogapi.Login(ctx, base, marketEmail, marketPassword)
			if loginErr != nil {
				msg := loginErr.Error()
				var authErr *novablogapi.AuthError
				if errors.As(loginErr, &authErr) && authErr.Message != "" {
					// 透传官方登录错误文案（如"邮箱或密码错误"）
					msg = authErr.Message
				} else if bizErr, ok := mapUpstreamError(loginErr).(*enum.BizError); ok {
					msg = bizErr.Msg
				}
				setJob(func(j *ThemeInstallJob) {
					j.Status = "failed"
					j.Stage = "done"
					j.Message = "官方账号登录失败：" + msg
					j.FinishedAt = time.Now()
				})
				SetupLogger.Error("首装官方账号登录失败", zap.Error(loginErr))
				return
			}
			marketToken = pair.AccessToken
		}
	}

	setJob(func(j *ThemeInstallJob) { j.Stage = "installing" })
	theme, err := themeLogic.InstallDefault(ctx, marketToken)
	if err != nil {
		msg := err.Error()
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			msg = bizErr.Msg
		}
		setJob(func(j *ThemeInstallJob) {
			j.Status = "failed"
			j.Stage = "done"
			j.Message = msg
			j.FinishedAt = time.Now()
		})
		SetupLogger.Error("首装主题安装失败", zap.Error(err))
		return
	}

	setJob(func(j *ThemeInstallJob) { j.Stage = "activating" })
	if _, err := themeLogic.Activate(ctx, theme.ID); err != nil {
		msg := err.Error()
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			msg = bizErr.Msg
		}
		setJob(func(j *ThemeInstallJob) {
			j.Status = "failed"
			j.Stage = "done"
			j.Message = msg
			j.FinishedAt = time.Now()
		})
		SetupLogger.Error("首装主题激活失败", zap.Error(err))
		return
	}

	setJob(func(j *ThemeInstallJob) {
		j.Status = "success"
		j.Stage = "done"
		j.ThemeID = theme.ThemeID
		j.ThemeName = theme.Name
		j.ThemeVersion = theme.Version
		j.FinishedAt = time.Now()
	})
	SetupLogger.Info("首装主题安装完成",
		zap.String("theme_id", theme.ThemeID),
		zap.String("version", theme.Version),
	)
}

// themeInstallStatusRes 任务状态 → 响应 DTO。
func themeInstallStatusRes(job *ThemeInstallJob) *res.ThemeInstallStatusRes {
	started := job.StartedAt
	out := &res.ThemeInstallStatusRes{
		Status:       job.Status,
		Stage:        job.Stage,
		Message:      job.Message,
		ThemeID:      job.ThemeID,
		ThemeName:    job.ThemeName,
		ThemeVersion: job.ThemeVersion,
		StartedAt:    &started,
	}
	if !job.FinishedAt.IsZero() {
		finished := job.FinishedAt
		out.FinishedAt = &finished
	}
	return out
}
