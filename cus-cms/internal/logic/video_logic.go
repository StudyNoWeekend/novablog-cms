package logic

import (
	"context"
	"fmt"

	"cus-cms/internal/dto/req"
	"cus-cms/internal/dto/res"
	"cus-cms/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VideoLogic 视频作品业务逻辑结构体。
type VideoLogic struct {
	videoModel        *model.VideoWorkModel
	platformLinkModel *model.VideoPlatformLinkModel
}

// NewVideoLogic 创建 VideoLogic 实例。
func NewVideoLogic() *VideoLogic {
	return &VideoLogic{
		videoModel:        model.NewVideoWork(),
		platformLinkModel: model.NewVideoPlatformLink(),
	}
}

// Create 创建视频作品，事务内创建视频及平台链接。
func (l *VideoLogic) Create(ctx context.Context, r *req.CreateVideoReq) (*res.VideoWorkRes, error) {
	video := &model.VideoWork{
		ID:          uuid.New().String(),
		Title:       r.Title,
		CoverURL:    r.CoverURL,
		Description: r.Description,
	}
	if r.Status != nil {
		video.Status = *r.Status
	}

	links := make([]model.VideoPlatformLink, 0, len(r.Platforms))
	for _, p := range r.Platforms {
		links = append(links, model.VideoPlatformLink{
			ID:       uuid.New().String(),
			VideoID:  video.ID,
			Platform: p.Platform,
			URL:      p.URL,
		})
	}

	if err := l.videoModel.Transaction(ctx, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(video).Error; err != nil {
			return fmt.Errorf("创建视频作品失败: %w", err)
		}
		if err := l.platformLinkModel.BatchCreate(ctx, tx, links); err != nil {
			return fmt.Errorf("创建平台链接失败: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return l.toVideoWorkRes(video, links), nil
}

// GetList 分页查询视频作品列表，含每个视频的平台链接。
func (l *VideoLogic) GetList(ctx context.Context, r *req.VideoListReq) (*res.VideoWorkListRes, error) {
	list, total, err := l.videoModel.GetList(ctx, r.Keyword, r.Status, r.GetPage(), r.GetPageSize())
	if err != nil {
		return nil, fmt.Errorf("查询视频作品列表失败: %w", err)
	}

	ids := make([]string, 0, len(list))
	for _, v := range list {
		ids = append(ids, v.ID)
	}

	allLinks, err := l.platformLinkModel.GetByVideoIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("查询平台链接失败: %w", err)
	}

	linkMap := make(map[string][]model.VideoPlatformLink)
	for _, link := range allLinks {
		linkMap[link.VideoID] = append(linkMap[link.VideoID], link)
	}

	items := make([]res.VideoWorkRes, 0, len(list))
	for i := range list {
		items = append(items, *l.toVideoWorkRes(&list[i], linkMap[list[i].ID]))
	}

	return &res.VideoWorkListRes{
		List:       items,
		Total:      total,
		Page:       r.GetPage(),
		PageSize:   r.GetPageSize(),
		TotalPages: calcTotalPages(total, r.GetPageSize()),
	}, nil
}

// GetByID 查询视频作品详情（含平台链接列表）。
func (l *VideoLogic) GetByID(ctx context.Context, id string) (*res.VideoWorkRes, error) {
	video, err := l.videoModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("视频作品不存在")
	}

	links, err := l.platformLinkModel.GetByVideoID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询平台链接失败: %w", err)
	}

	return l.toVideoWorkRes(video, links), nil
}

// GetPublicList 获取已发布视频作品列表（强制 status=1，含平台链接）。
func (l *VideoLogic) GetPublicList(ctx context.Context, r *req.VideoListReq) (*res.VideoWorkListRes, error) {
	published := 1
	r.Status = &published
	return l.GetList(ctx, r)
}

// GetPublicDetail 获取已发布视频作品详情（验证 status=1，含平台链接）。
func (l *VideoLogic) GetPublicDetail(ctx context.Context, id string) (map[string]interface{}, error) {
	detail, err := l.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("视频作品不存在")
	}
	if detail.Status != 1 {
		return nil, fmt.Errorf("视频作品不存在")
	}
	result := map[string]interface{}{
		"id":          detail.ID,
		"title":       detail.Title,
		"cover_url":   detail.CoverURL,
		"description": detail.Description,
		"status":      detail.Status,
		"sort_order":  detail.SortOrder,
		"platforms":   detail.Platforms,
		"created_at":  detail.CreatedAt,
		"updated_at":  detail.UpdatedAt,
	}
	return result, nil
}

