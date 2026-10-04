package logic

import (
	"context"
	"fmt"
	"strings"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"

	"github.com/google/uuid"
)

// folderForbiddenRunes 文件夹名称禁止字符（含路径分隔符与 Windows 保留字符）。
const folderForbiddenRunes = `/\:*?"<>|`

// MediaFolderLogic 媒体文件夹业务逻辑结构体。
type MediaFolderLogic struct {
	folderModel *model.MediaFolderModel
}

// NewMediaFolderLogic 创建 MediaFolderLogic 实例。
func NewMediaFolderLogic() *MediaFolderLogic {
	return &MediaFolderLogic{folderModel: model.NewMediaFolder()}
}

// Tree 获取完整文件夹树（含各文件夹直接媒体数）。
func (l *MediaFolderLogic) Tree(ctx context.Context) (*res.MediaFolderTreeRes, error) {
	folders, err := l.folderModel.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询文件夹失败: %w", err)
	}
	counts, err := l.folderModel.CountMediaByFolder(ctx)
	if err != nil {
		return nil, fmt.Errorf("统计文件夹媒体数失败: %w", err)
	}

	nodes := make([]res.MediaFolderRes, len(folders))
	nodeMap := make(map[string]bool, len(folders))
	for i, f := range folders {
		nodes[i] = res.MediaFolderRes{
			ID:         f.ID,
			Name:       f.Name,
			ParentID:   f.ParentID,
			ModuleKey:  f.ModuleKey,
			MediaCount: counts[f.ID],
			Children:   []res.MediaFolderRes{},
			CreatedAt:  f.CreatedAt,
		}
		nodeMap[f.ID] = true
	}

	// 父子索引（父缺失的文件夹按顶级兜底展示）
	childIdx := make(map[string][]int, len(folders))
	rootIdx := make([]int, 0)
	for i, f := range folders {
		if f.ParentID != nil && nodeMap[*f.ParentID] {
			childIdx[*f.ParentID] = append(childIdx[*f.ParentID], i)
		} else {
			rootIdx = append(rootIdx, i)
		}
	}

	var build func(idx int) res.MediaFolderRes
	build = func(idx int) res.MediaFolderRes {
		node := nodes[idx]
		for _, ci := range childIdx[node.ID] {
			node.Children = append(node.Children, build(ci))
		}
		return node
	}

	roots := make([]res.MediaFolderRes, 0, len(rootIdx))
	for _, ri := range rootIdx {
		roots = append(roots, build(ri))
	}
	return &res.MediaFolderTreeRes{List: roots}, nil
}

// Create 创建文件夹。
func (l *MediaFolderLogic) Create(ctx context.Context, r *req.MediaFolderCreateReq) (*res.MediaFolderRes, error) {
	name, err := sanitizeFolderName(r.Name)
	if err != nil {
		return nil, err
	}

	parentID := r.ParentID
	if parentID != nil && *parentID == "" {
		parentID = nil
	}
	if parentID != nil {
		parent, err := l.folderModel.GetByID(ctx, *parentID)
		if err != nil {
			return nil, fmt.Errorf("父文件夹不存在: %w", err)
		}
		depth, err := l.folderModel.Depth(ctx, parent)
		if err != nil {
			return nil, err
		}
		if depth >= model.MediaFolderMaxDepth {
			return nil, fmt.Errorf("文件夹层级已达上限（%d 级），无法继续创建子文件夹", model.MediaFolderMaxDepth)
		}
	}

	exists, err := l.folderModel.ExistsNameUnder(ctx, parentID, name, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("同级文件夹中已存在同名文件夹「%s」", name)
	}

	folder := &model.MediaFolder{ID: uuid.New().String(), Name: name, ParentID: parentID}
	if err := l.folderModel.Create(ctx, folder); err != nil {
		return nil, fmt.Errorf("创建文件夹失败: %w", err)
	}
	return folderToRes(folder, 0), nil
}

