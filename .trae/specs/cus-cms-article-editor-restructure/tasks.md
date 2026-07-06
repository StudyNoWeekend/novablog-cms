# Tasks

- [x] Task 1: 前端 ArticleForm 组件重构
  - [x] SubTask 1.1: 移除表单中的状态 radio group UI
  - [x] SubTask 1.2: 移除表单中的文章类型 radio group UI 及 article_type 字段
  - [x] SubTask 1.3: 将剩余控件（摘要、封面图、分类、标签、置顶、评论）调整为紧凑水平工具栏布局
  - [x] SubTask 1.4: 更新 ArticleFormData interface，移除 status 和 article_type

- [x] Task 2: 前端创建/编辑页面布局重构
  - [x] SubTask 2.1: ArticleCreateView.vue 移除右侧 editor-sidebar，在 editor-toolbar 下方添加居中的元信息栏（使用 ArticleForm）
  - [x] SubTask 2.2: ArticleEditView.vue 执行与创建页相同的布局重构
  - [x] SubTask 2.3: 调整 editor-body/editor-content 样式，使编辑器占满全宽

- [x] Task 3: 前端类型与 store 清理
  - [x] SubTask 3.1: types/article.ts 中移除 article_type 字段
  - [x] SubTask 3.2: 检查并清理 stores/article.ts、api/article.ts 中 article_type 的引用（如有）

- [x] Task 4: 后端移除文章类型字段
  - [x] SubTask 4.1: dto/req/article_req.go 中移除 ArticleType 字段
  - [x] SubTask 4.2: model/article.go 中移除 ArticleType 字段
  - [x] SubTask 4.3: logic/article_logic.go 中移除 article_type 的默认值处理与赋值逻辑
  - [x] SubTask 4.4: dto/res/article_res.go 中移除 ArticleType 字段
  - [x] SubTask 4.5: 更新 migrations/002_article_migration.up.sql，移除 article_type 列定义与注释

# Task Dependencies
- Task 2 依赖 Task 1（页面布局依赖组件重构后的接口）
- Task 3 可与 Task 1 并行
- Task 4 为纯后端任务，可与前端任务并行
