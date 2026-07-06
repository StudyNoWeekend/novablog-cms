# Tasks

- [x] Task 1: 后端初始化状态检查接口
  - [x] SubTask 1.1: 编写 `internal/dto/req/setup_req.go` 初始化请求 DTO（InitReq: username/password/nickname）
  - [x] SubTask 1.2: 编写 `internal/dto/res/setup_res.go` 初始化响应 DTO（StatusRes: initialized, InitRes: success/message）
  - [x] SubTask 1.3: 编写 `internal/logic/setup_logic.go` 初始化业务逻辑（CheckStatus/InitBlogger）
  - [x] SubTask 1.4: 编写 `internal/controller/setup_controller.go` 初始化控制器（GetStatus/Init）
  - [x] SubTask 1.5: 在 `internal/router/public.go` 中注册初始化路由（GET /setup/status, POST /setup/init）

- [x] Task 2: 前端初始化引导页面
  - [x] SubTask 2.1: 编写 `src/api/setup.ts` 初始化 API 接口（getStatus/init）
  - [x] SubTask 2.2: 编写 `src/views/setup/SetupView.vue` 初始化引导页面组件
  - [x] SubTask 2.3: 实现用户名输入框（校验 3-50 字符）
  - [x] SubTask 2.4: 实现密码输入框（校验至少 6 位，含可见切换）
  - [x] SubTask 2.5: 实现昵称输入框（可选）
  - [x] SubTask 2.6: 实现"完成初始化"按钮（loading 状态）
  - [x] SubTask 2.7: 实现初始化成功后跳转登录页
  - [x] SubTask 2.8: 实现页面视觉符合设计文档规范（深色背景、品牌色）

- [x] Task 3: 前端路由守卫扩展
  - [x] SubTask 3.1: 在 `src/router/routes.ts` 中添加 `/setup` 路由
  - [x] SubTask 3.2: 在 `src/router/guards.ts` 中添加初始化状态检查逻辑
  - [x] SubTask 3.3: 在 `src/utils/storage.ts` 中添加初始化状态缓存读写方法
  - [x] SubTask 3.4: 实现未初始化时拦截跳转 `/setup`
  - [x] SubTask 3.5: 实现已初始化时访问 `/setup` 跳转 `/auth/login`

- [x] Task 4: 联调验证
  - [x] SubTask 4.1: 验证 `GET /api/v1/public/setup/status` 空数据库返回 `{ initialized: false }`
  - [x] SubTask 4.2: 验证 `POST /api/v1/public/setup/init` 创建博主成功
  - [x] SubTask 4.3: 验证已初始化后 `/setup/status` 返回 `{ initialized: true }`
  - [x] SubTask 4.4: 验证已初始化后 `/setup/init` 返回 403
  - [x] SubTask 4.5: 验证前端未初始化时自动跳转 `/setup`
  - [x] SubTask 4.6: 验证初始化成功后跳转登录页
  - [x] SubTask 4.7: 验证已初始化后访问 `/setup` 跳转登录页

# Task Dependencies
- Task 2 依赖 Task 1（需要后端 API）
- Task 3 依赖 Task 1、Task 2（需要 API 和页面）
- Task 4 依赖 Task 1、Task 2、Task 3