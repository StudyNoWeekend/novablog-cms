# 验证清单

## API 层与类型
- [x] `src/types/storage.ts` 定义了 StorageConfig（含 is_active / access_secret / extra）、MigrationTask、MigrationItem、AnalyzeResult 等类型
- [x] `src/api/storage.ts` 封装 5 个存储配置 API（getList/upsert/delete/activate/test）和 4 个迁移 API（analyze/getAnalyzeResult/startMigration/getMigrationStatus）
- [x] upsert 请求体包含所有必填字段
- [x] test 请求体 access_secret 为明文
- [x] startMigration 支持 media_ids[]（批量）和 all=true（一键）

## SettingsLayout 子导航
- [x] `SettingsLayout.vue` 左侧垂直 Menu 包含「个人资料」和「对象存储」两个菜单项
- [x] 点击左侧菜单，右侧内容区通过 `<router-view />` 切换对应子路由组件
- [x] 当前路由激活时，左侧菜单项高亮（使用 Menu selectedKeys）
- [x] 布局为卡片样式，左侧导航宽度 200px，有分隔线

## 存储配置管理
- [x] `StorageConfigView.vue` 顶部为 Tabs（存储配置 / 素材迁移），默认显示存储配置
- [x] 配置列表表格展示：provider（中文映射 阿里云 OSS / 腾讯云 COS / MinIO）、endpoint、bucket、is_active（Tag 标签）、updated_at
- [x] 已激活行有绿色「当前使用中」Tag，操作列激活按钮禁用
- [x] 未激活行操作列有「激活」「删除」「编辑」按钮
- [x] 新增 Modal 表单：provider 下拉联动显示平台参数提示
  - 阿里云：endpoint 示例、region 必填提示、bucket 纯桶名提示
  - 腾讯云：endpoint 示例、bucket 带 APPID 提示
  - MinIO：endpoint 不含 scheme 提示、use_ssl 开关
- [x] 编辑 Modal：access_secret 留空表示不修改，placeholder「留空不修改」
- [x] 测试连通性按钮调 test 接口，结果显示成功/失败
- [x] 激活确认弹窗「激活后将切换存储平台，确定继续？」
- [x] 激活失败「迁移进行中」显示相应提示
- [x] 删除确认弹窗，active 配置不可删时显示提示

## 素材迁移
- [x] 迁移 Tab 显示当前活跃平台名称
- [x] 「开始分析」按钮调 analyze 接口，2 秒轮询分析结果
- [x] 分析中显示 Spin 加载状态
- [x] 分析完成展示统计：「共 N 条素材，M 条已存在，K 条需要迁移」
- [x] 「已存在」表格灰色/不可操作，列展示 filename + storage_type
- [x] 「需迁移」表格高亮、可勾选
- [x] 「一键迁移所有」按钮调 startMigration({ all: true })
- [x] 「批量迁移」按钮仅勾选时可用
- [x] 迁移进度展示：进度条 + 实时计数 + 2 秒轮询
- [x] 迁移完成：进度条 100%、显示总耗时
- [x] 失败项可展开查看 error 详情
- [x] 迁移完成后提供「重新分析」按钮
- [x] 无活跃配置时显示提示「请先在「存储配置」中配置并激活一个存储平台」

## 路由与菜单
- [x] `/profile` 使用 SettingsLayout 布局，redirect → `/profile/info`
- [x] `/profile/info` 渲染 ProfileView（个人资料）
- [x] `/profile/storage` 渲染 StorageConfigView（对象存储）
- [x] `AppSidebar.vue` 中「设置」链接为 `/profile/info`
- [x] `ProfileView.vue` 精简为纯内容组件

## 编译验证
- [x] `npm run build` 通过，无 TypeScript 编译错误
- [x] 无未使用的 import
