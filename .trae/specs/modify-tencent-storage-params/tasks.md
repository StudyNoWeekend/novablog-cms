# Tasks

- [ ] Task 1: 后端 — tencent.go 签名与逻辑修改
  - [ ] SubTask 1.1: `NewTencentProvider` 签名改为 `(bucketURL, secretID, secretKey, region, pathPrefix, customDomain string)`，解析 bucketURL 为 url.URL 作为 BucketURL，设置 ServiceURL 为 `https://cos.{region}.myqcloud.com`，无 scheme 时自动补 https://
  - [ ] SubTask 1.2: `buildURL` 改为直接用 bucketURL（不再从 bucket+region 拼接），customDomain 优先逻辑不变

- [ ] Task 2: 后端 — factory.go 与 DTO 修改
  - [ ] SubTask 2.1: `factory.go` tencent 分支改为传 `config.Endpoint` 作为 bucketURL（不再传 config.Bucket）
  - [ ] SubTask 2.2: `storage_req.go` 中 `StorageConfigReq` 和 `StorageTestReq` 的 `Bucket` 字段去掉 `binding:"required"`

- [ ] Task 3: 前端 — StorageConfigView.vue 腾讯云表单差异化
  - [ ] SubTask 3.1: provider=tencent 时，Endpoint 字段标签改为 "Bucket URL"，placeholder 改为完整桶 URL 示例，隐藏 Bucket 字段
  - [ ] SubTask 3.2: Access Key 标签改为 "Secret ID"，Access Secret 标签改为 "Secret Key"
  - [ ] SubTask 3.3: 提交时 endpoint 字段值 = 表单 bucket_url 输入值

- [ ] Task 4: 编译验证
  - [ ] SubTask 4.1: `go build ./...` 通过
  - [ ] SubTask 4.2: `npm run build` 通过

# Task Dependencies
- Task 1 → Task 2（factory 调用 tencent 构造函数）
- Task 3 可与 Task 1/2 并行（前端独立）
- Task 4 依赖全部
