package res

// MarketUserRes 官方账号用户信息（脱敏，仅后台展示所需字段）。
type MarketUserRes struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
}

// MarketLoginRes 登录官方主题市场响应。
type MarketLoginRes struct {
	AccessToken string        `json:"access_token"`
	ExpiresIn   int64         `json:"expires_in"`
	User        MarketUserRes `json:"user"`
}

// ThemeMarketItemRes 官方主题市场条目（字段与官方 ThemeItem 一致）。
type ThemeMarketItemRes struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	AuthorID    string   `json:"author_id"`
	Author      string   `json:"author"`
	Price       string   `json:"price"`
	PriceAmount float64  `json:"price_amount"`
	Type        string   `json:"type"`
	Styles      []string `json:"styles"`
	Features    []string `json:"features"`
	Preview     string   `json:"preview"`
	Version     string   `json:"version"`
	Downloads   int64    `json:"downloads"`
	Likes       int64    `json:"likes"`
	Rating      float64  `json:"rating"`
	Status      int      `json:"status"`
	CreatedAt   string   `json:"created_at"`
}

// ThemeMarketDetailRes 主题详情（含下载地址与登录态个人互动状态）。
type ThemeMarketDetailRes struct {
	ThemeMarketItemRes
	DownloadURL string  `json:"download_url"`
	Liked       bool    `json:"liked"`
	UserRating  float64 `json:"user_rating"`
}

// ThemeMarketReleaseRes 主题版本历史条目（自 GitHub Release 同步的更新日志）。
type ThemeMarketReleaseRes struct {
	Version     string `json:"version"`
	Tag         string `json:"tag"`
	Notes       string `json:"notes"`
	PublishedAt string `json:"published_at"`
}

// ThemeMarketStatsRes 官方主题市场统计。
type ThemeMarketStatsRes struct {
	Total     int64 `json:"total"`
	Authors   int64 `json:"authors"`
	Downloads int64 `json:"downloads"`
}

// ThemeMarketHotTagRes 热门风格标签。
type ThemeMarketHotTagRes struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// ThemeMarketLikeRes 点赞/取消点赞结果。
type ThemeMarketLikeRes struct {
	Liked bool `json:"liked"`
}

// ThemeMarketFavoriteRes 收藏/取消收藏结果。
type ThemeMarketFavoriteRes struct {
	Favorited bool `json:"favorited"`
}

// ThemeMarketRatingRes 评分结果（重算后的主题平均分）。
type ThemeMarketRatingRes struct {
	Rating float64 `json:"rating"`
}

// ThemeMarketDownloadRes 下载/安装结果。
type ThemeMarketDownloadRes struct {
	DownloadURL string `json:"download_url"`
}

// ThemeMarketDefaultRes 官方默认主题（部署首装直接拉取）。
type ThemeMarketDefaultRes struct {
	ThemeMarketItemRes
	IsDefault   bool   `json:"is_default"`
	DownloadURL string `json:"download_url"`
}
