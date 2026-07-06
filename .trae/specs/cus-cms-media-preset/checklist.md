# 验收检查清单

## 后端
- [ ] `migrations/004_media_preset.up.sql` 已创建 `media_preset` 表，含所有约定字段与索引
- [ ] `migrations/004_media_preset.down.sql` 可正确回滚
- [ ] `internal/model/media_preset.go` 已定义 `MediaPreset` 模型与 `Create/GetByMediaID/GetByID/SoftDelete` 方法
- [ ] `CreatePresetReq` 已定义 media_id、name、frame_config、display_params 字段
- [ ] `MediaPresetRes` 与 `MediaPresetListRes` 已定义
- [ ] `MediaLogic.CreatePreset` 已将成品图上传到对象存储 `presets/{media_id}/{uuid}.jpg` 并写库
- [ ] `MediaLogic.GetOriginalReader` 已从对象存储拉取原图
- [ ] `MediaController.GetOriginal` 已以 `image/*` Content-Type 流式返回原图并设置 `Cache-Control`
- [ ] 路由已注册：`POST /media/preset`、`GET /media/:id/presets`、`DELETE /media/preset/:id`、`GET /media/:id/original`
- [ ] 原图记录在任何预设操作下均不被修改
- [ ] `go build ./...` 通过

## 前端
- [ ] 已安装 `exifr`、`piexifjs` 依赖
- [ ] `types/api.ts` 已定义 `MediaPreset`、`FrameConfig`、`DisplayParams`、`ExifInfo` 类型
- [ ] `api/media.ts` 已实现 `createPreset`、`getPresets`、`deletePreset`、`getOriginalUrl`
- [ ] `MediaPresetWorkbench.vue` 已通过后端代理接口加载原图为 blob
- [ ] 三栏布局已实现（左样式参数、中预览、右 EXIF/预设库/导出）
- [ ] 左侧已实现「画廊 / 电影 / 悬浮」模板风格切换
- [ ] 左侧样式参数（字体大小、边框尺寸、边框背景、文字颜色、字体）已可调整并实时预览
- [ ] 左侧品牌/型号展示模式（文字/图片）已实现，文字模式可编辑
- [ ] 左侧参数设定（品牌、型号、镜头、焦距、光圈、快门、ISO、日期）已可编辑并自动从 EXIF 填充
- [ ] 右侧只读 EXIF 面板已展示文件、分辨率、大小、厂商、型号、镜头、软件、FOCAL/AP/SHTR/ISO 卡片
- [ ] 中间 Canvas 已实时绘制原图 + 相框 + 信息条
- [ ] 中间底部预设缩略图栏已展示该原图已有预设并支持删除/新增
- [ ] 右侧预设库保存按钮已能合成成品图并创建预设
- [ ] 右侧导出保存按钮行为与预设库保存一致
- [ ] 保存时 JPEG 已用 piexifjs 写入 EXIF 显示参数
- [ ] 预览大图已缩放至单边 ≤4096px，保存按原图尺寸合成
- [ ] `MediaLibraryView.vue` 图片卡片已有「预设」入口与数量徽标
- [ ] Canvas 未被跨域污染，可正常 toBlob 导出
- [ ] `npm run build` 通过

## 联调
- [ ] JPEG 原图：创建预设（改相框 + EXIF 显示参数）→ 成品图正常且 EXIF 已更新
- [ ] PNG 原图：仅相框合成可正常保存
- [ ] 预设列表、删除、底部缩略图栏刷新正常
- [ ] 原图在各操作后保持不变
