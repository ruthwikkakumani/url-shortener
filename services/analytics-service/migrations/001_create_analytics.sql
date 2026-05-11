-- Analytics click events table
CREATE TABLE IF NOT EXISTS click_events (
    id            BIGSERIAL PRIMARY KEY,
    short_code    VARCHAR(20)   NOT NULL,
    original_url  TEXT,
    ip_address    TEXT,
    country       VARCHAR(100),
    country_code  CHAR(2),
    city          VARCHAR(100),
    device_type   VARCHAR(20),   -- desktop | mobile | tablet
    os            VARCHAR(50),   -- Windows | macOS | Linux | Android | iOS | Other
    browser       VARCHAR(50),   -- Chrome | Firefox | Safari | Edge | Other
    referer       TEXT,
    clicked_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_click_events_short_code  ON click_events (short_code);
CREATE INDEX IF NOT EXISTS idx_click_events_clicked_at  ON click_events (clicked_at);
CREATE INDEX IF NOT EXISTS idx_click_events_country     ON click_events (country);
CREATE INDEX IF NOT EXISTS idx_click_events_device_type ON click_events (device_type);
CREATE INDEX IF NOT EXISTS idx_click_events_os          ON click_events (os);
CREATE INDEX IF NOT EXISTS idx_click_events_browser     ON click_events (browser);
-- Composite for time-series queries per code
CREATE INDEX IF NOT EXISTS idx_click_events_code_time   ON click_events (short_code, clicked_at);
