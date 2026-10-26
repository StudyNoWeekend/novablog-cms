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
	imageMediaID, imagePresetID, imageURL, err := normalizeMediaRef(ctx, r.ImageMediaID, r.ImagePresetID, r.ImageURL)
	if err != nil {
		return nil, err
	}
	equipment := &model.PhotoEquipment{
		ID:            uuid.New().String(),
		Name:          r.Name,
		ImageURL:      imageURL,
		ImageMediaID:  mediaRefPtr(imageMediaID),
		ImagePresetID: mediaRefPtr(imagePresetID),
		Brand:         r.Brand,
		Description:   r.Description,
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
	l.resolveImageURLs(ctx, list)
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
	l.resolveImageURL(ctx, equipment)

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
	if r.ImageURL != nil || r.ImageMediaID != nil {
		mediaID, presetID := "", ""
		if r.ImageMediaID != nil {
			mediaID = *r.ImageMediaID
		}
		if r.ImagePresetID != nil {
			presetID = *r.ImagePresetID
		}
		imageURL := ""
		if r.ImageURL != nil {
			imageURL = *r.ImageURL
		}
		imageMediaID, imagePresetID, url, err := normalizeMediaRef(ctx, mediaID, presetID, imageURL)
		if err != nil {
			return nil, err
		}
		equipment.ImageURL = url
		equipment.ImageMediaID = mediaRefPtr(imageMediaID)
		equipment.ImagePresetID = mediaRefPtr(imagePresetID)
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

// resolveImageURLs 批量解析设备图片媒体引用为当前完整 URL（外链行保持原值）。
func (l *EquipmentLogic) resolveImageURLs(ctx context.Context, equipments []model.PhotoEquipment) {
	ptrs := make([]*model.PhotoEquipment, len(equipments))
	for i := range equipments {
		ptrs[i] = &equipments[i]
	}
	model.ResolveCoverURLs(ctx, ptrs,
		func(e *model.PhotoEquipment) (string, string) { return refStr(e.ImageMediaID), refStr(e.ImagePresetID) },
		func(e *model.PhotoEquipment, u string) { e.ImageURL = u },
	)
}

// resolveImageURL 解析单台设备图片媒体引用。
func (l *EquipmentLogic) resolveImageURL(ctx context.Context, e *model.PhotoEquipment) {
	model.ResolveCoverURLs(ctx, []*model.PhotoEquipment{e},
		func(x *model.PhotoEquipment) (string, string) { return refStr(x.ImageMediaID), refStr(x.ImagePresetID) },
		func(x *model.PhotoEquipment, u string) { x.ImageURL = u },
	)
}

// toEquipmentRes 转换为个人设备响应。
func (l *EquipmentLogic) toEquipmentRes(e *model.PhotoEquipment) *res.EquipmentRes {
	return &res.EquipmentRes{
		ID:            e.ID,
		Name:          e.Name,
		ImageURL:      e.ImageURL,
		ImageMediaID:  refStr(e.ImageMediaID),
		ImagePresetID: refStr(e.ImagePresetID),
		Brand:         e.Brand,
		Description:   e.Description,
		SortOrder:     e.SortOrder,
		ViewCount:     int64(e.ViewCount),
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}
