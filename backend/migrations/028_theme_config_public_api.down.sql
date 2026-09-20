-- 回滚博客公开 API 地址字段
ALTER TABLE theme_market_configs DROP COLUMN IF EXISTS public_api_base;
