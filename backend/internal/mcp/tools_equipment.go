package mcp

import (
	"context"
	"fmt"
	"strings"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerEquipmentTools 注册个人设备工具（无状态、无分类，仅 name 必填）。
func (s *Service) registerEquipmentTools(server *sdk.Server) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_equipments",
		Description: "分页查询个人设备列表（摄影器材等，含图片、品牌、描述与排序）。",
	}, wrapTool(s, "list_equipments", s.listEquipments))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_equipment",
		Description: "新增一台个人设备（name 必填，image_url 可先调用 upload_image 上传获取）。",
	}, wrapTool(s, "create_equipment", s.createEquipment))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_equipment",
		Description: "更新设备信息（整体替换传入了的字段，未传入的保持不变）。",
	}, wrapTool(s, "update_equipment", s.updateEquipment))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_equipment",
		Description: "删除设备（软删除）。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_equipment", s.deleteEquipment))
}

// ListEquipmentsIn list_equipments 入参。
type ListEquipmentsIn struct {
	Keyword  string `json:"keyword,omitempty" jsonschema:"按设备名称模糊搜索"`
	Brand    string `json:"brand,omitempty" jsonschema:"按品牌精确筛选"`
	Page     int    `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize int    `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// ListEquipmentsOut list_equipments 出参。
type ListEquipmentsOut struct {
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Items []res.EquipmentRes `json:"items"`
}

// listEquipments 分页查询设备列表。
func (s *Service) listEquipments(ctx context.Context, _ *sdk.CallToolRequest, in ListEquipmentsIn) (*sdk.CallToolResult, ListEquipmentsOut, error) {
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	page, err := s.equipmentLogic.GetList(ctx, &req.EquipmentListReq{
		PageReq: req.PageReq{Page: in.Page, PageSize: in.PageSize},
		Keyword: stringPtr(in.Keyword),
		Brand:   stringPtr(in.Brand),
	})
	if err != nil {
		return nil, ListEquipmentsOut{}, fmt.Errorf("查询设备列表失败: %w", err)
	}
	return nil, ListEquipmentsOut{Total: page.Total, Page: page.Page, Items: page.List}, nil
}

// CreateEquipmentIn create_equipment 入参。
type CreateEquipmentIn struct {
	Name        string `json:"name" jsonschema:"设备名称（必填，1-255 字）"`
	ImageURL    string `json:"image_url,omitempty" jsonschema:"设备图片 URL，可先调用 upload_image 上传获取"`
	Brand       string `json:"brand,omitempty" jsonschema:"品牌"`
	Description string `json:"description,omitempty" jsonschema:"设备描述"`
}

// createEquipment 新增设备。
func (s *Service) createEquipment(ctx context.Context, _ *sdk.CallToolRequest, in CreateEquipmentIn) (*sdk.CallToolResult, res.EquipmentRes, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, res.EquipmentRes{}, fmt.Errorf("name 不能为空")
	}
	created, err := s.equipmentLogic.Create(ctx, &req.CreateEquipmentReq{
		Name:        in.Name,
		ImageURL:    in.ImageURL,
		Brand:       in.Brand,
		Description: in.Description,
	})
	if err != nil {
		return nil, res.EquipmentRes{}, fmt.Errorf("创建设备失败: %w", err)
	}
	return nil, *created, nil
}

// UpdateEquipmentIn update_equipment 入参（指针字段=不传不修改）。
type UpdateEquipmentIn struct {
	ID          string  `json:"id" jsonschema:"要更新的设备 ID"`
	Name        *string `json:"name,omitempty" jsonschema:"新名称"`
	ImageURL    *string `json:"image_url,omitempty" jsonschema:"新图片 URL"`
	Brand       *string `json:"brand,omitempty" jsonschema:"新品牌"`
	Description *string `json:"description,omitempty" jsonschema:"新描述"`
}

// updateEquipment 更新设备。
func (s *Service) updateEquipment(ctx context.Context, _ *sdk.CallToolRequest, in UpdateEquipmentIn) (*sdk.CallToolResult, res.EquipmentRes, error) {
	if in.ID == "" {
		return nil, res.EquipmentRes{}, fmt.Errorf("id 不能为空")
	}
	updated, err := s.equipmentLogic.Update(ctx, in.ID, &req.UpdateEquipmentReq{
		Name:        in.Name,
		ImageURL:    in.ImageURL,
		Brand:       in.Brand,
		Description: in.Description,
	})
	if err != nil {
		return nil, res.EquipmentRes{}, fmt.Errorf("更新设备失败: %w", err)
	}
	return nil, *updated, nil
}

// deleteEquipment 删除设备。
func (s *Service) deleteEquipment(ctx context.Context, _ *sdk.CallToolRequest, in DeleteIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.ID == "" {
		return nil, DeleteOut{}, fmt.Errorf("id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteOut{}, fmt.Errorf("删除是不可逆操作，请确认后显式传 confirm=true 再执行")
	}
	if err := s.equipmentLogic.Delete(ctx, in.ID); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("删除设备失败: %w", err)
	}
	return nil, DeleteOut{Deleted: true, ID: in.ID}, nil
}
