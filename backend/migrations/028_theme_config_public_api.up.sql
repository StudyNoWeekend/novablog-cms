-- 官方市场配置表扩展：博客公开 API 地址（独立 API 域名场景，注入主题 theme-config.js）
ALTER TABLE theme_market_configs ADD COLUMN IF NOT EXISTS public_api_base VARCHAR(500) NOT NULL DEFAULT '';