// Update 更新文件夹（重命名和/或移动）。
func (l *MediaFolderLogic) Update(ctx context.Context, id string, r *req.MediaFolderUpdateReq) (*res.MediaFolderRes, error) {
	folder, err := l.folderModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("文件夹不存在: %w", err)
	}

	// 系统文件夹（模块播种，ModuleKey 非空）的名称参与存储路径前缀与上传归属链路，
	// 重命名/移动会导致未来上传的文件 URL 与既有文件脱节，一律拒绝（无实际变更的幂等请求放行）。
	if folder.ModuleKey != nil {
		if r.Name != nil {
			return nil, fmt.Errorf("系统文件夹由模块固定使用，不支持重命名")
		}
		if r.ParentID != nil && !sameParentID(folder.ParentID, *r.ParentID) {
			return nil, fmt.Errorf("系统文件夹由模块固定使用，不支持移动")
		}
	}

	if r.Name != nil {
		name, err := sanitizeFolderName(*r.Name)
		if err != nil {
			return nil, err
		}
		exists, err := l.folderModel.ExistsNameUnder(ctx, folder.ParentID, name, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("同级文件夹中已存在同名文件夹「%s」", name)
		}
		folder.Name = name
	}

	if r.ParentID != nil {
		newParentID := *r.ParentID
		switch {
		case newParentID == "":
			folder.ParentID = nil
		case newParentID == id:
			return nil, fmt.Errorf("不能将文件夹移动到自身")
		default:
			parent, err := l.folderModel.GetByID(ctx, newParentID)
			if err != nil {
				return nil, fmt.Errorf("目标文件夹不存在: %w", err)
			}
			if err := l.ensureNotDescendant(ctx, parent, id); err != nil {
				return nil, err
			}
			depth, err := l.folderModel.Depth(ctx, parent)
			if err != nil {
				return nil, err
			}
			if depth >= model.MediaFolderMaxDepth {
				return nil, fmt.Errorf("移动后超出最大层级（%d 级）", model.MediaFolderMaxDepth)
			}
			folder.ParentID = &newParentID
		}
	}

	if err := l.folderModel.Update(ctx, folder); err != nil {
		return nil, fmt.Errorf("更新文件夹失败: %w", err)
	}
	count, err := l.folderModel.CountMedia(ctx, id)
	if err != nil {
		return nil, err
	}
	return folderToRes(folder, count), nil
}

// Delete 删除空文件夹（存在子文件夹或媒体时拒绝）。
func (l *MediaFolderLogic) Delete(ctx context.Context, id string) error {
	folder, err := l.folderModel.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("文件夹不存在: %w", err)
	}
	// 系统文件夹承载模块固定素材，删除后模块上传会重建出空文件夹，造成素材"消失"的假象。
	if folder.ModuleKey != nil {
		return fmt.Errorf("系统文件夹由模块固定使用，不支持删除")
	}
	subs, err := l.folderModel.CountSubFolders(ctx, id)
	if err != nil {
		return err
	}
	if subs > 0 {
		return fmt.Errorf("文件夹内还有子文件夹，请先删除子文件夹")
	}
	mediaCount, err := l.folderModel.CountMedia(ctx, id)
	if err != nil {
		return err
	}
	if mediaCount > 0 {
		return fmt.Errorf("文件夹内还有 %d 个文件，请先移动或删除", mediaCount)
	}
	return l.folderModel.Delete(ctx, id)
}

// ensureNotDescendant 校验目标文件夹不是 selfID 文件夹自身的后代，防止移动成环。
func (l *MediaFolderLogic) ensureNotDescendant(ctx context.Context, target *model.MediaFolder, selfID string) error {
	current := target
	for {
		if current.ID == selfID {
			return fmt.Errorf("不能将文件夹移动到其子文件夹内")
		}
		if current.ParentID == nil {
			return nil
		}
		parent, err := l.folderModel.GetByID(ctx, *current.ParentID)
		if err != nil {
			return fmt.Errorf("查询父文件夹失败: %w", err)
		}
		current = parent
	}
}

// folderToRes 构建文件夹响应。
func folderToRes(folder *model.MediaFolder, mediaCount int64) *res.MediaFolderRes {
	return &res.MediaFolderRes{
		ID:         folder.ID,
		Name:       folder.Name,
		ParentID:   folder.ParentID,
		ModuleKey:  folder.ModuleKey,
		MediaCount: mediaCount,
		Children:   []res.MediaFolderRes{},
		CreatedAt:  folder.CreatedAt,
	}
}

// sameParentID 判断请求中的目标父级与当前父级是否一致（空串与 nil 均表示根目录）。
func sameParentID(current *string, target string) bool {
	if target == "" {
		return current == nil
	}
	return current != nil && *current == target
}

// sanitizeFolderName 校验并清理文件夹名称。
func sanitizeFolderName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("文件夹名称不能为空")
	}
	if name == "." || name == ".." {
		return "", fmt.Errorf("文件夹名称不合法")
	}
	if len([]rune(name)) > 50 {
		return "", fmt.Errorf("文件夹名称不能超过 50 个字符")
	}
	if strings.ContainsAny(name, folderForbiddenRunes) {
		return "", fmt.Errorf("文件夹名称不能包含以下字符：%s", folderForbiddenRunes)
	}
	return name, nil
}
