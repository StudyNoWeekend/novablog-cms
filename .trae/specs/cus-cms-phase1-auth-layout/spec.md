# Cus-CMS Phase 1: 认证系统与主布局 Spec

## Why
Cus-CMS 是面向博主的单用户内容管理系统。Phase 1 需要建立项目基础架构，实现博主身份认证全流程（登录、Token 管理、鉴权）以及系统主页面布局，为后续业务模块开发奠定基础。

## What Changes
- 在 `cus-cms/` 目录下创建前后端项目脚手架
- 后端（Go + Gin + GORM）：实现认证模块（登录、Token 刷新、登出、密码修改）
- 前端（Vue 3 + Ant Design Vue + Pinia）：实现登录页、系统主页面布局、路由守卫
- 建立前后端统一状态码规范，特别是 401 未授权自动跳转登录页的前端处理机制
- 后端严格遵循 `gin-gorm-backend` 架构规范，前端遵循设计文档的视觉规范

## Impact
- Affected specs: 无（新项目）
- Affected code: `cus-cms/` 目录下全部新建代码

---

## ADDED Requirements

### Requirement: 项目脚手架搭建
系统 SHALL 在 `cus-cms/` 目录下创建符合规范的前后端项目结构。

#### Scenario: 后端项目结构
- **WHEN** 开发人员查看 `cus-cms/` 目录
- **THEN** 后端代码按 `cmd/`、`config/`、`bootstrap/`、`internal/`（含 `middleware/`、`controller/`、`dto/`、`logic/`、`model/`、`cache/`、`router/`）、`utils/`、`enum/`、`migrations/` 分层组织
- **AND** 提供 `Makefile`、`.golangci.yml`、`go.mod`

#### Scenario: 前端项目结构
- **WHEN** 开发人员查看 `cus-cms/cus-cms-web/` 目录
- **THEN** 前端代码按 `src/api/`、`src/components/`（含 `common/`、`layout/`）、`src/views/`、`src/router/`、`src/stores/`、`src/utils/`、`src/types/`、`src/assets/styles/` 分层组织
- **AND** 提供 `package.json`、`vite.config.ts`、`tsconfig.json`、`uno.config.ts`

---

### Requirement: 登录页面
系统 SHALL 提供完整的博主登录认证界面。

#### Scenario: 登录页元素完整展示
- **WHEN** 用户访问登录页
- **THEN** 页面展示 Logo、"创作者后台"标题、用户名/邮箱输入框、密码输入框（含可见/不可见切换图标）、"记住我"复选框、"忘记密码"链接、登录按钮
- **AND** 页面底部展示 "用专业的方式，展示你的创作" 品牌标语
- **AND** 整体视觉符合设计文档定义的"创作者的暗房"设计理念：大面积留白、品牌色点缀（墨蓝 #1a1a2e + 电光蓝 #4a6cf7）

#### Scenario: 表单实时校验
- **WHEN** 用户输入不合规的登录信息（空用户名、空密码）
- **THEN** 系统实时显示校验错误提示
- **AND** 输入框 focus 时边框变亮色（电光蓝）

#### Scenario: 登录按钮交互状态
- **WHEN** 用户点击登录按钮
- **THEN** 按钮进入 loading 状态（脉冲动画），禁用重复点击
- **WHEN** 登录成功
- **THEN** 过渡动画跳转至 Dashboard
- **WHEN** 登录失败
- **THEN** 显示错误提示，按钮恢复可点击状态

---

### Requirement: 后端身份认证逻辑
系统 SHALL 实现完整的 JWT 双令牌认证机制。

#### Scenario: 博主登录
- **WHEN** 前端发送 `POST /api/v1/auth/login` 请求（含 username + password）
- **THEN** 后端校验用户名与 bcrypt 哈希密码
- **AND** 校验成功后生成 Access Token（有效期 2h）和 Refresh Token（有效期 7d）
- **AND** 响应包含 `{ access_token, refresh_token, expires_in }`
- **AND** 更新 blogger 记录的 `last_login_at`

#### Scenario: 登录参数校验
- **WHEN** 前端发送的登录请求缺少 username 或 password
- **THEN** 返回 `code: 400001, msg: "请求参数错误", httpCode: 400`

