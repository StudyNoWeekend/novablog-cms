# 验证清单

## 数据层
- [x] storage_config 表创建，provider 唯一约束、is_active 全局唯一（部分索引保证仅一个 true）
- [x] storage_migration_task / storage_migration_item 表创建
- [x] media 表新增 storage_type 字段，存量数据回填 `local`
- [x] StorageConfig 模型 CRUD 方法齐全（GetAll/GetByProvider/GetActive/Upsert/Delete/SetActive）
- [x] MigrationTask/MigrationItem 模型方法齐全（Create/Update/GetByID/GetItems）

## 存储抽象层
- [x] StorageProvider 接口定义：Upload/Download/Delete/Exists/Type
- [x] 阿里云 OSS 实现：endpoint 含 https、region 必填、V4 签名、bucket 纯桶名
- [x] 腾讯云 COS 实现：bucket 名带 APPID、由 bucket+region 拼接访问 URL
- [x] MinIO 实现：endpoint 不含 scheme、useSSL 取自 extra JSON
- [x] factory 按 provider + config（含 extra 解析）构建对应 Provider

## 热重载与安全
- [x] StorageManager 用 atomic.Pointer 持有活跃 Provider，GetProvider 无锁读
- [x] Reload 从 DB 读 active 配置 → 解密 access_secret → 构建 Provider → 原子替换
- [x] access_secret 使用 AES-GCM 加密落库，密钥来自 config.crypto.secret_key
- [x] 配置列表接口 access_secret 返回 `******` 脱敏
- [x] test 接口不落库，用临时配置校验 BucketExists
- [x] 未配置存储时上传返回「对象存储未配置」错误

## 配置激活流程
- [x] activate 接口：先将所有 active 置 false，再置目标 true，触发 manager.Reload
- [x] 激活后上传立即走新 Provider，无需重启
- [x] 迁移任务 running 期间 activate 被拒，返回明确错误

## 媒体上传改造
- [x] MediaLogic 接收 *storage.Manager，UploadFile 调用 Provider.Upload
- [x] 对象 key 沿用 `images/2006/01/uuid.ext`，与 path_prefix 拼接
- [x] url 优先使用 custom_domain，否则平台默认 URL
- [x] media.storage_type 写入当前 Provider 类型
- [x] RegisterRoutes 签名变更：移除 uploadDir/baseURL，新增 storageMgr
- [x] main.go 启动构建 StorageManager 并 Reload
- [x] 文件大小限制 100MB 沿用
- [x] go build 通过，go vet 无警告

## 迁移分析
- [x] analyze 接口异步任务，遍历 media 记录
- [x] storage_type == target_provider 视为已存在跳过
- [x] 否则调目标 Provider.Exists 校验
- [x] 返回缺失/已存在清单（id/filename/storage_path/storage_type），分页可查

## 迁移执行
- [x] start 接口支持 media_ids[] 批量 与 all=true 一键两种模式
- [x] 创建 migration_task（status=pending），返回 task_id
- [x] 异步 goroutine：拉源（local 走本地磁盘，云平台构建源 Provider）→ 上传目标 → Exists 验证 → 更新 media.url/storage_type
- [x] 迁移成功并验证后保留源文件（不删除）
- [x] 失败记入 migration_item，支持单条重试
- [x] status 接口返回 total/succeeded/failed/started_at/finished_at/error 及失败明细

## 路由与鉴权
- [x] 存储配置路由：GET/POST/PUT/DELETE /api/v1/storage/config，POST /config/:provider/activate，POST /config/test
- [x] 迁移路由：POST /storage/migration/analyze、GET /analyze/:taskId、POST /start、GET /status/:taskId
- [x] 全部新接口接入 authMiddleware

## 联调验证
- [x] go build 通过，go vet 无警告
- [ ] 启动服务，配置 MinIO（本地 docker）激活后上传图片成功（需运行环境）
- [ ] 切换平台后 analyze 显示缺失清单（需运行环境）
- [ ] 执行迁移，进度正确更新，迁移后 media.storage_type 更新为目标平台（需运行环境）
- [x] 迁移成功后源文件仍保留（代码层面已验证：migrateOne 中无删除源文件操作）