// Update 更新视频作品，若 Platforms 非空则事务内重建平台链接。
func (l *VideoLogic) Update(ctx context.Context, id string, r *req.UpdateVideoReq) (*res.VideoWorkRes, error) {
	video, err := l.videoModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("视频作品不存在")
	}

	if r.Title != nil {
		video.Title = *r.Title
	}
	if r.CoverURL != nil {
		video.CoverURL = *r.CoverURL
	}
	if r.Description != nil {
		video.Description = *r.Description
	}
	if r.Status != nil {
		video.Status = *r.Status
	}

	if r.Platforms != nil {
		links := make([]model.VideoPlatformLink, 0, len(r.Platforms))
		for _, p := range r.Platforms {
			links = append(links, model.VideoPlatformLink{
				ID:       uuid.New().String(),
				VideoID:  id,
				Platform: p.Platform,
				URL:      p.URL,
			})
		}

		if err := l.videoModel.Transaction(ctx, func(tx *gorm.DB) error {
			if err := l.videoModel.Update(ctx, video); err != nil {
				return fmt.Errorf("更新视频作品失败: %w", err)
			}
			if err := l.platformLinkModel.SoftDeleteByVideoID(ctx, tx, id); err != nil {
				return fmt.Errorf("软删除旧平台链接失败: %w", err)
			}
			if err := l.platformLinkModel.BatchCreate(ctx, tx, links); err != nil {
				return fmt.Errorf("创建新平台链接失败: %w", err)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	} else {
		if err := l.videoModel.Update(ctx, video); err != nil {
			return nil, fmt.Errorf("更新视频作品失败: %w", err)
		}
	}

	links, err := l.platformLinkModel.GetByVideoID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询平台链接失败: %w", err)
	}

	return l.toVideoWorkRes(video, links), nil
}

// Delete 删除视频作品，事务内级联软删除平台链接。
func (l *VideoLogic) Delete(ctx context.Context, id string) error {
	if _, err := l.videoModel.GetByID(ctx, id); err != nil {
		return fmt.Errorf("视频作品不存在")
	}

	return l.videoModel.Transaction(ctx, func(tx *gorm.DB) error {
		if err := l.platformLinkModel.SoftDeleteByVideoID(ctx, tx, id); err != nil {
			return fmt.Errorf("软删除平台链接失败: %w", err)
		}
		if err := l.videoModel.SoftDelete(ctx, id); err != nil {
			return fmt.Errorf("软删除视频作品失败: %w", err)
		}
		return nil
	})
}

// ParseVideo 解析视频元信息
func (l *VideoLogic) ParseVideo(ctx context.Context, r *req.ParseVideoReq) (*res.ParseVideoRes, error) {
	meta, err := ParseVideoMeta(ctx, r.Platform, r.URL)
	if err != nil {
		return nil, fmt.Errorf("解析视频信息失败: %w", err)
	}
	return &res.ParseVideoRes{
		Title:       meta.Title,
		CoverURL:    meta.CoverURL,
		Description: meta.Description,
	}, nil
}

// toVideoWorkRes 转换为视频作品响应。
func (l *VideoLogic) toVideoWorkRes(v *model.VideoWork, links []model.VideoPlatformLink) *res.VideoWorkRes {
	platforms := make([]res.VideoPlatformLinkRes, 0, len(links))
	for _, link := range links {
		platforms = append(platforms, res.VideoPlatformLinkRes{
			ID:        link.ID,
			VideoID:   link.VideoID,
			Platform:  link.Platform,
			URL:       link.URL,
			CreatedAt: link.CreatedAt,
			UpdatedAt: link.UpdatedAt,
		})
	}
	return &res.VideoWorkRes{
		ID:          v.ID,
		Title:       v.Title,
		CoverURL:    v.CoverURL,
		Description: v.Description,
		Status:      v.Status,
		SortOrder:   v.SortOrder,
		Platforms:   platforms,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}
