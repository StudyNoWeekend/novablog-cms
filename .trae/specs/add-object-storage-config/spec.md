# 对象存储多平台配置与管理 Spec

## Why
当前系统媒体上传仅支持本地磁盘（[media_logic.go](file:///Users/apple/GolandProjects/ai-cus/cus-cms/internal/logic/media_logic.go) 中硬编码 `os.Create`），无法接入云对象存储，且配置写死在 `config.yaml`，切换平台需重启。需要支持阿里云 OSS、腾讯云 COS、MinIO 三种平台的动态配置、热重载上传、以及平台更换时的素材分析与迁移。

## What Changes
- 新增 `storage_config` 数据表，持久化多平台对象存储配置，敏感字段 AES 加密落库
- 新增 `storage_migration_task` / `storage_migration_item` 数据表，记录迁移任务与明细
- **BREAKING** `media` 表新增 `storage_type` 字段，记录每个素材实际所在平台
- 新增 `StorageProvider` 接口及阿里云/腾讯云/MinIO 三个实现
- 新增 `StorageManager`，使用 `atomic.Pointer` 持有当前活跃 Provider，支持配置变更后不重启热重载
- 新增迁移引擎：分析目标平台缺失素材 + 异步迁移执行 + 进度轮询
- 新增存储配置与迁移的 controller/router/dto
- **BREAKING** `MediaLogic` / `MediaController` / `RegisterRoutes` 改为接收 `*StorageManager`，移除 `uploadDir`/`baseURL` 透传
- 本地磁盘降级为「迁移只读源」（用于把存量本地文件一次性搬到首个云平台），不再作为可配置存储类型
- `config.yaml` 新增 `crypto.secret_key` 用于密钥加解密

## Impact
- Affected specs: cus-cms-media-library（媒体库上传链路变更）
- Affected code:
  - [media_logic.go](file:///Users/apple/GolandProjects/ai-cus/cus-cms/internal/logic/media_logic.go)（上传逻辑替换为 Provider 调用）
  - [media_controller.go](file:///Users/apple/GolandProjects/ai-cus/cus-cms/internal/controller/media_controller.go)（构造函数签名变更）
  - [router.go](file:///Users/apple/GolandProjects/ai-cus/cus-cms/internal/router/router.go)（透传 StorageManager）
  - [app.go](file:///Users/apple/GolandProjects/ai-cus/cus-cms/bootstrap/app.go)（构建 StorageManager）
  - [main.go](file:///Users/apple/GolandProjects/ai-cus/cus-cms/cmd/api/main.go)（启动时 Reload）
  - [config.yaml](file:///Users/apple/GolandProjects/ai-cus/cus-cms/config/config.yaml)（新增 crypto 段、保留 upload.dir 作为迁移源）
  - [go.mod](file:///Users/apple/GolandProjects/ai-cus/cus-cms/go.mod)（新增三个 SDK 依赖）

## 平台参数差异对照

各平台填写参数各不相同，统一抽象为 `endpoint / region / bucket / access_key / access_secret / extra(JSON)`：

| 参数字段 | 阿里云 OSS | 腾讯云 COS | MinIO |
|---------|-----------|-----------|-------|
| `endpoint` | `https://oss-cn-hangzhou.aliyuncs.com`（region 级端点，含 https） | `https://cos.ap-guangzhou.myqcloud.com`（region 服务端点，bucket 在 URL 路径） | `play.min.io` 或 `localhost:9000`（**不含 scheme**） |
| `region` | `cn-hangzhou`（V4 签名必填） | `ap-guangzhou` | `us-east-1`（可选，可空） |
| `bucket` | `mybucket`（纯桶名） | `examplebucket-1250000000`（**必须带 APPID**） | `mybucket`（纯桶名） |
| `access_key` | AccessKeyId | SecretId | accessKey |
| `access_secret` | AccessKeySecret（加密落库） | SecretKey（加密落库） | secretKey（加密落库） |
| `extra` (JSON) | `{"sign_version":"v4"}` | `{}` | `{"use_ssl":true}` |

SDK 选型：
- 阿里云：`github.com/aliyun/aliyun-oss-go-sdk/oss`（V1，稳定；V4 签名）
- 腾讯云：`github.com/tencentyun/cos-go-sdk-v5`
- MinIO：`github.com/minio/minio-go/v7`

腾讯云访问 URL 由 `bucket + endpoint` 拼接：`https://{bucket}.cos.{region}.myqcloud.com`，其中 bucket 字段本身已含 appid，故只需存 bucket 与 region。

## ADDED Requirements

### Requirement: 多平台存储配置管理
系统 SHALL 支持在后台配置阿里云 OSS、腾讯云 COS、MinIO 三种对象存储，每种平台保留一行配置（provider 唯一），全局仅一个 `is_active=true`。

#### Scenario: 新增并激活某平台配置
- **WHEN** 管理员提交某平台配置（含 endpoint/region/bucket/access_key/access_secret）
- **THEN** 系统校验参数完整性，access_secret AES 加密后落库，is_active 默认 false
- **WHEN** 管理员调用 activate 接口
- **THEN** 系统先将该平台所有 active 置 false，再将目标置 true，触发 StorageManager.Reload 热重载，上传立即走新 Provider

#### Scenario: 查询配置列表脱敏
- **WHEN** 管理员查询配置列表
- **THEN** access_secret 返回 `******` 脱敏串，不泄露明文

### Requirement: 配置连通性测试
系统 SHALL 提供配置连通性测试接口，不落库即测，激活前必须通过测试。

#### Scenario: 测试未保存的配置
- **WHEN** 管理员提交临时配置调用 test 接口
- **THEN** 系统用该配置构建临时 Provider，执行 `BucketExists` 校验，返回连通成功/失败及错误详情

### Requirement: 存储抽象层与热重载
系统 SHALL 通过 `StorageProvider` 接口抽象上传/下载/删除/存在性判断，`StorageManager` 用 `atomic.Pointer` 持有当前活跃 Provider，读路径无锁。

#### Scenario: 未配置存储时上传
- **WHEN** 管理员上传文件但无 active 配置
- **THEN** 返回「对象存储未配置」错误，不落库

#### Scenario: 配置变更后不重启上传
- **WHEN** 配置被激活且 Reload 成功
- **THEN** 后续上传请求立即走新 Provider，无需重启进程

### Requirement: 素材迁移分析
系统 SHALL 支持分析当前所有素材在目标平台中的存在情况，输出「已存在 / 缺失」清单。

#### Scenario: 分析目标平台缺失素材
- **WHEN** 管理员发起 analyze（指定 target_provider）
- **THEN** 系统遍历 media 记录：
  - 若记录 `storage_type == target_provider` → 已存在（跳过）
  - 否则用目标 Provider 调 `Exists(storage_path)` 校验
- **AND** 返回缺失清单（id/filename/storage_path/storage_type）与已存在清单，分页可查

### Requirement: 异步迁移执行与进度
系统 SHALL 异步执行素材迁移，支持批量（指定 ID）和一键（全部缺失）两种模式，提供进度轮询。

#### Scenario: 启动迁移任务
- **WHEN** 管理员发起 migration/start（body: media_ids[] 或 all=true，target_provider）
- **THEN** 系统创建 migration_task 记录，status=pending，返回 task_id
- **AND** 异步 goroutine 拉源文件（按记录 storage_type：local 走本地磁盘读，云平台用对应配置构建源 Provider）→ 上传目标 → 验证 `Exists` → 更新 `media.url` 与 `media.storage_type` → 累加 succeeded/failed
- **AND** 迁移成功并验证后**保留源文件**（只读副本，不自动删除）

#### Scenario: 轮询迁移进度
- **WHEN** 管理员查询 migration/status/:task_id
- **THEN** 返回 status/total/succeeded/failed/started_at/finished_at/error，以及失败明细列表

#### Scenario: 迁移中切换 active 被拒
- **WHEN** 有迁移任务 status=running 且尝试 activate 新配置
- **THEN** 返回「迁移进行中，禁止切换存储配置」错误

## MODIFIED Requirements

### Requirement: 媒体上传
原有 [media_logic.go](file:///Users/apple/GolandProjects/ai-cus/cus-cms/internal/logic/media_logic.go) 的 `UploadFile` 写本地磁盘，修改为：

- 通过 `StorageManager.GetProvider()` 获取当前活跃 Provider
- 对象 key 沿用现有结构 `images/2006/01/uuid.ext`（与 path_prefix 拼接：`{path_prefix}/images/2006/01/uuid.ext`）
- 调用 `Provider.Upload(ctx, key, reader, size, contentType)` 返回 url
- `media.storage_type` 写入当前 Provider 的 type
- `media.url` 使用 `custom_domain`（若配置）拼接，否则用平台默认访问 URL
- 文件大小限制沿用 100MB

### Requirement: 路由与依赖注入
[router.go](file:///Users/apple/GolandProjects/ai-cus/cus-cms/internal/router/router.go) 的 `RegisterRoutes` 签名变更：移除 `uploadDir, baseURL string` 参数，新增 `storageMgr *storage.Manager` 参数。新增存储配置与迁移路由组。

## REMOVED Requirements

### Requirement: 本地存储作为可配置存储类型
**Reason**: 用户选择仅保留云存储。
**Migration**: 存量本地文件通过迁移引擎一次性搬到首个云平台；本地磁盘（`upload.dir`）保留为只读迁移源，`config.yaml` 的 `upload` 段保留不删除，仅用于迁移读取。

## 安全要求
- `access_secret` 使用 AES-GCM 加密落库，密钥来自 `config.yaml` 的 `crypto.secret_key`
- 接口返回脱敏
- 激活前必须 test 通过，避免坏配置顶掉好配置
- 迁移任务运行期锁定 active 切换
- 文件大小限制 100MB 沿用
