CREATE TABLE IF NOT EXISTS travel_guides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    summary VARCHAR(500),
    cover_image VARCHAR(500),
    status SMALLINT DEFAULT 1,
    destination VARCHAR(200) NOT NULL,
    region VARCHAR(50) NOT NULL,
    days INT DEFAULT 1,
    best_month VARCHAR(100),
    view_count INT DEFAULT 0,
    like_count INT DEFAULT 0,
    rating DECIMAL(2,1) DEFAULT 0.0,
    review_count INT DEFAULT 0,
    attractions JSONB DEFAULT '[]',
    itinerary JSONB DEFAULT '[]',
    reviews JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_travel_guides_deleted_at ON travel_guides(deleted_at);
CREATE INDEX IF NOT EXISTS idx_travel_guides_status ON travel_guides(status);
CREATE INDEX IF NOT EXISTS idx_travel_guides_region ON travel_guides(region);
