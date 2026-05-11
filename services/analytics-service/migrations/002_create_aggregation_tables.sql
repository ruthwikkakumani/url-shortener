-- Aggregation Tables for lightning-fast reads
CREATE TABLE IF NOT EXISTS agg_hourly_clicks (
    short_code VARCHAR(20) NOT NULL,
    bucket_hour TIMESTAMPTZ NOT NULL,
    clicks BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (short_code, bucket_hour)
);

CREATE TABLE IF NOT EXISTS agg_country_clicks (
    short_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL,
    clicks BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (short_code, country)
);

CREATE TABLE IF NOT EXISTS agg_city_clicks (
    short_code VARCHAR(20) NOT NULL,
    city VARCHAR(100) NOT NULL,
    clicks BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (short_code, city)
);

CREATE TABLE IF NOT EXISTS agg_device_clicks (
    short_code VARCHAR(20) NOT NULL,
    device_type VARCHAR(20) NOT NULL,
    clicks BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (short_code, device_type)
);

CREATE TABLE IF NOT EXISTS agg_os_clicks (
    short_code VARCHAR(20) NOT NULL,
    os VARCHAR(50) NOT NULL,
    clicks BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (short_code, os)
);

CREATE TABLE IF NOT EXISTS agg_browser_clicks (
    short_code VARCHAR(20) NOT NULL,
    browser VARCHAR(50) NOT NULL,
    clicks BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (short_code, browser)
);

-- For fast Unique IPs counting without doing DISTINCT over millions of rows
CREATE TABLE IF NOT EXISTS agg_unique_ips (
    short_code VARCHAR(20) NOT NULL,
    ip_address TEXT NOT NULL,
    PRIMARY KEY (short_code, ip_address)
);
