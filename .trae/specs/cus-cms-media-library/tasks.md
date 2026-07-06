# Tasks

- [x] Task 1: 扩展后端媒体列表查询参数
  - [x] SubTask 1.1: 在 `internal/dto/req` 中新增 `MediaListReq`，内嵌 `PageReq` 并增加 `file_type` 和 `keyword` 字段
  - [x] SubTask 1.2: 修改 `internal/model/media.go` 的 `GetList` 方法，支持按 `file_type` 过滤和按 `filename` 模糊搜索
  - [x] SubTask 1.3: 修改 `internal/logic/media_logic.go` 的 `GetList`，透传过滤条件给 model 层
  - [x] SubTask 1.4: 修改 `internal/controller/media_controller.go` 的 `GetList`，绑定 `MediaListReq` 并调用 logic

- [x] Task 2: 实现前端媒体库页面
  - [x] SubTask 2.1: 扩展 `cus-cms-web/src/api/media.ts` 的 `getList` 参数类型，增加 `file_type`、`keyword`、`page`、`page_size`
  - [x] SubTask 2.2: 在 `MediaLibraryView.vue` 中实现顶部「图片 / 视频」Tab 切换
  - [x] SubTask 2.3: 在 `MediaLibraryView.vue` 中实现工具栏：搜索框（带防抖）、展示类型切换按钮、上传占位按钮
  - [x] SubTask 2.4: 在 `MediaLibraryView.vue` 中实现图标视图网格展示（图片预览 + 视频占位）
  - [x] SubTask 2.5: 在 `MediaLibraryView.vue` 中实现文件视图表格展示（文件名、类型、尺寸、上传时间）
  - [x] SubTask 2.6: 在 `MediaLibraryView.vue` 中实现底部分页，支持翻页与每页条数切换

- [ ] Task 3: 联调与兼容性验证
  - [ ] SubTask 3.1: 确保 `MediaPicker.vue` 调用 `mediaApi.getList` 后仍能正常选择图片
  - [ ] SubTask 3.2: 启动前后端服务，验证分类筛选、搜索、视图切换、分页功能正常
  - [ ] SubTask 3.3: 执行 `go build ./...` 与前端 `npm run build`，确保无编译错误

# Task Dependencies
- Task 2 依赖 Task 1（前端需要后端接口支持新参数）
- SubTask 3.1 与 SubTask 3.2 依赖 Task 1 和 Task 2