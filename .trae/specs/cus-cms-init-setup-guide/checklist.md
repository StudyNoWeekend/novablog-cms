# Checklist

## 后端初始化接口
- [x] `internal/dto/req/setup_req.go` 定义了 InitReq（username/password/nickname）
- [x] `internal/dto/res/setup_res.go` 定义了 StatusRes（initialized）和 InitRes（success/message）
- [x] `internal/logic/setup_logic.go` 实现了 CheckStatus（查询博主是否存在）和 InitBlogger（创建博主）
- [x] `internal/controller/setup_controller.go` 实现了 GetStatus 和 Init 控制器方法
- [x] Init 方法在已初始化时返回 403
- [x] 参数校验：username 3-50 字符、password 至少 6 位
- [x] 密码使用 bcrypt 哈希存储

## 后端路由注册
- [x] `internal/router/public.go` 注册了 GET /setup/status（无需认证）
- [x] `internal/router/public.go` 注册了 POST /setup/init（无需认证，但内部检查初始化状态）

## 前端初始化页面
- [x] `src/api/setup.ts` 定义了 getStatus 和 init 接口
- [x] `src/views/setup/SetupView.vue` 包含 Logo、欢迎语、说明文字
- [x] 用户名输入框校验 3-50 字符
- [x] 密码输入框校验至少 6 位、含可见切换
- [x] 昵称输入框可选
- [x] "完成初始化"按钮有 loading 状态
- [x] 初始化成功后显示提示并跳转登录页
- [x] 页面视觉符合设计文档规范（深色背景、品牌色）

## 前端路由守卫
- [x] `src/router/routes.ts` 添加了 `/setup` 路由
- [x] `src/router/guards.ts` 在守卫开始时检查初始化状态
- [x] 未初始化时自动跳转 `/setup`
- [x] 已初始化时访问 `/setup` 跳转 `/auth/login`
- [x] `src/utils/storage.ts` 添加了初始化状态缓存方法

## 联调验证
- [x] 空数据库时 `/setup/status` 返回 `{ initialized: false }`
- [x] `/setup/init` 创建博主成功
- [x] 创建后 `/setup/status` 返回 `{ initialized: true }`
- [x] 已初始化后 `/setup/init` 返回 403
- [x] 前端未初始化时自动跳转 `/setup`
- [x] 初始化成功后跳转登录页
- [x] 已初始化后访问 `/setup` 跳转登录页