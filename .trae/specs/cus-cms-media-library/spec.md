# 媒体库模块规格说明

## Why
当前 CMS 已具备基础的媒体上传、列表与删除接口，但前端媒体库页面仅有一个占位卡片，无法按图片/视频分类浏览、按文件名搜索，也不支持图标视图与文件视图切换。本规格补齐后端查询能力并重建前端媒体库页面，实现前后端联调，使用户能够高效管理和浏览媒体资源。

## What Changes
- 后端媒体列表接口新增查询参数：
  - `file_type`：按文件类型过滤，1=图片，2=视频
  - `keyword`：按文件名模糊搜索
- 后端 `MediaListReq` / `MediaLogic.GetList` / `MediaModel.GetList` 增加过滤与搜索逻辑
- 前端 `MediaLibraryView.vue` 按设计图实现完整页面：
  - 顶部 Tab 切换「图片 / 视频」
  - 工具栏：搜索框、展示类型切换（图标视图 / 文件视图）、上传按钮（暂缓实现，仅占位）
  - 内容区：图标视图使用网格卡片展示缩略图；文件视图使用表格展示文件名、类型、尺寸、上传时间等信息
  - 底部分页组件
- 前端 `api/media.ts` 的 `getList` 扩展查询参数
- 前端 `MediaPicker.vue` 兼容新接口，保持原有图片选择功能不变

## Impact
- 受影响接口：`GET /api/v1/media`
- 受影响后端文件：
  - `internal/dto/req/page_req.go`（新增 `MediaListReq`）
  - `internal/controller/media_controller.go`
  - `internal/logic/media_logic.go`
  - `internal/model/media.go`
- 受影响前端文件：
  - `cus-cms-web/src/views/media/MediaLibraryView.vue`
  - `cus-cms-web/src/api/media.ts`
  - `cus-cms-web/src/components/media/MediaPicker.vue`（可选兼容）

## ADDED Requirements

### Requirement: 媒体分类筛选
后端媒体列表接口 SHALL 支持按 `file_type` 过滤，仅返回指定类型的媒体文件。

#### Scenario: 请求图片列表
- **WHEN** 用户切换到「图片」Tab
- **THEN** 前端请求 `GET /api/v1/media?file_type=1`
- **AND** 后端仅返回 `file_type = 1` 的媒体记录

#### Scenario: 请求视频列表
- **WHEN** 用户切换到「视频」Tab
- **THEN** 前端请求 `GET /api/v1/media?file_type=2`
- **AND** 后端仅返回 `file_type = 2` 的媒体记录

### Requirement: 按文件名搜索
后端媒体列表接口 SHALL 支持按文件名关键字模糊搜索。

#### Scenario: 搜索媒体
- **WHEN** 用户在搜索框输入关键字并触发查询
- **THEN** 前端请求 `GET /api/v1/media?keyword={keyword}`
- **AND** 后端返回文件名包含该关键字的记录（不区分大小写）

### Requirement: 展示类型切换
前端媒体库页面 SHALL 支持「图标视图」与「文件视图」两种展示方式。

#### Scenario: 图标视图
- **WHEN** 用户选择图标视图
- **THEN** 页面以网格卡片形式展示媒体缩略图
- **AND** 图片显示预览图，视频显示视频占位/缩略图

#### Scenario: 文件视图
- **WHEN** 用户选择文件视图
- **THEN** 页面以表格形式展示文件名、类型、尺寸、上传时间等信息

### Requirement: 分页
媒体列表 SHALL 支持分页展示。

#### Scenario: 翻页
- **WHEN** 用户切换页码或每页条数
- **THEN** 前端请求对应分页参数并刷新列表
- **AND** 后端返回正确的 `list`、`total`、`page`、`page_size`、`total_pages`

## MODIFIED Requirements

### Requirement: 媒体列表接口
现有 `GET /api/v1/media` 接口在保持返回格式不变的前提下，增加 `file_type` 与 `keyword` 可选查询参数。

- `file_type` 为空时返回全部类型
- `keyword` 为空时不进行文件名过滤
- 分页参数仍使用 `page` 和 `page_size`
- 返回字段保持与现有 `MediaListRes` 一致

## REMOVED Requirements
无移除需求。上传功能本次暂缓，不新增上传相关实现。