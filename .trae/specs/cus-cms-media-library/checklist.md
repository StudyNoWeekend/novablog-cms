# 验收检查清单

- [ ] 后端 `MediaListReq` 已定义 `file_type` 与 `keyword` 字段，并正确内嵌 `PageReq`
- [ ] 后端 `MediaModel.GetList` 已按 `file_type` 过滤并按 `filename` 模糊搜索
- [ ] 后端 `MediaController.GetList` 已替换为 `MediaListReq` 参数绑定
- [ ] 前端 `api/media.ts` 的 `getList` 已支持 `file_type`、`keyword`、`page`、`page_size` 参数
- [ ] 前端媒体库页面已实现「图片 / 视频」Tab 切换并触发对应请求
- [ ] 前端媒体库页面已实现搜索框，输入关键字后可按文件名搜索
- [ ] 前端媒体库页面已实现「图标视图 / 文件视图」切换
- [ ] 图标视图下，图片显示预览图，视频显示视频占位/缩略图
- [ ] 文件视图下，表格展示文件名、类型、尺寸、上传时间等信息
- [ ] 前端媒体库页面已实现分页，切换页码或每页条数后正确刷新
- [ ] 上传按钮仅作占位，不实现实际功能
- [ ] `MediaPicker.vue` 仍可正常选择图片且不受新接口影响
- [ ] 后端 `go build ./...` 通过
- [ ] 前端 `npm run build` 通过
- [ ] 前后端联调通过，分类筛选、搜索、视图切换、分页均正常