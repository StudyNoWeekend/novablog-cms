package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// MediaFolder 媒体文件夹模型，对应 media_folders 数据表。
// ParentID 为空表示顶级文件夹；ModuleKey 标记由业务模块播种的文件夹（上传时按
// module key 自动归属），用户自建文件夹该字段为空。
type MediaFolder struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:varchar(50);not null"`
	ParentID  *string   `gorm:"type:uuid;index;column:parent_id"`
	ModuleKey *string   `gorm:"type:varchar(50);column:module_key"`
	Sort      int       `gorm:"type:int;default:0"`
	CreatedAt time.Time `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt time.Time `gorm:"type:timestamptz;autoUpdateTime"`
}

// TableName 指定数据表名称。
func (MediaFolder) TableName() string {
	return "media_folders"
}

// MediaFolderModel 媒体文件夹模型操作结构体。
type MediaFolderModel struct {
	db *gorm.DB
}

// NewMediaFolder 创建 MediaFolderModel 实例。
func NewMediaFolder() *MediaFolderModel {
	return &MediaFolderModel{db: DB}
}

// Create 创建文件夹记录。
func (m *MediaFolderModel) Create(ctx context.Context, folder *MediaFolder) error {
	return m.db.WithContext(ctx).Create(folder).Error
}

// GetByID 根据 ID 查询文件夹。
func (m *MediaFolderModel) GetByID(ctx context.Context, id string) (*MediaFolder, error) {
	var folder MediaFolder
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

// GetByIDs 按 ID 列表批量查询文件夹。
func (m *MediaFolderModel) GetByIDs(ctx context.Context, ids []string) ([]MediaFolder, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var folders []MediaFolder
	err := m.db.WithContext(ctx).Where("id IN ?", ids).Find(&folders).Error
	if err != nil {
		return nil, err
	}
	return folders, nil
}

// GetAll 查询全部文件夹（按排序权重、创建时间升序）。
func (m *MediaFolderModel) GetAll(ctx context.Context) ([]MediaFolder, error) {
	var folders []MediaFolder
	err := m.db.WithContext(ctx).Order("sort ASC, created_at ASC").Find(&folders).Error
	if err != nil {
		return nil, err
	}
	return folders, nil
}

// GetTopByModuleKey 按模块 key 查询顶级模块文件夹。
func (m *MediaFolderModel) GetTopByModuleKey(ctx context.Context, moduleKey string) (*MediaFolder, error) {
	var folder MediaFolder
	err := m.db.WithContext(ctx).
		Where("parent_id IS NULL AND module_key = ?", moduleKey).
		First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

// GetTopByName 按名称查询顶级文件夹（模块文件夹 module_key 缺失时的兜底匹配）。
func (m *MediaFolderModel) GetTopByName(ctx context.Context, name string) (*MediaFolder, error) {
	var folder MediaFolder
	err := m.db.WithContext(ctx).
		Where("parent_id IS NULL AND name = ?", name).
		First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

// ExistsNameUnder 检查同一父文件夹下是否已存在同名文件夹（excludeID 用于重命名时排除自身）。
func (m *MediaFolderModel) ExistsNameUnder(ctx context.Context, parentID *string, name string, excludeID string) (bool, error) {
	query := m.db.WithContext(ctx).Model(&MediaFolder{}).Where("name = ?", name)
	if parentID == nil || *parentID == "" {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Update 保存文件夹的名称/父级变更。
func (m *MediaFolderModel) Update(ctx context.Context, folder *MediaFolder) error {
	return m.db.WithContext(ctx).Model(folder).Select("name", "parent_id", "updated_at").Updates(folder).Error
}

// Delete 物理删除文件夹记录（调用前需确认文件夹为空）。
func (m *MediaFolderModel) Delete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&MediaFolder{}).Error
}

// CountSubFolders 统计文件夹下的直接子文件夹数量。
func (m *MediaFolderModel) CountSubFolders(ctx context.Context, id string) (int64, error) {
	var count int64
	err := m.db.WithContext(ctx).Model(&MediaFolder{}).Where("parent_id = ?", id).Count(&count).Error
	return count, err
}

// CountMedia 统计文件夹下的直接媒体数量。
func (m *MediaFolderModel) CountMedia(ctx context.Context, id string) (int64, error) {
	var count int64
	err := m.db.WithContext(ctx).Model(&Media{}).Where("folder_id = ?", id).Count(&count).Error
	return count, err
}

// CountMediaByFolder 统计各文件夹下的直接媒体数量，返回 folder_id → 数量映射。
func (m *MediaFolderModel) CountMediaByFolder(ctx context.Context) (map[string]int64, error) {
	type row struct {
		FolderID string `gorm:"column:folder_id"`
		Count    int64  `gorm:"column:count"`
	}
	var rows []row
	err := m.db.WithContext(ctx).Model(&Media{}).
		Select("folder_id, COUNT(*) AS count").
		Where("folder_id IS NOT NULL").
		Group("folder_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(rows))
	for _, r := range rows {
		result[r.FolderID] = r.Count
	}
	return result, nil
}

// AncestorNames 返回从顶级到当前文件夹的名称链（含自身），用于物理存储路径前缀。
func (m *MediaFolderModel) AncestorNames(ctx context.Context, folder *MediaFolder) ([]string, error) {
	names := []string{folder.Name}
	current := folder
	for current.ParentID != nil {
		if len(names) >= MediaFolderMaxDepth {
			break // 深度兜底，防脏数据成环
		}
		parent, err := m.GetByID(ctx, *current.ParentID)
		if err != nil {
			return nil, err
		}
		names = append([]string{parent.Name}, names...)
		current = parent
	}
	return names, nil
}

// Depth 返回文件夹所处层级（顶级为 1）。
func (m *MediaFolderModel) Depth(ctx context.Context, folder *MediaFolder) (int, error) {
	depth := 1
	current := folder
	for current.ParentID != nil {
		if depth >= MediaFolderMaxDepth {
			break
		}
		parent, err := m.GetByID(ctx, *current.ParentID)
		if err != nil {
			return 0, err
		}
		depth++
		current = parent
	}
	return depth, nil
}

// MediaFolderMaxDepth 文件夹最大层级深度。
const MediaFolderMaxDepth = 5
