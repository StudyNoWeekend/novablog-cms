# 验证清单

## 后端 tencent.go
- [x] NewTencentProvider 签名改为 `(bucketURL, secretID, secretKey, region, pathPrefix, customDomain string)`
- [x] bucketURL 直接解析为 url.URL 作为 SDK BaseURL.BucketURL
- [x] region 非空时设置 BaseURL.ServiceURL 为 `https://cos.{region}.myqcloud.com`
- [x] bucketURL 无 scheme 时自动补 `https://`
- [x] buildURL 用 bucketURL 直接拼 `/key`，不再从 bucket+region 构造
- [x] customDomain 非空时优先使用 customDomain

## 后端 factory.go
- [x] tencent 分支传 `config.Endpoint` 作为 bucketURL（不再传 config.Bucket）
- [x] 阿里云/MinIO 分支不变

## 后端 storage_req.go
- [x] StorageConfigReq.Bucket 去掉 `binding:"required"`
- [x] StorageTestReq.Bucket 去掉 `binding:"required"`
- [x] StorageConfigReq.Endpoint 和 StorageTestReq.Endpoint 保持 `binding:"required"`

## 前端 StorageConfigView.vue
- [x] provider=tencent 时 Endpoint 字段标签显示 "Bucket URL"
- [x] provider=tencent 时 endpoint placeholder 为 `https://examplebucket-1250000000.cos.ap-guangzhou.myqcloud.com`
- [x] provider=tencent 时 access_key 标签显示 "Secret ID"
- [x] provider=tencent 时 access_secret 标签显示 "Secret Key"
- [x] provider=tencent 时隐藏 Bucket 字段
- [x] provider=aliyun/minio 时表单字段和标签不变
- [x] 提交腾讯云配置时 endpoint 字段 = bucket_url 输入值

## 编译验证
- [x] 后端 `go build ./...` 通过
- [x] 后端 `go vet ./...` 通过
- [x] 前端 `npm run build` 通过
