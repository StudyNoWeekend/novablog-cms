# 系统初始化引导 Spec

## Why
Cus-CMS 是单博主系统，刚部署后数据库中没有任何博主数据。用户无法登录，系统应提供初始化引导流程，让用户创建第一个博主账号，而不是直接展示登录页面导致用户困惑。

## What Changes
- 后端新增 `/api/v1/public/setup/status` 接口，返回系统是否已初始化
- 后端新增 `/api/v1/public/setup/init` 接口，用于创建初始博主账号（仅在未初始化时可调用）
- 前端新增初始化引导页面 `SetupView.vue`，包含博主账号创建表单
- 前端路由守卫增加初始化状态检查，未初始化时跳转引导页

## Impact
- Affected specs: cus-cms-phase1-auth-layout（扩展登录流程）
- Affected code: 
  - 后端：`internal/router/public.go`、`internal/controller/setup_controller.go`、`internal/logic/setup_logic.go`
  - 前端：`src/views/setup/SetupView.vue`、`src/router/guards.ts`、`src/api/setup.ts`

---

## ADDED Requirements

### Requirement: 系统初始化状态检查接口
系统 SHALL 提供公开接口检查是否已完成初始化。

#### Scenario: 查询初始化状态
- **WHEN** 前端调用 `GET /api/v1/public/setup/status`
- **THEN** 后端查询 bloggers 表是否存在记录
- **AND** 返回 `{ initialized: boolean }`

#### Scenario: 已初始化状态
- **WHEN** bloggers 表中存在至少一条记录
- **THEN** 返回 `{ initialized: true }`

#### Scenario: 未初始化状态
- **WHEN** bloggers 表为空
- **THEN** 返回 `{ initialized: false }`

---

### Requirement: 初始化博主账号接口
系统 SHALL 提供公开接口创建初始博主账号（仅在未初始化状态可调用）。

#### Scenario: 创建初始博主
- **WHEN** 前端调用 `POST /api/v1/public/setup/init`（系统未初始化）
- **AND** 请求包含 `username`、`password`、`nickname`（可选）
- **THEN** 后端创建博主记录（密码 bcrypt 哈希）
- **AND** 返回 `{ success: true, message: "初始化成功" }`

#### Scenario: 已初始化时拒绝创建
- **WHEN** 前端调用 `POST /api/v1/public/setup/init`（系统已初始化）
- **THEN** 返回 `code: 403000, msg: "系统已初始化，无法重复创建", httpCode: 403`

#### Scenario: 参数校验
- **WHEN** 请求缺少 `username` 或 `password`
- **THEN** 返回 `code: 400001, msg: "请求参数错误", httpCode: 400`

#### Scenario: 用户名格式校验
- **WHEN** username 长度小于 3 或大于 50
- **THEN** 返回 `code: 400001, msg: "用户名长度需在 3-50 之间", httpCode: 400`

#### Scenario: 密码格式校验
- **WHEN** password 长度小于 6
- **THEN** 返回 `code: 400001, msg: "密码长度至少 6 位", httpCode: 400`

---

### Requirement: 前端初始化引导页面
系统 SHALL 提供初始化引导页面，引导用户创建博主账号。

#### Scenario: 引导页面展示
- **WHEN** 系统未初始化，用户访问任何页面
- **THEN** 跳转至 `/setup` 初始化引导页
- **AND** 页面展示：
  - Logo + "Cus CMS" 标题
  - "欢迎使用 Cus CMS" 欢迎语
  - "创建您的博主账号以开始使用系统" 说明文字
  - 用户名输入框（placeholder: "设置登录用户名"）
  - 密码输入框（placeholder: "设置登录密码"，含可见切换）
  - 昵称输入框（placeholder: "您的昵称"，可选）
  - "完成初始化" 按钮
  - 底部品牌标语

#### Scenario: 表单校验
- **WHEN** 用户输入不合规信息
- **THEN** 实时显示校验错误提示
- **AND** 用户名：3-50 字符
- **AND** 密码：至少 6 位

#### Scenario: 初始化成功
- **WHEN** 用户提交有效信息并初始化成功
- **THEN** 显示成功提示 "初始化成功！请使用新账号登录"
- **AND** 自动跳转至登录页 `/auth/login`

#### Scenario: 初始化失败
- **WHEN** 初始化请求失败（网络错误、参数错误等）
- **THEN** 显示错误提示，用户可重新提交

---

### Requirement: 前端路由守卫扩展
系统 SHALL 在路由守卫中增加初始化状态检查。

#### Scenario: 未初始化时拦截
- **WHEN** 用户访问任何页面（包括登录页）
- **AND** 系统未初始化
- **THEN** 自动跳转至 `/setup` 引导页

#### Scenario: 已初始化时跳过引导页
- **WHEN** 用户访问 `/setup`
- **AND** 系统已初始化
- **THEN** 自动跳转至 `/auth/login`

#### Scenario: 初始化状态缓存
- **WHEN** 前端首次获取初始化状态
- **THEN** 将状态缓存至 localStorage（key: `cus_cms_initialized`）
- **AND** 后续路由守卫优先读取缓存，减少 API 调用
- **WHEN** 初始化成功后
- **THEN** 更新缓存为 `initialized: true`