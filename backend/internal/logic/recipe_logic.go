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

// RecipeLogic 美食菜谱业务逻辑结构体。
type RecipeLogic struct {
	recipeModel *model.RecipeModel
}

// NewRecipeLogic 创建 RecipeLogic 实例。
func NewRecipeLogic() *RecipeLogic {
	return &RecipeLogic{recipeModel: model.NewRecipe()}
}

// Create 创建菜谱。
func (l *RecipeLogic) Create(ctx context.Context, r *req.CreateRecipeReq) (*res.RecipeRes, error) {
	coverMediaID, coverPresetID, cover, err := normalizeMediaRef(ctx, r.CoverMediaID, r.CoverPresetID, r.Cover)
	if err != nil {
		return nil, err
	}
	m := &model.Recipe{
		ID:            uuid.New().String(),
		Title:         strings.TrimSpace(r.Title),
		Cover:         cover,
		CoverMediaID:  mediaRefPtr(coverMediaID),
		CoverPresetID: mediaRefPtr(coverPresetID),
		Summary:       r.Summary,
		Ingredients:   emptyIfNil(r.Ingredients),
		Steps:         emptyIfNil(r.Steps),
		Tags:          r.Tags,
	}
	if r.Difficulty != nil {
		m.Difficulty = *r.Difficulty
	}
	if r.Minutes != nil {
		m.Minutes = *r.Minutes
	}
	if r.Servings != nil {
		m.Servings = *r.Servings
	}
	if r.Status != nil {
		m.Status = *r.Status
	}
	if r.SortOrder != nil {
		m.SortOrder = *r.SortOrder
	}
	if r.CategoryID != nil {
		m.CategoryID = r.CategoryID
	}

	if err := l.recipeModel.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建菜谱失败: %w", err)
	}
	SyncOnUpdatedAsync(TargetTypeRecipe, m.ID)
	return l.toRes(m), nil
}

// GetList 分页查询菜谱列表。
func (l *RecipeLogic) GetList(ctx context.Context, r *req.RecipeListReq) (*res.RecipeListRes, error) {
	list, total, err := l.recipeModel.GetList(ctx, r.Keyword, r.Difficulty, r.CategoryID, r.Status, r.GetPage(), r.GetPageSize())
	if err != nil {
		return nil, fmt.Errorf("查询菜谱列表失败: %w", err)
	}

	items := make([]res.RecipeCardRes, 0, len(list))
	l.resolveCoverURLs(ctx, list)
	for i := range list {
		items = append(items, l.toCard(&list[i]))
	}
	return &res.RecipeListRes{
		List:       items,
		Total:      total,
		Page:       r.GetPage(),
		PageSize:   r.GetPageSize(),
		TotalPages: calcTotalPages(total, r.GetPageSize()),
	}, nil
}

// GetByID 查询菜谱详情。
func (l *RecipeLogic) GetByID(ctx context.Context, id string) (*res.RecipeRes, error) {
	m, err := l.recipeModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("菜谱不存在")
	}
	l.resolveCoverURL(ctx, m)
	return l.toRes(m), nil
}

// Update 更新菜谱（部分更新语义）。
func (l *RecipeLogic) Update(ctx context.Context, id string, r *req.UpdateRecipeReq) (*res.RecipeRes, error) {
	m, err := l.recipeModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("菜谱不存在")
	}

	if r.Title != nil {
		m.Title = strings.TrimSpace(*r.Title)
	}
	if r.Cover != nil || r.CoverMediaID != nil {
		mediaID, presetID := "", ""
		if r.CoverMediaID != nil {
			mediaID = *r.CoverMediaID
		}
		if r.CoverPresetID != nil {
			presetID = *r.CoverPresetID
		}
		coverURL := ""
		if r.Cover != nil {
			coverURL = *r.Cover
		}
		coverMediaID, coverPresetID, cover, err := normalizeMediaRef(ctx, mediaID, presetID, coverURL)
		if err != nil {
			return nil, err
		}
		m.Cover = cover
		m.CoverMediaID = mediaRefPtr(coverMediaID)
		m.CoverPresetID = mediaRefPtr(coverPresetID)
	}
	if r.Summary != nil {
		m.Summary = *r.Summary
	}
	if r.Ingredients != nil {
		m.Ingredients = r.Ingredients
	}
	if r.Steps != nil {
		m.Steps = r.Steps
	}
	if r.Difficulty != nil {
		m.Difficulty = *r.Difficulty
	}
	if r.Minutes != nil {
		m.Minutes = *r.Minutes
	}
	if r.Servings != nil {
		m.Servings = *r.Servings
	}
	if r.Tags != nil {
		m.Tags = *r.Tags
	}
	if r.Status != nil {
		m.Status = *r.Status
	}
	if r.SortOrder != nil {
		m.SortOrder = *r.SortOrder
	}
	if r.CategoryID != nil {
		m.CategoryID = r.CategoryID
	}

	if err := l.recipeModel.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("更新菜谱失败: %w", err)
	}
	SyncOnUpdatedAsync(TargetTypeRecipe, m.ID)
	return l.toRes(m), nil
}

