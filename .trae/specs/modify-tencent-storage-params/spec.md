# 简化腾讯云存储配置参数 Spec

## Why
当前腾讯云 COS 配置要求填写 endpoint + bucket + access_key + access_secret + region 共 5 个字段，其中 endpoint 字段实际未被使用（被忽略），bucket 与 region 拼接出 URL。用户希望简化为只需 4 个字段：bucket_url、secret_id、secret_key、region。

## What Changes
- **MODIFIED** 腾讯云 COS 配置参数简化：用户只需输入 `bucket_url`（完整桶 URL）+ `secret_id` + `secret_key` + `region`
- **MODIFIED** 后端 `tencent.go`：`NewTencentProvider` 签名变更，接收 `bucketURL` 直接用作 SDK BucketURL，不再从 bucket+region 拼接
- **MODIFIED** 后端 `factory.go`：tencent 分支传 `config.Endpoint` 作为 bucket_url（复用 endpoint 字段存储 bucket_url）
- **MODIFIED** 后端 `storage_req.go`：`bucket` 字段不再 `binding:"required"`（腾讯云不需要单独的 bucket）
- **MODIFIED** 前端 `StorageConfigView.vue`：腾讯云表单展示差异化字段标签，隐藏 bucket 字段

## 参数映射对照

| 用户输入字段 | 存入 StorageConfig 字段 | 说明 |
|------------|----------------------|------|
| bucket_url | `endpoint` | 完整桶 URL，如 `https://examplebucket-1250000000.cos.ap-guangzhou.myqcloud.com` |
| secret_id | `access_key` | 腾讯云 SecretId |
| secret_key | `access_secret` | 腾讯云 SecretKey（加密落库） |
| region | `region` | 如 `ap-guangzhou`，用于 SDK ServiceURL |

阿里云和 MinIO 的字段不变。

## Impact
- Affected code:
  - [tencent.go](file:///Users/apple/GolandProjects/ai-cus/cus-cms/internal/storage/tencent.go)（NewTencentProvider 签名 + buildURL 逻辑）
  - [factory.go](file:///Users/apple/GolandProjects/ai-cus/cus-cms/internal/storage/factory.go)（tencent 分支参数传递）
  - [storage_req.go](file:///Users/apple/GolandProjects/ai-cus/cus-cms/internal/dto/req/storage_req.go)（bucket 去掉 required）
  - [StorageConfigView.vue](file:///Users/apple/GolandProjects/ai-cus/cus-cms/cus-cms-web/src/views/storage/StorageConfigView.vue)（表单标签 + 字段显示）

## MODIFIED Requirements

### Requirement: 腾讯云 COS Provider 构建
原 `NewTencentProvider(endpoint, region, bucket, secretID, secretKey, pathPrefix, customDomain)` 修改为 `NewTencentProvider(bucketURL, secretID, secretKey, region, pathPrefix, customDomain)`。

#### Scenario: 使用 bucket_url 构建 COS 客户端
- **WHEN** 构建 TencentProvider
- **THEN** 将 `bucketURL`（如 `https://examplebucket-1250000000.cos.ap-guangzhou.myqcloud.com`）直接解析为 `url.URL` 作为 SDK 的 `BaseURL.BucketURL`
- **AND** 同时设置 `BaseURL.ServiceURL` 为 `https://cos.{region}.myqcloud.com`（region 非空时）
- **AND** 若 bucketURL 无 scheme（如 `examplebucket-1250000000.cos.ap-guangzhou.myqcloud.com`），自动补 `https://`

#### Scenario: 构造对象访问 URL
- **WHEN** 调用 `buildURL(key)`
- **THEN** 若 `customDomain` 非空，返回 `{customDomain}/{key}`
- **AND** 否则返回 `{bucketURL}/{key}`（直接用 bucket_url，不再从 bucket+region 拼接）

### Requirement: 配置请求 DTO 校验
原 `StorageConfigReq` 和 `StorageTestReq` 中 `Bucket` 字段 `binding:"required"` 改为非必填。

#### Scenario: 腾讯云配置提交
- **WHEN** 提交腾讯云配置，`endpoint` 字段存 bucket_url（必填），`bucket` 字段为空
- **THEN** 后端接受请求，不再因 bucket 为空报校验错误
- **AND** `endpoint` 仍保持 `binding:"required"`（bucket_url 必填）

### Requirement: 前端表单差异化展示
腾讯云表单字段与阿里云/MinIO 不同。

#### Scenario: 选择腾讯云时的表单
- **WHEN** provider 选择 `tencent`
- **THEN** 表单显示以下字段（标签更名）：
  - "Bucket URL"（placeholder: `https://examplebucket-1250000000.cos.ap-guangzhou.myqcloud.com`，help: 完整桶访问地址）
  - "Secret ID"（placeholder: 请输入 SecretId）
  - "Secret Key"（placeholder: 请输入 SecretKey，编辑时 placeholder: 留空不修改）
  - "Region"（placeholder: 如 ap-guangzhou）
  - "路径前缀"（可选，不变）
  - "自定义域名"（可选，不变）
- **AND** 不显示 "Endpoint" 和 "Bucket" 两个独立字段
- **AND** 提交时 `endpoint` 字段值 = 表单中的 bucket_url，`access_key` = secret_id，`access_secret` = secret_key

#### Scenario: 阿里云/MinIO 表单不变
- **WHEN** provider 选择 `aliyun` 或 `minio`
- **THEN** 表单字段和标签保持原有设计不变
