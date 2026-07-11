ALTER TABLE travel_guides DROP CONSTRAINT IF EXISTS fk_travel_guides_category;
ALTER TABLE travel_guides DROP COLUMN IF EXISTS category_id;
