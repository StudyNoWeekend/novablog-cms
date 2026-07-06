# 媒体库图片预设工作台规格说明

## Why
当前媒体库仅支持上传原图并展示。摄影场景下，用户需要对同一张原图应用不同的相框样式与 EXIF 展示参数，产出多张成品图。每次调整都应基于原图重新合成，避免画质损失。本规格引入「预设」概念：上传原图后可在工作台中调整相框样式与 EXIF 展示参数，保存时生成新的成品图上传到对象存储，原图始终保持不变。

## 实现方案选型
- **图像处理（EXIF 读取 + 相框合成）由前端完成**：工作台需要实时预览，Canvas 即时渲染交互体验最佳。
- **后端只负责接收成品图 Blob 并转存对象存储 + 记录预设元数据**。
- **EXIF 读取**：使用 `exifr`。
- **EXIF 写入**：使用 `piexifjs`，仅 JPEG 支持；PNG 成品图不写 EXIF。
- **原图获取**：由后端提供代理下载接口 `GET /api/v1/media/:id/original`，避免依赖对象存储 CORS。

## What Changes
- 新增数据表 `media_preset`，记录原图与预设的关联、显示参数、相框配置、成品图存储信息。
- 后端新增预设 CRUD 接口与原图代理下载接口。
- 前端新增 `MediaPresetWorkbench.vue` 工作台组件，采用三栏布局：左侧样式/参数调整、中间预览+预设栏、右侧只读 EXIF/预设库/导出。
- 前端媒体库图片卡片新增「预设」入口与预设数量徽标。
- 前端新增 `exifr`、`piexifjs` 依赖。

## Impact
- 新增后端文件：
  - `internal/model/media_preset.go`
  - `internal/dto/req/media_req.go` 增加 `CreatePresetReq`
  - `internal/dto/res/media_res.go` 增加 `MediaPresetRes`
  - `internal/logic/media_logic.go` 增加预设方法
  - `internal/controller/media_controller.go` 增加预设 handler
  - `internal/router/media.go` 增加路由
  - `migrations/004_media_preset.up.sql` / `.down.sql`
- 新增前端文件：
  - `cus-cms-web/src/components/media/MediaPresetWorkbench.vue`
- 受影响前端文件：
  - `cus-cms-web/src/views/media/MediaLibraryView.vue`
  - `cus-cms-web/src/api/media.ts`
  - `cus-cms-web/src/types/api.ts`
  - `cus-cms-web/package.json`

## ADDED Requirements

### Requirement: 预设数据模型
系统 SHALL 新增 `media_preset` 表，存储每个预设的名称、相框配置（JSON）、显示参数（JSON）、成品图 URL 与存储路径，并通过 `media_id` 关联原图。原图记录不可变。

#### Scenario: 预设与原图的关系
- **GIVEN** 一张已上传的原图（media 记录）
- **WHEN** 用户为该图创建预设
- **THEN** 系统在 `media_preset` 表插入一条记录，`media_id` 指向原图
- **AND** 原图记录的 URL、StoragePath 等字段保持不变

### Requirement: 原图代理下载
后端 SHALL 提供 `GET /api/v1/media/:id/original` 接口，根据 media 记录的存储信息从对象存储拉取原图并以 `image/*` 内容类型返回，供前端工作台加载原图 blob。

#### Scenario: 前端加载原图
- **WHEN** 前端打开预设工作台
- **THEN** 前端请求 `GET /api/v1/media/:id/original`
- **AND** 后端返回原图二进制流
- **AND** 响应头设置 `Cache-Control` 便于浏览器缓存

### Requirement: 创建预设接口
后端 SHALL 提供 `POST /api/v1/media/preset` 接口，接收 multipart 表单（成品图文件 + 预设名称 + media_id + frame_config JSON + display_params JSON），将成品图转存对象存储并记录预设。

#### Scenario: 保存预设
- **WHEN** 前端在工作台点击保存
- **THEN** 前端合成成品图 Blob，连同参数上传到 `POST /api/v1/media/preset`
- **AND** 后端将成品图上传到对象存储 `presets/{media_id}/{uuid}.jpg`
- **AND** 后端在 `media_preset` 表插入记录
- **AND** 返回新建预设信息（含成品图 URL）

### Requirement: 预设列表接口
后端 SHALL 提供 `GET /api/v1/media/:id/presets` 接口，返回指定原图下的所有预设。

#### Scenario: 查询预设列表
- **WHEN** 前端请求某原图预设列表
- **THEN** 后端返回该原图下所有未删除的预设，按创建时间倒序

