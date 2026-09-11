package logic

import (
	"context"
	"fmt"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"

	"github.com/google/uuid"
)

type TagLogic struct {
	model *model.TagModel
}

func NewTagLogic() *TagLogic {
	return &TagLogic{model: model.NewTag()}
}

func (l *TagLogic) Create(ctx context.Context, r *req.CreateTagReq) (*res.TagRes, error) {
	existing, _ := l.model.GetByName(ctx, r.Name)
	if existing != nil {
		return nil, fmt.Errorf("标签名称已存在")
	}

	tag := &model.Tag{
		ID:   uuid.New().String(),
		Name: r.Name,
	}

	if err := l.model.Create(ctx, tag); err != nil {
		return nil, fmt.Errorf("创建标签失败: %w", err)
	}

	return &res.TagRes{
		ID:        tag.ID,
		Name:      tag.Name,
		CreatedAt: tag.CreatedAt,
	}, nil
}

func (l *TagLogic) GetAll(ctx context.Context) ([]res.TagRes, error) {
	tags, err := l.model.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	var result []res.TagRes
	for _, t := range tags {
		result = append(result, res.TagRes{
			ID:        t.ID,
			Name:      t.Name,
			CreatedAt: t.CreatedAt,
		})
	}
	if result == nil {
		result = []res.TagRes{}
	}
	return result, nil
}

func (l *TagLogic) GetByID(ctx context.Context, id string) (*res.TagRes, error) {
	t, err := l.model.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("标签不存在")
	}
	return &res.TagRes{
		ID:        t.ID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
	}, nil
}

func (l *TagLogic) Update(ctx context.Context, id string, r *req.UpdateTagReq) (*res.TagRes, error) {
	t, err := l.model.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("标签不存在")
	}

	existing, _ := l.model.GetByName(ctx, r.Name)
	if existing != nil && existing.ID != id {
		return nil, fmt.Errorf("标签名称已存在")
	}
	t.Name = r.Name

	if err := l.model.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("更新标签失败: %w", err)
	}

	return &res.TagRes{
		ID:        t.ID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
	}, nil
}

func (l *TagLogic) Delete(ctx context.Context, id string) error {
	// Check if any articles use this tag
	count, err := l.model.CountArticlesByTagID(ctx, id)
	if err != nil {
		return fmt.Errorf("查询标签关联失败: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("该标签已被 %d 篇文章使用，请先移除关联后再删除", count)
	}
	if err := l.model.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除标签失败: %w", err)
	}
	return nil
}
