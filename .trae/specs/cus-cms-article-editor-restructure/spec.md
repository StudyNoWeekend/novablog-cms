# 文章编辑器重构 Spec

## Why
当前文章创建/编辑页面将摘要、封面图、分类等元信息放在右侧边栏中，与主编辑器并排，导致编辑区域被挤压。同时"状态"和"文章类型"字段增加了不必要的复杂度——保存草稿/发布按钮已足以控制状态，而文章类型暂无实际业务用途。

## What Changes
- 前端创建/编辑页：移除表单中的"状态"选择器 UI，保留保存草稿/发布按钮逻辑
- 前后端：彻底移除"文章类型"（article_type）字段及相关代码
- 前端布局重构：将摘要、封面图、分类、标签、置顶、评论模块从右侧边栏移至编辑器顶部、标题下方的居中位置
- 右侧边栏移除，编辑器区域扩展为全宽

## Impact
- 受影响页面：ArticleCreateView.vue、ArticleEditView.vue
- 受影响组件：ArticleForm.vue
- 受影响前端类型：types/article.ts
- 受影响后端模块：dto/req/article_req.go、model/article.go、logic/article_logic.go、dto/res/article_res.go
- 不受影响：文章列表页状态筛选与操作（保留）

## ADDED Requirements
无新增功能需求。

## MODIFIED Requirements
### Requirement: 文章创建/编辑页布局
The system SHALL 在文章创建与编辑页面中将元信息（摘要、封面图、分类、标签、置顶、评论）展示在标题输入框与编辑器之间的居中区域。

#### Scenario: 创建文章页
- **WHEN** 用户打开创建文章页
- **THEN** 页面顶部为工具栏（返回、标题输入、编辑器切换、操作按钮）
- **AND** 工具栏下方居中显示紧凑元信息栏
- **AND** 元信息栏下方为占满剩余宽度的编辑器

### Requirement: 文章表单组件重构
The system SHALL 在 ArticleForm 组件中移除"状态"和"文章类型"表单控件，并将剩余控件以紧凑的水平工具栏形式排布。

#### Scenario: 元信息编辑
- **WHEN** 用户在元信息栏操作
- **THEN** 可编辑摘要（文本域）、封面图（点击上传/更换）、分类（下拉单选）、标签（下拉多选）、置顶开关、评论开关

## REMOVED Requirements
### Requirement: 文章类型字段
**Reason**: 文章类型（普通/攻略）暂无实际业务用途，增加冗余。
**Migration**: 从前后端代码及数据库迁移脚本中删除 article_type 字段。已运行迁移的环境需手动执行 `ALTER TABLE articles DROP COLUMN article_type;` 或添加新迁移文件。
