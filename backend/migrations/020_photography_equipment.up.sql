-- 摄影器材表迁移

CREATE TABLE IF NOT EXISTS photo_equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    image_url VARCHAR(1024),
    brand VARCHAR(255),
    description TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_photo_equipment_deleted_at ON photo_equipment(deleted_at);
CREATE INDEX IF NOT EXISTS idx_photo_equipment_sort_order ON photo_equipment(sort_order);

COMMENT ON TABLE photo_equipment IS '摄影器材表，存储摄影器材信息';
COMMENT ON COLUMN photo_equipment.id IS '器材唯一标识';
COMMENT ON COLUMN photo_equipment.name IS '器材名称';
COMMENT ON COLUMN photo_equipment.image_url IS '器材图片地址';
COMMENT ON COLUMN photo_equipment.brand IS '器材品牌';
COMMENT ON COLUMN photo_equipment.description IS '器材介绍';
COMMENT ON COLUMN photo_equipment.sort_order IS '排序权重';
COMMENT ON COLUMN photo_equipment.created_at IS '创建时间';
COMMENT ON COLUMN photo_equipment.updated_at IS '更新时间';
COMMENT ON COLUMN photo_equipment.deleted_at IS '软删除时间';