#### Scenario: 登录凭据错误
- **WHEN** 用户名或密码不正确
- **THEN** 返回 `code: 401001, msg: "用户名或密码错误", httpCode: 401`

#### Scenario: Token 刷新
- **WHEN** 前端发送 `POST /api/v1/auth/refresh` 请求（含 refresh_token）
- **THEN** 后端校验 Refresh Token 有效性
- **AND** 生成新的 Access Token + Refresh Token 对
- **AND** 将旧的 Refresh Token 加入 Redis 黑名单

#### Scenario: 登出
- **WHEN** 前端发送 `POST /api/v1/auth/logout` 请求（含 Authorization Header）
- **THEN** 后端将当前 Access Token 的 JTI 加入 Redis 黑名单，TTL 为 Token 剩余有效期
- **AND** 返回成功响应

#### Scenario: 未认证访问受保护接口
- **WHEN** 前端不带有效 Token 访问 `/api/v1/*` 后台接口
- **THEN** 后端 JWT 中间件拦截，返回 `code: 401000, msg: "未认证，请先登录", httpCode: 401`

---

### Requirement: CMS 主页面布局
系统 SHALL 提供登录后的内容管理系统主页面。

#### Scenario: 侧边栏导航
- **WHEN** 用户登录后进入主页面
- **THEN** 左侧显示深色侧边栏（240px 宽，背景色 #111827）
- **AND** 顶部展示 Logo，下方为导航菜单（工作台/文章管理/分类标签/媒体库/摄影作品/视频管理/评论管理/模板风格）
- **AND** 当前路由对应菜单项高亮（左侧蓝色竖条 + 背景高亮）
- **AND** 侧边栏支持折叠/展开（折叠后 72px 仅显示图标）

#### Scenario: 顶部栏
- **WHEN** 用户登录后进入主页面
- **THEN** 顶部显示 Header（56px 高），包含侧边栏折叠按钮、面包屑导航、通知图标、用户头像下拉菜单
- **AND** 用户下拉菜单包含"个人资料"和"退出登录"选项

#### Scenario: 内容区
- **WHEN** 用户登录后进入主页面
- **THEN** 主内容区居中显示（max-width: 1400px），展示 Dashboard 概览
- **AND** Dashboard 包含统计卡片（文章数、访问量、评论数、作品数）、快速操作入口、最近文章列表

#### Scenario: 退出登录
- **WHEN** 用户点击"退出登录"
- **THEN** 系统调用退出接口，清除本地 Token，跳转至登录页

---

### Requirement: 前后端状态码规范
系统 SHALL 建立统一的状态码规范体系，前端实现 401 未授权自动跳转。

#### Scenario: 统一状态码定义
- **WHEN** 后端返回任何响应
- **THEN** 响应格式为 `{ code: int, msg: string, data: any, trace_id: string }`
- **AND** code=0 表示成功
- **AND** 4xxxxx 表示客户端错误（400xxx 参数错误、401xxx 未认证、403xxx 无权限、404xxx 资源不存在）
- **AND** 5xxxxx 表示服务端错误

| code | 含义 | HTTP 状态码 |
|------|------|-------------|
| 0 | 成功 | 200 |
| 400001 | 请求参数错误 | 400 |
| 401000 | 未认证，请先登录 | 401 |
| 401001 | 用户名或密码错误 | 401 |
| 401002 | Token 已过期 | 401 |
| 401003 | Token 无效 | 401 |
| 403000 | 无权限访问该资源 | 403 |
| 404001 | 资源不存在 | 404 |
| 500001 | 系统内部错误 | 500 |

#### Scenario: 前端 401 自动跳转
- **WHEN** 前端 Axios 响应拦截器接收到 HTTP 401 响应
- **THEN** 系统尝试使用 Refresh Token 刷新 Access Token
- **WHEN** Refresh Token 刷新也失败
- **THEN** 清除本地所有 Token 与用户状态
- **AND** 立即自动跳转至登录页 `/auth/login`
- **AND** 登录成功后重定向回原访问页面

#### Scenario: 路由守卫拦截
- **WHEN** 用户未登录直接访问后台页面（如 `/dashboard`）
- **THEN** Vue Router 路由守卫拦截，跳转至 `/auth/login?redirect=/dashboard`
- **WHEN** 用户已登录访问登录页
- **THEN** 路由守卫自动重定向至 `/dashboard`