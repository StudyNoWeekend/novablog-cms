// 分类数据启动回填：项目没有自动执行 migrations/*.sql 的机制（AutoMigrate 只同步表结构），
// 因此数据级迁移在此幂等执行。本文件内容与 migrations/037、038 的 up 脚本一致，可重复运行，
// 与手工执行 SQL 互不冲突；任何失败仅跳过对应步骤，不阻塞服务启动。
package bootstrap

import (
	"fmt"

	"novablog/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// backfillCategories 把项目经历/技术栈/健身的旧分类数据回填到统一分类表。
// 先做一次廉价探测：相关表缺失、无待回填数据且 slug 唯一性已归一时直接跳过，
// 避免每次启动都重复执行迁移逻辑（探测本身只花几次目录查询）。
func backfillCategories(db *gorm.DB) {
	if !categoryBackfillNeeded(db) {
		return
	}

	normalizeCategorySlugUniqueness(db)
	// 项目经历：旧自由文本 category → type=project 分类
	backfillTextCategories(db, "project", "projects", "category", nil)
	// 技术栈：旧约定值/自由文本 category → type=tech_stack 分类
	backfillTextCategories(db, "tech_stack", "tech_stack_items", "category", map[string]string{
		"language":  "语言",
		"framework": "框架",
		"tool":      "工具",
		"database":  "数据库",
		"other":     "其他",
	})
	// 健身训练：旧固定枚举 type → type=fitness 分类
	backfillTextCategories(db, "fitness", "fitness_records", "type", map[string]string{
		"strength": "力量",
		"cardio":   "有氧",
		"stretch":  "拉伸",
	})
}

// categoryBackfillNeeded 判断是否需要执行分类回填：
//  1. 任一相关表缺失（AutoMigrate 异常等极端情况）→ 不执行，避免无谓报错；
//  2. 三张内容表均无「category_id 为空且 legacy 值非空」的待回填记录，
//     且全局 slug 唯一约束/索引已移除、联合唯一索引已就位 → 无需执行。
func categoryBackfillNeeded(db *gorm.DB) bool {
	for _, table := range []string{"categories", "projects", "tech_stack_items", "fitness_records"} {
		if !db.Migrator().HasTable(table) {
			return false
		}
	}

	var need bool
	if err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM projects
			WHERE category_id IS NULL AND category IS NOT NULL AND category <> ''
		) OR EXISTS (
			SELECT 1 FROM tech_stack_items
			WHERE category_id IS NULL AND category IS NOT NULL AND category <> ''
		) OR EXISTS (
			SELECT 1 FROM fitness_records
			WHERE category_id IS NULL AND type IS NOT NULL AND type <> ''
		) OR EXISTS (
			SELECT 1 FROM pg_constraint
			WHERE conrelid = 'categories'::regclass AND conname = 'categories_slug_key'
		) OR EXISTS (
			SELECT 1 FROM pg_indexes
			WHERE schemaname = current_schema() AND tablename = 'categories' AND indexname = 'idx_categories_slug'
		) OR NOT EXISTS (
			SELECT 1 FROM pg_indexes
			WHERE schemaname = current_schema() AND tablename = 'categories' AND indexname = 'uq_categories_type_slug'
		)`).Scan(&need).Error; err != nil {
		// 探测失败按无需回填处理，不影响启动
		return false
	}
	return need
}

// normalizeCategorySlugUniqueness 把 categories.slug 的全局唯一（旧 SQL 迁移的列约束或
// GORM 旧模型的单列索引）改为 (type, slug) 联合唯一，允许跨模块使用同名分类。
func normalizeCategorySlugUniqueness(db *gorm.DB) {
	_ = db.Exec(`ALTER TABLE categories DROP CONSTRAINT IF EXISTS categories_slug_key`).Error
	_ = db.Exec(`DROP INDEX IF EXISTS idx_categories_slug`).Error
	_ = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uq_categories_type_slug ON categories(type, slug)`).Error
}

// backfillTextCategories 把 legacy 自由文本列的每个未回填取值播种为 type=categoryType 的分类
// （known 中的取值使用映射中文名，slug 一律保留原值），再把对应记录的 category_id 回填为分类 ID。
// table/column 均为编译期常量，仅用于拼接 SQL，不涉及用户输入。
func backfillTextCategories(db *gorm.DB, categoryType, table, column string, known map[string]string) {
	// 仅取尚无 category_id 的记录涉及的取值：避免给已被管理端挂好分类的镜像值重复播种
	var values []string
	if err := db.Raw(
		fmt.Sprintf("SELECT DISTINCT %s FROM %s WHERE %s IS NOT NULL AND %s <> '' AND category_id IS NULL", column, table, column, column),
	).Scan(&values).Error; err != nil || len(values) == 0 {
		return
	}

	for _, value := range values {
		var existing model.Category
		if err := db.Where("type = ? AND (slug = ? OR name = ?)", categoryType, value, value).First(&existing).Error; err == nil {
			continue
		}
		name := value
		if n, ok := known[value]; ok {
			name = n
		}
		cat := model.Category{
			ID:   uuid.New().String(),
			Name: name,
			Slug: value,
			Type: categoryType,
		}
		if err := db.Create(&cat).Error; err != nil {
			continue
		}
	}

	// 按 slug 或 name 匹配回填；category_id IS NULL 条件保证重复执行天然跳过
	_ = db.Exec(fmt.Sprintf(
		"UPDATE %s t SET category_id = c.id FROM categories c WHERE c.type = ? AND (c.slug = t.%s OR c.name = t.%s) AND t.%s IS NOT NULL AND t.%s <> '' AND t.category_id IS NULL",
		table, column, column, column, column,
	), categoryType).Error
}
