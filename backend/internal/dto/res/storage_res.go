package res

import "time"

// StorageConfigRes 存储配置响应（access_secret 脱敏）
type StorageConfigRes struct {
	ID           string    `json:"id"`
	Provider     string    `json:"provider"`
	Endpoint     string    `json:"endpoint"`
	Region       string    `json:"region"`
	Bucket       string    `json:"bucket"`
	AccessKey    string    `json:"access_key"`
	AccessSecret string    `json:"access_secret"` // 返回 "******"
	PathPrefix   string    `json:"path_prefix"`
	CustomDomain string    `json:"custom_domain"`
	Extra        string    `json:"extra"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// MigrationTaskRes 迁移任务状态响应
type MigrationTaskRes struct {
	ID             string     `json:"id"`
	TaskType       string     `json:"task_type"` // analyze/migrate
	TargetProvider string     `json:"target_provider"`
	Status         string     `json:"status"` // pending/running/completed/failed/canceled
	Total          int        `json:"total"`
	Succeeded      int        `json:"succeeded"`
	Failed         int        `json:"failed"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	Error          string     `json:"error"`
	CreatedAt      time.Time  `json:"created_at"`
}

// MigrationItemRes 迁移明细响应
type MigrationItemRes struct {
	ID      string `json:"id"`
	MediaID string `json:"media_id"`
	Status  string `json:"status"` // for analyze: exist/missing; for migrate: pending/success/failed
	Error   string `json:"error"`
}

// AnalyzeResultRes 分析结果响应
type AnalyzeResultRes struct {
	Task     MigrationTaskRes   `json:"task"`
	Missing  []MigrationItemRes `json:"missing"`
	Existing []MigrationItemRes `json:"existing"`
}
