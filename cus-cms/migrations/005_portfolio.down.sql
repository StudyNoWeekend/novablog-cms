-- 回滚摄影作品集表迁移

DROP INDEX IF EXISTS idx_portfolio_items_sort_order;
DROP INDEX IF EXISTS idx_portfolio_items_deleted_at;
DROP INDEX IF EXISTS idx_portfolio_items_preset_id;
DROP INDEX IF EXISTS idx_portfolio_items_portfolio_id;
DROP TABLE IF EXISTS portfolio_items;

DROP INDEX IF EXISTS idx_portfolios_sort_order;
DROP INDEX IF EXISTS idx_portfolios_status;
DROP INDEX IF EXISTS idx_portfolios_deleted_at;
DROP TABLE IF EXISTS portfolios;
