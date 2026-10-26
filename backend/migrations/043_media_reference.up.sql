-- 媒体引用改造：业务表图片字段支持引用媒体库文件 ID（media.id / media_presets.id），
-- 读取时由后端按当前存储配置动态拼接完整 URL，存储迁移/换域名后自动跟随，不再写死落库 URL。
-- 同时把 media.url / media_presets.output_url 归一为相对对象键（与 storage_path 一致），
-- 完整 URL 一律读取时拼接。
--
-- 说明：本项目无自动执行 migrations/*.sql 的机制（AutoMigrate 只同步表结构），
-- 数据级回填另由 bootstrap/media_reference_backfill.go 在启动时异步幂等执行，
-- 本脚本与 Go 回填逻辑一致、可重复执行、互不冲突；供手工执行部署时使用。

-- ============================================================
-- 1. media_presets 补 storage_type（预设成品图所在平台，继承所属媒体）
-- ============================================================
ALTER TABLE media_presets ADD COLUMN IF NOT EXISTS storage_type VARCHAR(20);
COMMENT ON COLUMN media_presets.storage_type IS '预设成品图所在存储平台（local/aliyun/tencent/minio），用于动态拼接 output_url';

-- ============================================================
-- 2. 业务表新增媒体引用列
-- ============================================================
ALTER TABLE articles ADD COLUMN IF NOT EXISTS cover_media_id UUID;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS cover_preset_id UUID;
COMMENT ON COLUMN articles.cover_media_id IS '封面媒体文件ID，关联 media.id；非空时 cover_image 留空';
COMMENT ON COLUMN articles.cover_preset_id IS '封面预设ID，关联 media_presets.id；非空时优先按预设解析';
CREATE INDEX IF NOT EXISTS idx_articles_cover_media_id ON articles(cover_media_id);
CREATE INDEX IF NOT EXISTS idx_articles_cover_preset_id ON articles(cover_preset_id);

ALTER TABLE songs ADD COLUMN IF NOT EXISTS cover_media_id UUID;
ALTER TABLE songs ADD COLUMN IF NOT EXISTS cover_preset_id UUID;
COMMENT ON COLUMN songs.cover_media_id IS '封面媒体文件ID，关联 media.id；非空时 cover_url 留空';
COMMENT ON COLUMN songs.cover_preset_id IS '封面预设ID，关联 media_presets.id；非空时优先按预设解析';
CREATE INDEX IF NOT EXISTS idx_songs_cover_media_id ON songs(cover_media_id);
CREATE INDEX IF NOT EXISTS idx_songs_cover_preset_id ON songs(cover_preset_id);

ALTER TABLE third_party_playlists ADD COLUMN IF NOT EXISTS cover_media_id UUID;
ALTER TABLE third_party_playlists ADD COLUMN IF NOT EXISTS cover_preset_id UUID;
COMMENT ON COLUMN third_party_playlists.cover_media_id IS '封面媒体文件ID，关联 media.id；非空时 cover_url 留空';
COMMENT ON COLUMN third_party_playlists.cover_preset_id IS '封面预设ID，关联 media_presets.id；非空时优先按预设解析';
CREATE INDEX IF NOT EXISTS idx_third_party_playlists_cover_media_id ON third_party_playlists(cover_media_id);
CREATE INDEX IF NOT EXISTS idx_third_party_playlists_cover_preset_id ON third_party_playlists(cover_preset_id);

ALTER TABLE video_works ADD COLUMN IF NOT EXISTS cover_media_id UUID;
ALTER TABLE video_works ADD COLUMN IF NOT EXISTS cover_preset_id UUID;
COMMENT ON COLUMN video_works.cover_media_id IS '封面媒体文件ID，关联 media.id；非空时 cover_url 留空';
COMMENT ON COLUMN video_works.cover_preset_id IS '封面预设ID，关联 media_presets.id；非空时优先按预设解析';
CREATE INDEX IF NOT EXISTS idx_video_works_cover_media_id ON video_works(cover_media_id);
CREATE INDEX IF NOT EXISTS idx_video_works_cover_preset_id ON video_works(cover_preset_id);

ALTER TABLE travel_guides ADD COLUMN IF NOT EXISTS cover_media_id UUID;
ALTER TABLE travel_guides ADD COLUMN IF NOT EXISTS cover_preset_id UUID;
COMMENT ON COLUMN travel_guides.cover_media_id IS '封面媒体文件ID，关联 media.id；非空时 cover_image 留空';
COMMENT ON COLUMN travel_guides.cover_preset_id IS '封面预设ID，关联 media_presets.id；非空时优先按预设解析';
CREATE INDEX IF NOT EXISTS idx_travel_guides_cover_media_id ON travel_guides(cover_media_id);
CREATE INDEX IF NOT EXISTS idx_travel_guides_cover_preset_id ON travel_guides(cover_preset_id);

ALTER TABLE recipes ADD COLUMN IF NOT EXISTS cover_media_id UUID;
ALTER TABLE recipes ADD COLUMN IF NOT EXISTS cover_preset_id UUID;
COMMENT ON COLUMN recipes.cover_media_id IS '封面媒体文件ID，关联 media.id；非空时 cover 留空';
COMMENT ON COLUMN recipes.cover_preset_id IS '封面预设ID，关联 media_presets.id；非空时优先按预设解析';
CREATE INDEX IF NOT EXISTS idx_recipes_cover_media_id ON recipes(cover_media_id);
CREATE INDEX IF NOT EXISTS idx_recipes_cover_preset_id ON recipes(cover_preset_id);

ALTER TABLE books ADD COLUMN IF NOT EXISTS cover_media_id UUID;
ALTER TABLE books ADD COLUMN IF NOT EXISTS cover_preset_id UUID;
COMMENT ON COLUMN books.cover_media_id IS '封面媒体文件ID，关联 media.id；非空时 cover 留空';
COMMENT ON COLUMN books.cover_preset_id IS '封面预设ID，关联 media_presets.id；非空时优先按预设解析';
CREATE INDEX IF NOT EXISTS idx_books_cover_media_id ON books(cover_media_id);
CREATE INDEX IF NOT EXISTS idx_books_cover_preset_id ON books(cover_preset_id);

ALTER TABLE games ADD COLUMN IF NOT EXISTS cover_media_id UUID;
ALTER TABLE games ADD COLUMN IF NOT EXISTS cover_preset_id UUID;
COMMENT ON COLUMN games.cover_media_id IS '封面媒体文件ID，关联 media.id；非空时 cover 留空';
COMMENT ON COLUMN games.cover_preset_id IS '封面预设ID，关联 media_presets.id；非空时优先按预设解析';
CREATE INDEX IF NOT EXISTS idx_games_cover_media_id ON games(cover_media_id);
CREATE INDEX IF NOT EXISTS idx_games_cover_preset_id ON games(cover_preset_id);

ALTER TABLE photo_equipment ADD COLUMN IF NOT EXISTS image_media_id UUID;
ALTER TABLE photo_equipment ADD COLUMN IF NOT EXISTS image_preset_id UUID;
COMMENT ON COLUMN photo_equipment.image_media_id IS '图片媒体文件ID，关联 media.id；非空时 image_url 留空';
COMMENT ON COLUMN photo_equipment.image_preset_id IS '图片预设ID，关联 media_presets.id；非空时优先按预设解析';
CREATE INDEX IF NOT EXISTS idx_photo_equipment_image_media_id ON photo_equipment(image_media_id);
CREATE INDEX IF NOT EXISTS idx_photo_equipment_image_preset_id ON photo_equipment(image_preset_id);

ALTER TABLE projects ADD COLUMN IF NOT EXISTS cover_media_id UUID;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS cover_preset_id UUID;
COMMENT ON COLUMN projects.cover_media_id IS '封面媒体文件ID，关联 media.id；非空时 cover_url 留空';
COMMENT ON COLUMN projects.cover_preset_id IS '封面预设ID，关联 media_presets.id；非空时优先按预设解析';
CREATE INDEX IF NOT EXISTS idx_projects_cover_media_id ON projects(cover_media_id);
CREATE INDEX IF NOT EXISTS idx_projects_cover_preset_id ON projects(cover_preset_id);

ALTER TABLE tech_stack_items ADD COLUMN IF NOT EXISTS icon_media_id UUID;
COMMENT ON COLUMN tech_stack_items.icon_media_id IS '图标媒体文件ID，关联 media.id；非空时 icon 留空（Simple Icons 等外链仍存 icon）';
CREATE INDEX IF NOT EXISTS idx_tech_stack_items_icon_media_id ON tech_stack_items(icon_media_id);

ALTER TABLE bloggers ADD COLUMN IF NOT EXISTS avatar_media_id UUID;
ALTER TABLE bloggers ADD COLUMN IF NOT EXISTS blog_icon_media_id UUID;
ALTER TABLE bloggers ADD COLUMN IF NOT EXISTS page_background_media_id UUID;
COMMENT ON COLUMN bloggers.avatar_media_id IS '头像媒体文件ID，关联 media.id；非空时 avatar 留空';
COMMENT ON COLUMN bloggers.blog_icon_media_id IS '博客icon媒体文件ID，关联 media.id；非空时 blog_icon 留空';
COMMENT ON COLUMN bloggers.page_background_media_id IS '页面背景媒体文件ID，关联 media.id；非空时 page_background 留空';
CREATE INDEX IF NOT EXISTS idx_bloggers_avatar_media_id ON bloggers(avatar_media_id);
CREATE INDEX IF NOT EXISTS idx_bloggers_blog_icon_media_id ON bloggers(blog_icon_media_id);
CREATE INDEX IF NOT EXISTS idx_bloggers_page_background_media_id ON bloggers(page_background_media_id);

-- ============================================================
-- 3. media.url / media_presets.output_url 归一为相对对象键
--    （完整 URL 读取时按当前存储配置动态拼接；storage_path 跨平台不变）
-- ============================================================
UPDATE media SET url = storage_path
WHERE storage_type <> 'local' AND storage_type IS NOT NULL AND storage_type <> ''
  AND storage_path IS NOT NULL AND storage_path <> ''
  AND url <> storage_path;

UPDATE media_presets p SET storage_type = m.storage_type
FROM media m
WHERE p.media_id = m.id AND (p.storage_type IS NULL OR p.storage_type = '');

UPDATE media_presets SET output_url = output_storage_path
WHERE output_storage_path IS NOT NULL AND output_storage_path <> ''
  AND output_url <> output_storage_path;

-- ============================================================
-- 4. 业务表媒体引用回填：按 URL 与 media.url / storage_path / preset output_storage_path
--    的包含关系匹配；匹配不上的（纯外链）保持 URL 模式。幂等：仅处理引用列为空的行。
-- ============================================================

-- 4.1 先按 media 匹配填 *_media_id
UPDATE articles a SET cover_media_id = m.id
FROM media m
WHERE a.cover_media_id IS NULL AND a.deleted_at IS NULL
  AND a.cover_image <> ''
  AND (a.cover_image = m.url OR a.cover_image LIKE '%' || m.storage_path || '%');

UPDATE songs s SET cover_media_id = m.id
FROM media m
WHERE s.cover_media_id IS NULL
  AND s.cover_url <> ''
  AND (s.cover_url = m.url OR s.cover_url LIKE '%' || m.storage_path || '%');

UPDATE third_party_playlists pl SET cover_media_id = m.id
FROM media m
WHERE pl.cover_media_id IS NULL
  AND pl.cover_url <> ''
  AND (pl.cover_url = m.url OR pl.cover_url LIKE '%' || m.storage_path || '%');

UPDATE video_works v SET cover_media_id = m.id
FROM media m
WHERE v.cover_media_id IS NULL AND v.deleted_at IS NULL
  AND v.cover_url <> ''
  AND (v.cover_url = m.url OR v.cover_url LIKE '%' || m.storage_path || '%');

UPDATE travel_guides t SET cover_media_id = m.id
FROM media m
WHERE t.cover_media_id IS NULL AND t.deleted_at IS NULL
  AND t.cover_image <> ''
  AND (t.cover_image = m.url OR t.cover_image LIKE '%' || m.storage_path || '%');

UPDATE recipes r SET cover_media_id = m.id
FROM media m
WHERE r.cover_media_id IS NULL AND r.deleted_at IS NULL
  AND r.cover <> ''
  AND (r.cover = m.url OR r.cover LIKE '%' || m.storage_path || '%');

UPDATE books b SET cover_media_id = m.id
FROM media m
WHERE b.cover_media_id IS NULL AND b.deleted_at IS NULL
  AND b.cover <> ''
  AND (b.cover = m.url OR b.cover LIKE '%' || m.storage_path || '%');

UPDATE games g SET cover_media_id = m.id
FROM media m
WHERE g.cover_media_id IS NULL AND g.deleted_at IS NULL
  AND g.cover <> ''
  AND (g.cover = m.url OR g.cover LIKE '%' || m.storage_path || '%');

UPDATE photo_equipment e SET image_media_id = m.id
FROM media m
WHERE e.image_media_id IS NULL AND e.deleted_at IS NULL
  AND e.image_url <> ''
  AND (e.image_url = m.url OR e.image_url LIKE '%' || m.storage_path || '%');

UPDATE projects pr SET cover_media_id = m.id
FROM media m
WHERE pr.cover_media_id IS NULL AND pr.deleted_at IS NULL
  AND pr.cover_url <> ''
  AND (pr.cover_url = m.url OR pr.cover_url LIKE '%' || m.storage_path || '%');

UPDATE tech_stack_items ts SET icon_media_id = m.id
FROM media m
WHERE ts.icon_media_id IS NULL AND ts.deleted_at IS NULL
  AND ts.icon <> ''
  AND (ts.icon = m.url OR ts.icon LIKE '%' || m.storage_path || '%');

UPDATE bloggers bl SET avatar_media_id = m.id
FROM media m
WHERE bl.avatar_media_id IS NULL AND bl.deleted_at IS NULL
  AND bl.avatar <> ''
  AND (bl.avatar = m.url OR bl.avatar LIKE '%' || m.storage_path || '%');

UPDATE bloggers bl SET blog_icon_media_id = m.id
FROM media m
WHERE bl.blog_icon_media_id IS NULL AND bl.deleted_at IS NULL
  AND bl.blog_icon <> ''
  AND (bl.blog_icon = m.url OR bl.blog_icon LIKE '%' || m.storage_path || '%');

UPDATE bloggers bl SET page_background_media_id = m.id
FROM media m
WHERE bl.page_background_media_id IS NULL AND bl.deleted_at IS NULL
  AND bl.page_background <> ''
  AND (bl.page_background = m.url OR bl.page_background LIKE '%' || m.storage_path || '%');

-- 4.2 再按预设成品图匹配填 *_preset_id（仅限已匹配到 media 的行，且预设属于该 media）
UPDATE articles a SET cover_preset_id = p.id
FROM media_presets p
WHERE a.cover_media_id IS NOT NULL AND a.cover_preset_id IS NULL AND a.deleted_at IS NULL
  AND p.media_id = a.cover_media_id
  AND a.cover_image LIKE '%' || p.output_storage_path || '%';

UPDATE songs s SET cover_preset_id = p.id
FROM media_presets p
WHERE s.cover_media_id IS NOT NULL AND s.cover_preset_id IS NULL
  AND p.media_id = s.cover_media_id
  AND s.cover_url LIKE '%' || p.output_storage_path || '%';

UPDATE third_party_playlists pl SET cover_preset_id = p.id
FROM media_presets p
WHERE pl.cover_media_id IS NOT NULL AND pl.cover_preset_id IS NULL
  AND p.media_id = pl.cover_media_id
  AND pl.cover_url LIKE '%' || p.output_storage_path || '%';

UPDATE video_works v SET cover_preset_id = p.id
FROM media_presets p
WHERE v.cover_media_id IS NOT NULL AND v.cover_preset_id IS NULL AND v.deleted_at IS NULL
  AND p.media_id = v.cover_media_id
  AND v.cover_url LIKE '%' || p.output_storage_path || '%';

UPDATE travel_guides t SET cover_preset_id = p.id
FROM media_presets p
WHERE t.cover_media_id IS NOT NULL AND t.cover_preset_id IS NULL AND t.deleted_at IS NULL
  AND p.media_id = t.cover_media_id
  AND t.cover_image LIKE '%' || p.output_storage_path || '%';

UPDATE recipes r SET cover_preset_id = p.id
FROM media_presets p
WHERE r.cover_media_id IS NOT NULL AND r.cover_preset_id IS NULL AND r.deleted_at IS NULL
  AND p.media_id = r.cover_media_id
  AND r.cover LIKE '%' || p.output_storage_path || '%';

UPDATE books b SET cover_preset_id = p.id
FROM media_presets p
WHERE b.cover_media_id IS NOT NULL AND b.cover_preset_id IS NULL AND b.deleted_at IS NULL
  AND p.media_id = b.cover_media_id
  AND b.cover LIKE '%' || p.output_storage_path || '%';

UPDATE games g SET cover_preset_id = p.id
FROM media_presets p
WHERE g.cover_media_id IS NOT NULL AND g.cover_preset_id IS NULL AND g.deleted_at IS NULL
  AND p.media_id = g.cover_media_id
  AND g.cover LIKE '%' || p.output_storage_path || '%';

UPDATE photo_equipment e SET image_preset_id = p.id, image_media_id = p.media_id
FROM media_presets p
WHERE e.image_media_id IS NULL AND e.image_preset_id IS NULL AND e.deleted_at IS NULL
  AND p.output_storage_path <> ''
  AND e.image_url LIKE '%' || p.output_storage_path || '%';

UPDATE projects pr SET cover_preset_id = p.id
FROM media_presets p
WHERE pr.cover_media_id IS NOT NULL AND pr.cover_preset_id IS NULL AND pr.deleted_at IS NULL
  AND p.media_id = pr.cover_media_id
  AND pr.cover_url LIKE '%' || p.output_storage_path || '%';
