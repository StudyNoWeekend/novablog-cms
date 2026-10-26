-- 回滚：移除业务表媒体引用列与 media_presets.storage_type。
-- 注意：media.url / media_presets.output_url 的归一（完整 URL → 相对对象键）不可逆，
-- 回滚后需恢复 media 表 storage_type 对应平台的原始完整 URL（可重新执行存储迁移工具生成）。

ALTER TABLE articles DROP COLUMN IF EXISTS cover_media_id;
ALTER TABLE articles DROP COLUMN IF EXISTS cover_preset_id;
ALTER TABLE songs DROP COLUMN IF EXISTS cover_media_id;
ALTER TABLE songs DROP COLUMN IF EXISTS cover_preset_id;
ALTER TABLE third_party_playlists DROP COLUMN IF EXISTS cover_media_id;
ALTER TABLE third_party_playlists DROP COLUMN IF EXISTS cover_preset_id;
ALTER TABLE video_works DROP COLUMN IF EXISTS cover_media_id;
ALTER TABLE video_works DROP COLUMN IF EXISTS cover_preset_id;
ALTER TABLE travel_guides DROP COLUMN IF EXISTS cover_media_id;
ALTER TABLE travel_guides DROP COLUMN IF EXISTS cover_preset_id;
ALTER TABLE recipes DROP COLUMN IF EXISTS cover_media_id;
ALTER TABLE recipes DROP COLUMN IF EXISTS cover_preset_id;
ALTER TABLE books DROP COLUMN IF EXISTS cover_media_id;
ALTER TABLE books DROP COLUMN IF EXISTS cover_preset_id;
ALTER TABLE games DROP COLUMN IF EXISTS cover_media_id;
ALTER TABLE games DROP COLUMN IF EXISTS cover_preset_id;
ALTER TABLE photo_equipment DROP COLUMN IF EXISTS image_media_id;
ALTER TABLE photo_equipment DROP COLUMN IF EXISTS image_preset_id;
ALTER TABLE projects DROP COLUMN IF EXISTS cover_media_id;
ALTER TABLE projects DROP COLUMN IF EXISTS cover_preset_id;
ALTER TABLE tech_stack_items DROP COLUMN IF EXISTS icon_media_id;
ALTER TABLE bloggers DROP COLUMN IF EXISTS avatar_media_id;
ALTER TABLE bloggers DROP COLUMN IF EXISTS blog_icon_media_id;
ALTER TABLE bloggers DROP COLUMN IF EXISTS page_background_media_id;
ALTER TABLE media_presets DROP COLUMN IF EXISTS storage_type;
