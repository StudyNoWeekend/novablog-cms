package req

// UpdateSecurityConfigReq 更新安全配置请求参数。
type UpdateSecurityConfigReq struct {
	SecurityEnabled     *bool `json:"security_enabled"`
	BlacklistTTLMinutes *int  `json:"blacklist_ttl_minutes" binding:"omitempty,min=1,max=1440"`
	LogRetentionDays    *int  `json:"log_retention_days" binding:"omitempty,min=1,max=365"`
}

// BlacklistQueryReq 黑名单列表查询请求参数。
type BlacklistQueryReq struct {
	PageReq
}

// BlacklistReq 创建/更新黑名单请求参数。
type BlacklistReq struct {
	IP     string `json:"ip" binding:"required,ip"`
	Reason string `json:"reason" binding:"max=255"`
}

// ListBlacklistReq 黑名单列表查询请求参数。
type ListBlacklistReq struct {
	PageReq
	Keyword string `form:"keyword" binding:"max=64"`
}

// ListAccessLogReq IP 访问统计查询请求参数。
type ListAccessLogReq struct {
	PageReq
	IP     string `form:"ip" binding:"max=64"`
	Region string `form:"region" binding:"max=128"`
}
