package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Game 游戏库模型，对应 games 数据表。
type Game struct {
	ID          string         `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Cover       string         `gorm:"type:varchar(500)" json:"cover"`
	Platform    string         `gorm:"type:varchar(100);default:''" json:"platform"`                          // PC/PS5/Switch/Xbox/移动端等
	Genre       string         `gorm:"type:varchar(100);default:''" json:"genre"`                             // 游戏类型：RPG/FPS/独立游戏等
	PlayStatus  string         `gorm:"column:play_status;type:varchar(20);default:'want'" json:"play_status"` // want=想玩, playing=在玩, played=玩过
	PlayHours   int            `gorm:"column:play_hours;type:int;default:0" json:"play_hours"`                // 累计游玩时长（小时）
	Rating      int            `gorm:"type:int;default:0" json:"rating"`                                      // 0-10 分，0=未评分
	ShortReview string         `gorm:"type:varchar(1000)" json:"short_review"`                                // 短评
	Status      int            `gorm:"type:int;default:0" json:"status"`                                      // 0=草稿, 1=已发布
	SortOrder   int            `gorm:"column:sort_order;type:int;default:0" json:"sort_order"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定数据表名称。
func (Game) TableName() string {
	return "games"
}

// AfterFind GORM 查询后钩子，将相对路径 URL 拼接为完整 URL。
func (g *Game) AfterFind(tx *gorm.DB) error {
	g.Cover = resolveURL(g.Cover)
	return nil
}

// GameModel 游戏库模型操作结构体。
type GameModel struct {
	db *gorm.DB
}

// NewGame 创建 GameModel 实例。
func NewGame() *GameModel {
	return &GameModel{db: DB}
}

// Create 创建游戏记录。
func (m *GameModel) Create(ctx context.Context, g *Game) error {
	return m.db.WithContext(ctx).Create(g).Error
}

// GetByID 根据 ID 查询游戏。
func (m *GameModel) GetByID(ctx context.Context, id string) (*Game, error) {
	var g Game
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&g).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// GetList 分页查询游戏列表，支持 keyword 模糊搜索标题、platform/genre/playStatus 筛选、status 过滤。
func (m *GameModel) GetList(ctx context.Context, keyword *string, platform *string, genre *string, playStatus *string, status *int, page, pageSize int) ([]Game, int64, error) {
	var total int64
	query := m.db.WithContext(ctx).Model(&Game{})

	if keyword != nil && *keyword != "" {
		query = query.Where("title ILIKE ?", "%"+*keyword+"%")
	}
	if platform != nil && *platform != "" {
		query = query.Where("platform ILIKE ?", "%"+*platform+"%")
	}
	if genre != nil && *genre != "" {
		query = query.Where("genre ILIKE ?", "%"+*genre+"%")
	}
	if playStatus != nil && *playStatus != "" && *playStatus != "all" {
		query = query.Where("play_status = ?", *playStatus)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []Game
	offset := (page - 1) * pageSize
	err := query.
		Order("sort_order ASC, updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Update 更新游戏。
func (m *GameModel) Update(ctx context.Context, g *Game) error {
	return m.db.WithContext(ctx).Save(g).Error
}

// SoftDelete 软删除游戏。
func (m *GameModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&Game{}).Error
}
