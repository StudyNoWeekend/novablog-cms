package logic

import (
	"context"
	"fmt"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"

	"github.com/google/uuid"
)

// EquipmentLogic 个人设备业务逻辑结构体。
type EquipmentLogic struct {
	equipmentModel *model.PhotoEquipmentModel
}

// NewEquipmentLogic 创建 EquipmentLogic 实例。
func NewEquipmentLogic() *EquipmentLogic {
	return &EquipmentLogic{
		equipmentModel: model.NewPhotoEquipment(),
	}
}

// Create 创建个人设备。
func (l *EquipmentLogic) Create(ctx context.Context, r *req.CreateEquipmentReq) (*res.EquipmentRes, error) {
	equipment := &model.PhotoEquipment{
		ID:          uuid.New().String(),
		Name:        r.Name,
		ImageURL:    r.ImageURL,
		Brand:       r.Brand,
		Description: r.Description,
	}

	if err := l.equipmentModel.Create(ctx, equipment); err != nil {
		return nil, fmt.Errorf("创建个人设备失败: %w", err)
	}

	return l.toEquipmentRes(equipment), nil
}

// GetList 分页查询个人设备列表。
func (l *EquipmentLogic) GetList(ctx context.Context, r *req.EquipmentListReq) (*res.EquipmentListRes, error) {
	list, total, err := l.equipmentModel.GetList(ctx, r.Keyword, r.Brand, r.GetPage(), r.GetPageSize())
	if err != nil {
		return nil, fmt.Errorf("查询个人设备列表失败: %w", err)
	}

	items := make([]res.EquipmentRes, 0, len(list))
	for i := range list {
		items = append(items, *l.toEquipmentRes(&list[i]))
	}

	return &res.EquipmentListRes{
		List:       items,
		Total:      total,
		Page:       r.GetPage(),
		PageSize:   r.GetPageSize(),
		TotalPages: calcTotalPages(total, r.GetPageSize()),
	}, nil
}

// GetByID 查询个人设备详情。
func (l *EquipmentLogic) GetByID(ctx context.Context, id string) (*res.EquipmentRes, error) {
	equipment, err := l.equipmentModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("个人设备不存在")
	}

	return l.toEquipmentRes(equipment), nil
}

// Update 更新个人设备。
func (l *EquipmentLogic) Update(ctx context.Context, id string, r *req.UpdateEquipmentReq) (*res.EquipmentRes, error) {
	equipment, fetchErr := l.equipmentModel.GetByID(ctx, id)
	if fetchErr != nil {
		return nil, fmt.Errorf("个人设备不存在")
	}

	if r.Name != nil {
		equipment.Name = *r.Name
	}
	if r.ImageURL != nil {
		equipment.ImageURL = *r.ImageURL
	}
	if r.Brand != nil {
		equipment.Brand = *r.Brand
	}
	if r.Description != nil {
		equipment.Description = *r.Description
	}

	if updateErr := l.equipmentModel.Update(ctx, equipment); updateErr != nil {
		return nil, fmt.Errorf("更新个人设备失败: %w", updateErr)
	}

	return l.toEquipmentRes(equipment), nil
}

// Delete 删除个人设备。
func (l *EquipmentLogic) Delete(ctx context.Context, id string) error {
	if _, err := l.equipmentModel.GetByID(ctx, id); err != nil {
		return fmt.Errorf("个人设备不存在")
	}

	if err := l.equipmentModel.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("软删除个人设备失败: %w", err)
	}

	return nil
}

// toEquipmentRes 转换为个人设备响应。
func (l *EquipmentLogic) toEquipmentRes(e *model.PhotoEquipment) *res.EquipmentRes {
	return &res.EquipmentRes{
		ID:          e.ID,
		Name:        e.Name,
		ImageURL:    e.ImageURL,
		Brand:       e.Brand,
		Description: e.Description,
		SortOrder:   e.SortOrder,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}
