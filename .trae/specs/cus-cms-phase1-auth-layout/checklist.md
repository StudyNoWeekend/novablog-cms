# Checklist

## 后端项目脚手架
- [x] Go module 已初始化，`go.mod` 包含 gin、gorm、jwt、zap、bcrypt、redis、uuid 等依赖
- [x] 目录结构符合 `gin-gorm-backend` 规范（cmd/config/bootstrap/internal/utils/enum/migrations）
- [x] `config/config.yaml` 包含 app/http/log/mysql/redis/jwt 配置段
- [x] `bootstrap/` 模块完成 DB、Redis、Logger 初始化

## 后端认证模块
- [x] `enum/error.go` 定义了 BizError 结构体与所有错误码常量
- [x] `utils/response/response.go` 实现了 Success/Fail 统一响应
- [x] `utils/jwt/jwt.go` 实现了 Access Token（2h）+ Refresh Token（7d）生成与解析
- [x] `utils/hash/bcrypt.go` 实现了密码哈希与校验
- [x] `internal/model/blogger.go` 定义了 Blogger 表结构（UUID 主键）与 CRUD 方法
- [x] `internal/cache/token_cache.go` 实现了 Token 黑名单 Redis 操作
- [x] `internal/dto/req/auth_req.go` 定义了 LoginReq、RefreshReq、ChangePasswordReq
- [x] `internal/dto/res/auth_res.go` 定义了 LoginRes
- [x] `internal/logic/auth_logic.go` 实现了 Login/Refresh/Logout/ChangePassword 业务逻辑
- [x] `internal/controller/auth_controller.go` 调用 logic，使用 response.Fail 处理 BizError
- [x] Controller 不直接操作数据库，Logic 不依赖 Gin Context

## 后端中间件与路由
- [x] `internal/middleware/auth.go` JWT 中间件：解析 Token、检查黑名单、无效/过期返回 401
- [x] `internal/middleware/cors.go` 跨域处理
- [x] `internal/middleware/recovery.go` Panic 恢复
- [x] `internal/middleware/logger.go` 访问日志（zap）
- [x] `internal/middleware/trace.go` TraceID 注入
- [x] `internal/router/router.go` 路由总入口 + `/health`、`/ready` 端点
- [x] `internal/router/auth.go` 认证路由注册（login/refresh/logout 无需认证中间件，password 需要）
- [x] `cmd/api/main.go` 实现优雅关闭（SIGINT/SIGTERM）

## 数据库迁移
- [x] `migrations/001_init_schema.up.sql` 包含 blogger 表建表语句（UUID 主键、bcrypt 密码哈希字段）

## 后端代码质量
- [x] 所有导出类型/方法有中文注释
- [x] 表主键使用 UUID 而非自增 ID
- [x] 密码使用 bcrypt 哈希存储
- [x] 错误处理不吞错，每个 err != nil 记录日志
- [x] 日志统一使用 zap
- [x] 无 goto 语句

## 前端项目脚手架
- [x] Vite 5 + Vue 3 + TypeScript 项目已初始化
- [x] 已安装 ant-design-vue、pinia、vue-router、axios、@ant-design/icons-vue、unocss
- [x] `uno.config.ts` 定义了品牌色、快捷方式
- [x] CSS 变量文件定义了设计文档规定的色彩/阴影/间距/动效变量

## 前端基础设施
- [x] `src/types/api.ts` 定义了 ApiResponse<T>、PaginatedData<T> 等类型
- [x] `src/utils/storage.ts` 封装了 localStorage 操作
- [x] `src/api/request.ts` Axios 封装：请求拦截器注入 Authorization Header、响应拦截器处理 code 非 0 与 HTTP 401
- [x] `src/stores/auth.ts` 实现了 login/logout/refreshToken 动作
- [x] `src/stores/app.ts` 实现了 sidebarCollapsed/breadcrumbs 状态
- [x] `src/router/guards.ts` 实现了路由守卫（白名单放行、未登录跳转、已登录重定向）

## 登录页面
- [x] LoginView 包含 Logo、"创作者后台"标题、用户名/邮箱输入框、密码输入框（含切换图标）、"记住我"复选框、"忘记密码"链接、登录按钮
- [x] 页面底部显示品牌标语
- [x] 表单实时校验（空值提示）
- [x] 登录按钮 loading 状态
- [x] 登录成功跳转 Dashboard
- [x] 登录失败展示错误提示
- [x] 登录页视觉符合设计文档规范（墨蓝品牌色、大面积留白、字体系统）

## CMS 主页面布局
- [x] AppLayout 实现左侧侧边栏 + 右侧 Header + 内容区三段式布局
- [x] AppSidebar 深色背景（#111827）、240px 宽、导航菜单、折叠/展开
- [x] AppHeader 包含面包屑、通知图标、用户头像下拉菜单
- [x] 侧边栏折叠/展开动画过渡 350ms
- [x] 当前路由菜单项高亮（左侧蓝色竖条 + 背景高亮）
- [x] DashboardView 包含统计卡片、快速操作入口、最近文章列表

## 401 未授权处理
- [x] Axios 响应拦截器在 HTTP 401 时自动尝试 Refresh Token 刷新
- [x] Refresh Token 刷新失败时清除 Token 并跳转登录页
- [x] 路由守卫未登录访问后台页面时跳转 `/auth/login?redirect=原路径`
- [x] 已登录访问登录页时重定向 Dashboard
- [x] "记住我"功能：Token 持久化到 localStorage

## 联调验证
- [x] 后端 `/health` 返回 200
- [x] 后端 `/ready` 返回 200（DB 可用时）
- [x] `POST /api/v1/auth/login` 正确用户名密码返回 Token 对
- [x] `POST /api/v1/auth/login` 错误凭据返回 401
- [x] `POST /api/v1/auth/refresh` 有效 Refresh Token 返回新 Token 对
- [x] `POST /api/v1/auth/logout` 成功登出
- [x] 不带 Token 访问受保护接口返回 401
- [x] 前端登录页渲染正常，各交互元素工作正常
- [x] 前端登录成功跳转 Dashboard，主页面布局正常
- [x] 前端 401 自动刷新 Token 或跳转登录页
- [x] 路由守卫拦截与 redirect 参数正常工作