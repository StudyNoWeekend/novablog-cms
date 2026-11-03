# migrations 目录（历史 SQL 迁移文件已移除）

> **清理说明（2026-11）**：本目录原归档的 88 个历史迁移 SQL（`archive/001` ~ `archive/044`）
> 已全部删除。它们从未被任何代码或部署脚本执行（全仓无 golang-migrate/goose/自研 runner 引用），
> 且 9 张表从未有 CREATE TABLE 迁移，纯手工执行这些脚本无法从零建出完整库。

## 当前事实源

数据库的表结构与数据级变更以 `backend/bootstrap/` 中的启动逻辑为**唯一事实源**：

| 启动步骤 | 代码位置 | 职责 |
|---------|---------|------|
| 建库 | `bootstrap/db.go` `ensureDatabaseExists` | 数据库不存在时自动 CREATE DATABASE |
| 建表/加列 | `bootstrap/db.go` `autoMigrate` | GORM AutoMigrate 全部 45 个 model |
| 废弃对象清理 | `bootstrap/deprecated_cleanup.go` | 幂等 DROP 已退役的孤儿表/列（见下表） |
| 分类回填 | `bootstrap/category_backfill.go` | 旧分类文本播种为统一分类并回填 category_id；slug 唯一约束改为 `(type, slug)` |
| 计数回填 | `bootstrap/db.go` `backfillCounters` | 校正 articles/travel_guides 评论计数 |
| 媒体文件夹首装 | `bootstrap/media_folder_init.go` | 播种 12 个模块顶级文件夹并归类存量媒体 |
| 媒体引用回填 | `bootstrap/media_reference_backfill.go`（异步） | media.url 归一 + 14 组业务表图片字段回填媒体引用 + 遗留文件收编 |
| 配置播种 | `bootstrap/config_seed.go` | config.yaml 旧键幂等导入 system_configs |

## 废弃库对象的清理（已自动化）

以下孤儿对象由 `bootstrap/deprecated_cleanup.go` 在**每次启动时幂等清理**
（`DROP ... IF EXISTS`，失败仅记日志不阻塞启动），无需手工执行：

| 对象 | 废弃原因 |
|------|---------|
| `ip_blacklist` 整表（注意不是 `ip_blacklists`） | 旧版黑名单，运行时黑名单走 `ip_blacklists` 表 + Redis，模型已删除 |
| `theme_market_configs.market_base_url` 列 | 官方站点地址已统一托管到 community_push_configs，模型字段已删 |
| `security_configs` 的 9 个旧限流列（get/post/view/like_max_tokens、对应 4 个 window_seconds、blacklist_threshold） | 限流策略重构后模型删字段，AutoMigrate 不删列，全仓零读写 |

## 示例数据

示例内容种子脚本在 `backend/scripts/seed_photographer_data.sql`，
可通过 `go run ./cmd/seed` 或 psql 执行（详见该文件头注释）。
