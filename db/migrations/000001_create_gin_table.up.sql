CREATE TABLE IF NOT EXISTS gin (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    country VARCHAR(255) NOT NULL,
    botanicals TEXT[] DEFAULT ARRAY[]::TEXT[],
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_gin_name ON gin (LOWER(name));
CREATE INDEX IF NOT EXISTS idx_gin_country ON gin (LOWER(country));
CREATE INDEX IF NOT EXISTS idx_gin_created_at ON gin (created_at);