### Requirement: 删除预设接口
后端 SHALL 提供 `DELETE /api/v1/media/preset/:id` 接口，软删除指定预设。

#### Scenario: 删除预设
- **WHEN** 用户删除某预设
- **THEN** 前端请求 `DELETE /api/v1/media/preset/:id`
- **AND** 后端软删除该预设记录

### Requirement: 工作台布局
前端 SHALL 采用三栏布局实现 `MediaPresetWorkbench.vue`：
- 左侧：模板风格、样式参数、品牌/型号展示模式、EXIF 显示参数编辑。
- 中间：Canvas 实时预览区 + 底部预设缩略图栏（该原图已有的预设成品图）。
- 右侧：只读 EXIF 信息、预设库（名称输入+保存）、导出（文件名+保存按钮）。

### Requirement: 左侧模板风格
左侧 SHALL 提供「画廊 / 电影 / 悬浮」三种模板风格切换，不同风格对应不同的默认相框配置。切换时中间 Canvas 实时更新预览。

### Requirement: 左侧样式参数
左侧 SHALL 提供以下样式参数调整，并实时同步到 Canvas：
- 字体大小倍数（Slider，默认 1.0x）
- 边框尺寸倍数（Slider，默认 1.0x）
- 边框背景色（白色 / 黑色，或颜色选择器）
- 文字颜色（自动 / 黑色 / 白色）
- 字体选择（默认 Inter）

### Requirement: 品牌/型号展示模式
左侧 SHALL 提供「文字 / 图片」两种品牌展示模式：
- 文字模式：在相框信息条中直接展示品牌/型号字符串。
- 图片模式：用户可上传品牌 Logo 图片，相框信息条中展示该 Logo（Logo 走独立的图片上传接口，先占位，可后续扩展）。

### Requirement: 左侧 EXIF 显示参数编辑
左侧 SHALL 提供可编辑的 EXIF 显示参数字段，用于控制中间相框信息条展示内容：
- 品牌、型号、镜头、焦距、光圈、快门、ISO、日期。
- 打开工作台时，字段默认从原图 EXIF 中自动填充（JPEG 使用 exifr 读取，PNG 可留空）。

### Requirement: 右侧只读 EXIF 信息
右侧 SHALL 展示原图的真实 EXIF 信息，仅只读，不可编辑：
- 文件、分辨率、大小、厂商、型号、镜头、软件、FOCAL、AP、SHTR、ISO 等卡片。

### Requirement: 中间 Canvas 实时预览
中间 SHALL 使用 Canvas 实时绘制原图 + 相框 + 信息条，响应左侧所有参数变化，不产生网络请求。

#### Scenario: 实时预览
- **WHEN** 用户调整模板风格、样式参数或 EXIF 显示参数
- **THEN** Canvas 立即重绘预览

### Requirement: 底部预设缩略图栏
中间底部 SHALL 展示当前原图已有的所有预设成品图缩略图（通过 `GET /api/v1/media/:id/presets` 获取），支持点击切换选中、删除、新增预设入口。

### Requirement: 右侧预设库
右侧 SHALL 提供预设名称输入框与保存按钮。点击保存时，前端按当前参数合成成品图并调用 `POST /api/v1/media/preset` 创建预设。

### Requirement: 右侧导出
右侧 SHALL 提供文件名输入框、格式后缀（.JPG）与保存按钮。保存按钮行为与预设库保存一致：生成成品图并上传，成功后给出提示。BATCH 按钮先占位。

### Requirement: 保存合成
前端 SHALL 在用户点击保存时，按原图尺寸用 Canvas 合成最终图（原图 + 相框 + 信息条），对 JPEG 用 `piexifjs` 写入左侧调整后的 EXIF 显示参数，生成 Blob 上传后端。

#### Scenario: 大图处理
- **WHEN** 原图单边超过 4096px
- **THEN** 预览 Canvas 使用缩放尺寸，保存时仍按原图尺寸合成
- **AND** 若合成失败，提示用户

## MODIFIED Requirements
无。本规格全部为新增能力。

## REMOVED Requirements
无。

## 约束与注意事项
- 原图始终不可变。
- 预设成品图统一导出为 JPEG（即使原图为 PNG，也按 JPEG 保存，便于 EXIF 写入；若原图为 PNG 且无 EXIF，则跳过 EXIF 写入）。
- 前端 Canvas 加载原图需通过后端代理，避免跨域污染。
- 成品图文件大小限制沿用现有上传限制（100MB）。
