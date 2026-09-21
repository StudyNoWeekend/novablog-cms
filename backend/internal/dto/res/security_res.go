package res

import "time"

// SecurityConfigRes 安全配置响应结构体。
type SecurityConfigRes struct {
	SecurityEnabled     bool `json:"security_enabled"`
	BlacklistTTLMinutes int  `json:"blacklist_ttl_minutes"`
	LogRetentionDays    int  `json:"log_retention_days"`
}

// BlacklistItemRes 黑名单列表项响应结构体。
type BlacklistItemRes struct {
	ID        string    `json:"id"`
	IPAddress string    `json:"ip_address"`
	Reason    string    `json:"reason"`
	BannedAt  time.Time `json:"banned_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool      `json:"is_active"`
}

// BlacklistRes 手动维护黑名单响应结构体。
type BlacklistRes struct {
	ID        uint      `json:"id"`
	IP        string    `json:"ip"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

// ListBlacklistRes 手动维护黑名单列表分页响应结构体。
type ListBlacklistRes struct {
	List       []*BlacklistRes `json:"list"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// IPAccessStatsRes IP 访问统计项响应结构体。
type IPAccessStatsRes struct {
	IP           string    `json:"ip"`
	Region       string    `json:"region"` // IP 归属地区，如 "中国 广东省 深圳市"
	TotalCount   int64     `json:"total_count"`
	ErrorCount   int64     `json:"error_count"`
	LastAccessAt time.Time `json:"last_access_at"`
}

// ListIPAccessStatsRes IP 访问统计列表分页响应结构体。
type ListIPAccessStatsRes struct {
	List       []*IPAccessStatsRes `json:"list"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}
