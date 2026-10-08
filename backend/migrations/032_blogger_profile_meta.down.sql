-- 回滚博主个人资料的性格/星座与对外展示开关字段
ALTER TABLE bloggers DROP COLUMN IF EXISTS show_personality;
ALTER TABLE bloggers DROP COLUMN IF EXISTS show_zodiac;
ALTER TABLE bloggers DROP COLUMN IF EXISTS show_city;
ALTER TABLE bloggers DROP COLUMN IF EXISTS show_email;
ALTER TABLE bloggers DROP COLUMN IF EXISTS zodiac;
ALTER TABLE bloggers DROP COLUMN IF EXISTS personality;
