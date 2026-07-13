package logic

import (
	"context"
	"fmt"

	"cus-cms/internal/dto/req"
	"cus-cms/internal/dto/res"
	"cus-cms/internal/model"

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
	return l.model.Delete(ctx, id)
}
