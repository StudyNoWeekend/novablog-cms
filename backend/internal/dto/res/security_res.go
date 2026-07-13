package res

import "time"

// SecurityConfigRes 安全配置响应结构体。
type SecurityConfigRes struct {
	GetMaxTokens        int `json:"get_max_tokens"`
	GetWindowSeconds    int `json:"get_window_seconds"`
	PostMaxTokens       int `json:"post_max_tokens"`
	PostWindowSeconds   int `json:"post_window_seconds"`
	ViewMaxTokens       int `json:"view_max_tokens"`
	ViewWindowSeconds   int `json:"view_window_seconds"`
	LikeMaxTokens       int `json:"like_max_tokens"`
	LikeWindowSeconds   int `json:"like_window_seconds"`
	BlacklistThreshold  int `json:"blacklist_threshold"`
	BlacklistTTLMinutes int `json:"blacklist_ttl_minutes"`
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

// DailyCountRes 每日计数响应结构体。
type DailyCountRes struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// TopIPRes 违规 IP 排行项响应结构体。
type TopIPRes struct {
	IPAddress string `json:"ip_address"`
	Count     int64  `json:"count"`
}

// SecurityStatsRes 安全统计响应结构体。
type SecurityStatsRes struct {
	BlockedIPCount      int64           `json:"blocked_ip_count"`
	TodayRateLimitCount int64           `json:"today_rate_limit_count"`
	DailyTrend          []DailyCountRes `json:"daily_trend"`
	TopViolations       []TopIPRes      `json:"top_violations"`
}
