// 内容分类类型常量：统一分类表 categories.type 的合法取值（后端唯一权威）。
// 各内容模块的分类共用同一张 categories 表，按 type 区分；公开接口的 type 白名单、
// 管理端创建/更新校验、API 文档均引用本清单。新增内容模块接入分类时在此登记。
package enum

// 分类类型值（categories.type）。
const (
	CategoryTypeArticle    = "article"
	CategoryTypeTravel     = "travel"
	CategoryTypePortfolio  = "portfolio"
	CategoryTypeMusic      = "music"
	CategoryTypeVideo      = "video"
	CategoryTypeBook       = "book"
	CategoryTypeRecipe     = "recipe"
	CategoryTypeGame       = "game"
	CategoryTypeProject    = "project"
	CategoryTypeOpenSource = "open_source"
	CategoryTypeTechStack  = "tech_stack"
	CategoryTypeFitness    = "fitness"
)

// CategoryTypes 全部合法的分类类型值。
var CategoryTypes = []string{
	CategoryTypeArticle,
	CategoryTypeTravel,
	CategoryTypePortfolio,
	CategoryTypeMusic,
	CategoryTypeVideo,
	CategoryTypeBook,
	CategoryTypeRecipe,
	CategoryTypeGame,
	CategoryTypeProject,
	CategoryTypeOpenSource,
	CategoryTypeTechStack,
	CategoryTypeFitness,
}

// IsValidCategoryType 判断分类类型是否合法（空值由调用方决定默认值，视为合法）。
func IsValidCategoryType(categoryType string) bool {
	for _, v := range CategoryTypes {
		if v == categoryType {
			return true
		}
	}
	return false
}