// Delete 删除菜谱。
func (l *RecipeLogic) Delete(ctx context.Context, id string) error {
	if _, err := l.recipeModel.GetByID(ctx, id); err != nil {
		return fmt.Errorf("菜谱不存在")
	}
	if err := l.recipeModel.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("删除菜谱失败: %w", err)
	}
	SyncOnDeletedAsync(TargetTypeRecipe, id)
	return nil
}

// GetPublicList 获取已发布菜谱列表（强制 status=1）。
func (l *RecipeLogic) GetPublicList(ctx context.Context, r *req.RecipeListReq) (*res.RecipeListRes, error) {
	published := 1
	r.Status = &published
	return l.GetList(ctx, r)
}

// GetPublicDetail 获取已发布菜谱详情（未发布视为不存在）。
func (l *RecipeLogic) GetPublicDetail(ctx context.Context, id string) (*res.RecipeRes, error) {
	m, err := l.recipeModel.GetByID(ctx, id)
	if err != nil || m.Status != 1 {
		return nil, fmt.Errorf("菜谱不存在")
	}
	l.resolveCoverURL(ctx, m)
	return l.toRes(m), nil
}

// resolveCoverURLs 批量解析菜谱封面媒体引用为当前完整 URL（外链行保持原值）。
func (l *RecipeLogic) resolveCoverURLs(ctx context.Context, recipes []model.Recipe) {
	ptrs := make([]*model.Recipe, len(recipes))
	for i := range recipes {
		ptrs[i] = &recipes[i]
	}
	l.resolveCoverPtrs(ctx, ptrs)
}

// resolveCoverURL 解析单条菜谱封面媒体引用（原指针，解析结果直接写回）。
func (l *RecipeLogic) resolveCoverURL(ctx context.Context, m *model.Recipe) {
	l.resolveCoverPtrs(ctx, []*model.Recipe{m})
}

// resolveCoverPtrs 按指针切片解析封面。
func (l *RecipeLogic) resolveCoverPtrs(ctx context.Context, ptrs []*model.Recipe) {
	model.ResolveCoverURLs(ctx, ptrs,
		func(r *model.Recipe) (string, string) { return refStr(r.CoverMediaID), refStr(r.CoverPresetID) },
		func(r *model.Recipe, u string) { r.Cover = u },
	)
}

// emptyIfNil jsonb 数组字段为 nil 时返回空切片，避免入库为 JSON null。
func emptyIfNil(items []map[string]any) []map[string]any {
	if items == nil {
		return []map[string]any{}
	}
	return items
}

// toRes 转换为菜谱详情响应。
func (l *RecipeLogic) toRes(m *model.Recipe) *res.RecipeRes {
	return &res.RecipeRes{
		ID:            m.ID,
		Title:         m.Title,
		Cover:         m.Cover,
		CoverMediaID:  refStr(m.CoverMediaID),
		CoverPresetID: refStr(m.CoverPresetID),
		Summary:       m.Summary,
		Ingredients:   m.Ingredients,
		Steps:         m.Steps,
		Difficulty:    m.Difficulty,
		Minutes:       m.Minutes,
		Servings:      m.Servings,
		Tags:          m.Tags,
		Status:        m.Status,
		SortOrder:     m.SortOrder,
		CategoryID:    categoryIDOf(m.CategoryID),
		CategoryName:  categoryNameOf(m.Category),
		ViewCount:     int64(m.ViewCount),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// toCard 转换为菜谱列表卡片响应。
func (l *RecipeLogic) toCard(m *model.Recipe) res.RecipeCardRes {
	return res.RecipeCardRes{
		ID:           m.ID,
		Title:        m.Title,
		Cover:        m.Cover,
		Summary:      m.Summary,
		Difficulty:   m.Difficulty,
		Minutes:      m.Minutes,
		Servings:     m.Servings,
		Tags:         m.Tags,
		Status:       m.Status,
		SortOrder:    m.SortOrder,
		CategoryID:   categoryIDOf(m.CategoryID),
		CategoryName: categoryNameOf(m.Category),
		ViewCount:    int64(m.ViewCount),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
