-- 媒体预设表迁移

CREATE TABLE IF NOT EXISTS media_presets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    media_id UUID NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    frame_config JSONB NOT NULL DEFAULT '{}',
    display_params JSONB NOT NULL DEFAULT '{}',
    output_url VARCHAR(500) NOT NULL,
    output_storage_path VARCHAR(500) NOT NULL,
    output_size BIGINT NOT NULL DEFAULT 0,
    mime_type VARCHAR(100) NOT NULL DEFAULT 'image/jpeg',
    created_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_media_presets_media_id ON media_presets(media_id);
CREATE INDEX IF NOT EXISTS idx_media_presets_deleted_at ON media_presets(deleted_at);

COMMENT ON TABLE media_presets IS '媒体预设表，存储基于原图生成的成品图及其配置';
COMMENT ON COLUMN media_presets.id IS '预设唯一标识';
COMMENT ON COLUMN media_presets.media_id IS '关联的原图媒体ID';
COMMENT ON COLUMN media_presets.name IS '预设名称';
COMMENT ON COLUMN media_presets.frame_config IS '相框样式配置（JSON）';
COMMENT ON COLUMN media_presets.display_params IS 'EXIF显示参数（JSON）';
COMMENT ON COLUMN media_presets.output_url IS '成品图访问URL';
COMMENT ON COLUMN media_presets.output_storage_path IS '成品图对象存储路径';
COMMENT ON COLUMN media_presets.output_size IS '成品图文件大小（字节）';
COMMENT ON COLUMN media_presets.mime_type IS '成品图MIME类型';
COMMENT ON COLUMN media_presets.created_at IS '创建时间';
COMMENT ON COLUMN media_presets.deleted_at IS '软删除时间';
