ALTER TABLE portfolios DROP CONSTRAINT IF EXISTS fk_portfolios_category;
ALTER TABLE portfolios DROP COLUMN IF EXISTS category_id;
