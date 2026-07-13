package res

import "time"

// TravelGuideRes 旅行攻略列表项响应结构体。
type TravelGuideRes struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Summary      string    `json:"summary"`
	CoverImage   string    `json:"cover_image"`
	Status       int16     `json:"status"`
	Destination  string    `json:"destination"`
	Region       string    `json:"region"`
	CategoryID   string    `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Days         int       `json:"days"`
	BestMonth    string    `json:"best_month"`
	ViewCount    int       `json:"view_count"`
	LikeCount    int       `json:"like_count"`
	Rating       float64   `json:"rating"`
	ReviewCount  int       `json:"review_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TravelGuideDetailRes 旅行攻略详情响应结构体。
type TravelGuideDetailRes struct {
	TravelGuideRes
	Attractions []map[string]any `json:"attractions"`
	Itinerary   []map[string]any `json:"itinerary"`
	Reviews     []map[string]any `json:"reviews"`
}
