# Tasks

- [ ] Task 1: 后端预设数据层与迁移
  - [ ] SubTask 1.1: 新增 `migrations/004_media_preset.up.sql` 与 `.down.sql`，创建 `media_preset` 表（id uuid PK, media_id uuid FK, name varchar, frame_config jsonb, display_params jsonb, output_url varchar, output_storage_path varchar, output_size bigint, mime_type varchar, created_at timestamptz, deleted_at timestamptz，索引 media_id + deleted_at）
  - [ ] SubTask 1.2: 新增 `internal/model/media_preset.go`，定义 `MediaPreset` 模型与 `MediaPresetModel`（Create / GetByMediaID / GetByID / SoftDelete）

- [ ] Task 2: 后端预设与原图接口
  - [ ] SubTask 2.1: 在 `internal/dto/req/media_req.go` 新增 `CreatePresetReq`（media_id、name、frame_config、display_params，成品图文件由 FormFile 接收）
  - [ ] SubTask 2.2: 在 `internal/dto/res/media_res.go` 新增 `MediaPresetRes` 与 `MediaPresetListRes`
  - [ ] SubTask 2.3: 在 `internal/logic/media_logic.go` 增加 `CreatePreset`（转存成品图到 `presets/{media_id}/{uuid}.jpg` 并写库）、`GetPresetsByMediaID`、`DeletePreset`、`GetOriginalReader`（从对象存储拉取原图）
  - [ ] SubTask 2.4: 在 `internal/controller/media_controller.go` 增加 `CreatePreset`、`GetPresets`、`DeletePreset`、`GetOriginal` handler；`GetOriginal` 以 `image/*` Content-Type 流式返回并设置 `Cache-Control`
  - [ ] SubTask 2.5: 在 `internal/router/media.go` 注册路由：`POST /media/preset`、`GET /media/:id/presets`、`DELETE /media/preset/:id`、`GET /media/:id/original`

- [ ] Task 3: 前端类型与接口
  - [ ] SubTask 3.1: 在 `cus-cms-web/src/types/api.ts` 新增 `MediaPreset`、`FrameConfig`、`DisplayParams`、`ExifInfo` 类型
  - [ ] SubTask 3.2: 在 `cus-cms-web/src/api/media.ts` 新增 `createPreset`、`getPresets`、`deletePreset`、`getOriginalUrl` 方法
  - [ ] SubTask 3.3: 安装依赖 `exifr`、`piexifjs` 及 `@types/node` 已有，补充 piexifjs 类型声明（如无可直接类型定义）

- [ ] Task 4: 前端预设工作台组件 `MediaPresetWorkbench.vue`
  - [ ] SubTask 4.1: 搭建三栏布局（左 280-320px、中自适应、右 280-320px），使用 Ant Design Vue 组件；组件接收 `mediaId` prop，打开时请求原图代理接口并创建 blob URL
  - [ ] SubTask 4.2: 用 `exifr` 读取原图 EXIF，初始化右侧只读 EXIF 面板（文件、分辨率、大小、厂商、型号、镜头、软件、FOCAL/AP/SHTR/ISO 卡片）与左侧可编辑显示参数
  - [ ] SubTask 4.3: 左侧「模板风格」：画廊 / 电影 / 悬浮三种模板切换，每种模板对应一组默认 `FrameConfig`；切换时实时更新预览
  - [ ] SubTask 4.4: 左侧「样式参数」：字体大小倍数 Slider、边框尺寸倍数 Slider、边框背景色选择、文字颜色选择（自动/黑/白）、字体选择（默认 Inter）；所有变化实时同步到 Canvas
  - [ ] SubTask 4.5: 左侧「品牌/型号展示模式」：文字 / 图片 切换；文字模式直接编辑字符串；图片模式先占位（预留 Logo 上传入口，可后续扩展）
  - [ ] SubTask 4.6: 左侧「参数设定」：可编辑品牌、型号、镜头、焦距、光圈、快门、ISO、日期；打开时自动从 EXIF 填充
  - [ ] SubTask 4.7: 中间 Canvas 实时绘制原图 + 相框 + 底部信息条；根据文字颜色自动/黑白切换；信息条展示品牌/型号/镜头/焦距/光圈/快门/ISO/日期
  - [ ] SubTask 4.8: 中间底部预设缩略图栏：请求 `getPresets` 展示该原图已有预设，支持选中、删除、新增入口
  - [ ] SubTask 4.9: 右侧「预设库」：预设名称输入框 + 保存按钮；点击保存时按原图尺寸 Canvas 合成，JPEG 用 piexifjs 写入左侧显示参数中的 EXIF 字段，生成 Blob 调用 `createPreset`
  - [ ] SubTask 4.10: 右侧「导出」：文件名输入框 + .JPG 后缀 + 保存按钮（行为同预设库保存）；BATCH 按钮占位
  - [ ] SubTask 4.11: 大图处理：预览 Canvas 单边最大 4096px，保存按原图尺寸；合成异常提示用户

- [ ] Task 5: 媒体库入口
  - [ ] SubTask 5.1: 在 `MediaLibraryView.vue` 图片卡片增加「预设」入口与预设数量徽标
  - [ ] SubTask 5.2: 点击「预设」入口打开 `MediaPresetWorkbench.vue`（可用 Modal 或抽屉）

- [ ] Task 6: 联调与验证
  - [ ] SubTask 6.1: 验证 JPEG 原图：打开工作台 → 修改相框样式与 EXIF 显示参数 → 保存预设 → 成品图在对象存储与数据库正确生成
  - [ ] SubTask 6.2: 验证 PNG 原图：仅相框合成可正常保存
  - [ ] SubTask 6.3: 验证预设列表、删除、底部缩略图栏刷新
  - [ ] SubTask 6.4: 执行 `go build ./...` 与前端 `npm run build`，确保无编译错误

# Task Dependencies
- Task 2 依赖 Task 1
- Task 4 依赖 Task 3
- Task 5 依赖 Task 4
- Task 6 依赖 Task 1-5
