# Tasks

- [x] Task 1: 数据层 — 新增 storage_config / media.storage_type 迁移与模型
  - [x] SubTask 1.1: 编写 migration `003_storage_config.up/down.sql`，建 storage_config 表（provider 唯一约束、is_active 全局唯一部分索引）、storage_migration_task、storage_migration_item 表
  - [x] SubTask 1.2: 编写 migration 给 media 表加 `storage_type varchar(20)` 列，存量数据回填 `local`
  - [x] SubTask 1.3: 新增 `internal/model/storage_config.go`（StorageConfig 模型 + CRUD：GetAll/GetByProvider/GetActive/Upsert/Delete/SetActive）
  - [x] SubTask 1.4: 新增 `internal/model/storage_migration.go`（MigrationTask/MigrationItem 模型 + Create/Update/GetByID/GetItems）

- [x] Task 2: 存储抽象层 — Provider 接口与三平台实现
  - [x] SubTask 2.1: 定义 `internal/storage/provider.go`：StorageProvider 接口（Upload/Download/Delete/Exists/Type）
  - [x] SubTask 2.2: 实现 `internal/storage/aliyun.go`（aliyun-oss-go-sdk，V4 签名，bucket.PutObject / GetObject / DeleteObject / IsObjectExist）
  - [x] SubTask 2.3: 实现 `internal/storage/tencent.go`（cos-go-sdk-v5，BaseURL 由 bucket+region 拼接，Object.Put/Get/Delete/IsExist）
  - [x] SubTask 2.4: 实现 `internal/storage/minio.go`（minio-go/v7，PutObject/GetObject/RemoveObject/StatObject，useSSL 取自 extra）
  - [x] SubTask 2.5: 实现 `internal/storage/factory.go`：按 provider + config 构建 Provider（解析 extra JSON）

- [x] Task 3: StorageManager 热重载
  - [x] SubTask 3.1: 实现 `internal/storage/manager.go`：atomic.Pointer[activeBundle]，GetProvider 无锁读，Reload(ctx) 从 DB 读 active 配置解密后构建 Provider 原子替换
  - [x] SubTask 3.2: 实现 TestProvider(config) 临时构建并校验连通（不落库）
  - [x] SubTask 3.3: 实现 AES-GCM 加解密工具 `utils/crypto/aes.go`，密钥来自 config.crypto.secret_key

- [x] Task 4: 配置加解密与 bootstrap 集成
  - [x] SubTask 4.1: config.yaml 新增 `crypto.secret_key`，bootstrap 读取注入
  - [x] SubTask 4.2: main.go 启动时构建 StorageManager 并调用 Reload，注入 router

- [x] Task 5: 存储配置 controller/router/dto
  - [x] SubTask 5.1: req/res DTO（StorageConfigReq/TestReq/ActivateReq，返回 DTO 脱敏 access_secret）
  - [x] SubTask 5.2: storage_logic.go：CRUD + activate（触发 manager.Reload）+ test
  - [x] SubTask 5.3: storage_controller.go + router 注册（GET/POST/PUT/DELETE /config，POST /config/:provider/activate，POST /config/test）

- [x] Task 6: 媒体上传链路改造
  - [x] SubTask 6.1: MediaLogic 改为接收 *storage.Manager，UploadFile 调用 manager.GetProvider().Upload
  - [x] SubTask 6.2: MediaController / RegisterRoutes 签名变更，移除 uploadDir/baseURL 透传
  - [x] SubTask 6.3: 对象 key 沿用 `images/2006/01/uuid.ext`，与 path_prefix 拼接；url 优先 custom_domain
  - [x] SubTask 6.4: 编译验证 + 现有媒体列表/删除接口回归

- [x] Task 7: 迁移引擎 — 分析
  - [x] SubTask 7.1: storage_migration_logic.go：Analyze(target_provider) 异步任务，遍历 media 调目标 Provider.Exists
  - [x] SubTask 7.2: 返回缺失/已存在清单（分页），存入 migration_task（type=analyze）

- [x] Task 8: 迁移引擎 — 执行
  - [x] SubTask 8.1: StartMigration(task_id, media_ids/all) 创建 type=migrate 任务，goroutine 执行
  - [x] SubTask 8.2: 拉源（local 走本地磁盘，云平台构建源 Provider）→ 上传目标 → Exists 验证 → 更新 media.url/storage_type → 累加进度
  - [x] SubTask 8.3: 保留源文件不删除；失败记入 migration_item，支持单条重试
  - [x] SubTask 8.4: 迁移 running 期间 activate 被拒（查询 running 任务数）

- [x] Task 9: 迁移 controller/router
  - [x] SubTask 9.1: POST /storage/migration/analyze、GET /storage/migration/analyze/:taskId
  - [x] SubTask 9.2: POST /storage/migration/start、GET /storage/migration/status/:taskId
  - [x] SubTask 9.3: 全部接口接入 authMiddleware

- [x] Task 10: 联调与验证
  - [x] SubTask 10.1: go build 通过，go vet 无警告
  - [x] SubTask 10.2: 代码层面验证：配置 MinIO 激活后上传链路完整（需实际运行环境验证端到端）
  - [x] SubTask 10.3: 代码层面验证：切换平台后 analyze 显示缺失，执行迁移进度更新，迁移后 media.storage_type 更新

# Task Dependencies
- Task 2 依赖 Task 1（模型）
- Task 3 依赖 Task 2（Provider）
- Task 5 依赖 Task 3（Manager）
- Task 6 依赖 Task 3
- Task 7/8 依赖 Task 6（上传链路）与 Task 3
- Task 9 依赖 Task 7/8
- Task 4 可与 Task 2/3 并行（config 改动独立）
- Task 10 依赖全部
