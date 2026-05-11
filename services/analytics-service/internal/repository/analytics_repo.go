package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/models"
	"go.uber.org/zap"
)

// ClicksOverTime is one bucket (hour / day) with a click count.
type ClicksOverTime struct {
	Bucket time.Time `json:"bucket"`
	Clicks int64     `json:"clicks"`
}

// KV is a generic key-value pair used for breakdown queries.
type KV struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// PeakHour is an hour (0–23) and its total click count.
type PeakHour struct {
	Hour   int   `json:"hour"`
	Clicks int64 `json:"clicks"`
}

// RecentClick is a single row for the "recent events" endpoint.
type RecentClick struct {
	ClickedAt  time.Time `json:"clicked_at"`
	IPAddress  string    `json:"ip_address"`
	Country    string    `json:"country"`
	City       string    `json:"city"`
	DeviceType string    `json:"device_type"`
	OS         string    `json:"os"`
	Browser    string    `json:"browser"`
}

type AnalyticsRepo struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewAnalyticsRepo(pool *pgxpool.Pool, logger *zap.Logger) *AnalyticsRepo {
	return &AnalyticsRepo{pool: pool, logger: logger}
}

// InsertClickBatch persists multiple events and updates aggregation tables using pgx.Batch.
func (r *AnalyticsRepo) InsertClickBatch(ctx context.Context, events []models.EnrichedClickEvent) error {
	if len(events) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	for _, e := range events {
		// 1. Insert Raw Event
		batch.Queue(`
			INSERT INTO click_events
				(short_code, original_url, ip_address, country, country_code, city,
				 device_type, os, browser, referer, clicked_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			e.ShortCode, e.OriginalURL, e.IPAddress,
			e.Country, e.CountryCode, e.City,
			e.DeviceType, e.OS, e.Browser,
			e.Referer, e.ClickedAt,
		)

		// 2. Hourly Aggregation
		hourBucket := e.ClickedAt.Truncate(time.Hour)
		batch.Queue(`
			INSERT INTO agg_hourly_clicks (short_code, bucket_hour, clicks)
			VALUES ($1, $2, 1)
			ON CONFLICT (short_code, bucket_hour)
			DO UPDATE SET clicks = agg_hourly_clicks.clicks + 1`,
			e.ShortCode, hourBucket,
		)

		// 3. Country Aggregation
		country := e.Country
		if country == "" {
			country = "Unknown"
		}
		batch.Queue(`
			INSERT INTO agg_country_clicks (short_code, country, clicks)
			VALUES ($1, $2, 1)
			ON CONFLICT (short_code, country)
			DO UPDATE SET clicks = agg_country_clicks.clicks + 1`,
			e.ShortCode, country,
		)

		// 4. City Aggregation
		city := e.City
		if city == "" {
			city = "Unknown"
		}
		batch.Queue(`
			INSERT INTO agg_city_clicks (short_code, city, clicks)
			VALUES ($1, $2, 1)
			ON CONFLICT (short_code, city)
			DO UPDATE SET clicks = agg_city_clicks.clicks + 1`,
			e.ShortCode, city,
		)

		// 5. Device Aggregation
		device := e.DeviceType
		if device == "" {
			device = "Unknown"
		}
		batch.Queue(`
			INSERT INTO agg_device_clicks (short_code, device_type, clicks)
			VALUES ($1, $2, 1)
			ON CONFLICT (short_code, device_type)
			DO UPDATE SET clicks = agg_device_clicks.clicks + 1`,
			e.ShortCode, device,
		)

		// 6. OS Aggregation
		os := e.OS
		if os == "" {
			os = "Unknown"
		}
		batch.Queue(`
			INSERT INTO agg_os_clicks (short_code, os, clicks)
			VALUES ($1, $2, 1)
			ON CONFLICT (short_code, os)
			DO UPDATE SET clicks = agg_os_clicks.clicks + 1`,
			e.ShortCode, os,
		)

		// 7. Browser Aggregation
		browser := e.Browser
		if browser == "" {
			browser = "Unknown"
		}
		batch.Queue(`
			INSERT INTO agg_browser_clicks (short_code, browser, clicks)
			VALUES ($1, $2, 1)
			ON CONFLICT (short_code, browser)
			DO UPDATE SET clicks = agg_browser_clicks.clicks + 1`,
			e.ShortCode, browser,
		)

		// 8. Unique IPs
		if e.IPAddress != "" {
			batch.Queue(`
				INSERT INTO agg_unique_ips (short_code, ip_address)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING`,
				e.ShortCode, e.IPAddress,
			)
		}
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

// TotalClicks returns the total number of clicks for a short code.
func (r *AnalyticsRepo) TotalClicks(ctx context.Context, code string) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(clicks), 0)::BIGINT FROM agg_hourly_clicks WHERE short_code=$1`, code,
	).Scan(&count)
	return count, err
}

// UniqueIPs returns the number of distinct IP addresses.
func (r *AnalyticsRepo) UniqueIPs(ctx context.Context, code string) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM agg_unique_ips WHERE short_code=$1`, code,
	).Scan(&count)
	return count, err
}

// ClicksOverTimeHourly returns hourly click buckets.
func (r *AnalyticsRepo) ClicksOverTimeHourly(ctx context.Context, code string, since time.Time) ([]ClicksOverTime, error) {
	query := `
		SELECT bucket_hour AS bucket, COALESCE(SUM(clicks), 0)::BIGINT AS clicks
		FROM agg_hourly_clicks
		WHERE short_code=$1 AND bucket_hour >= $2
		GROUP BY bucket_hour
		ORDER BY bucket_hour`

	return r.scanTimeSeries(ctx, query, code, since)
}

// ClicksOverTimeDaily returns daily click buckets.
func (r *AnalyticsRepo) ClicksOverTimeDaily(ctx context.Context, code string, since time.Time) ([]ClicksOverTime, error) {
	query := `
		SELECT date_trunc('day', bucket_hour) AS bucket, COALESCE(SUM(clicks), 0)::BIGINT AS clicks
		FROM agg_hourly_clicks
		WHERE short_code=$1 AND bucket_hour >= $2
		GROUP BY bucket
		ORDER BY bucket`

	return r.scanTimeSeries(ctx, query, code, since)
}

func (r *AnalyticsRepo) scanTimeSeries(ctx context.Context, query, code string, since time.Time) ([]ClicksOverTime, error) {
	rows, err := r.pool.Query(ctx, query, code, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ClicksOverTime
	for rows.Next() {
		var row ClicksOverTime
		if err := rows.Scan(&row.Bucket, &row.Clicks); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ByCountry returns click counts grouped by country.
func (r *AnalyticsRepo) ByCountry(ctx context.Context, code string) ([]KV, error) {
	return r.breakdown(ctx, code, "agg_country_clicks", "country")
}

// ByCity returns click counts grouped by city.
func (r *AnalyticsRepo) ByCity(ctx context.Context, code string) ([]KV, error) {
	return r.breakdown(ctx, code, "agg_city_clicks", "city")
}

// ByDevice returns click counts grouped by device_type.
func (r *AnalyticsRepo) ByDevice(ctx context.Context, code string) ([]KV, error) {
	return r.breakdown(ctx, code, "agg_device_clicks", "device_type")
}

// ByOS returns click counts grouped by os.
func (r *AnalyticsRepo) ByOS(ctx context.Context, code string) ([]KV, error) {
	return r.breakdown(ctx, code, "agg_os_clicks", "os")
}

// ByBrowser returns click counts grouped by browser.
func (r *AnalyticsRepo) ByBrowser(ctx context.Context, code string) ([]KV, error) {
	return r.breakdown(ctx, code, "agg_browser_clicks", "browser")
}

func (r *AnalyticsRepo) breakdown(ctx context.Context, code, table, column string) ([]KV, error) {
	query := `
		SELECT ` + column + ` AS label, clicks AS count
		FROM ` + table + `
		WHERE short_code=$1
		ORDER BY count DESC`

	rows, err := r.pool.Query(ctx, query, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []KV
	for rows.Next() {
		var kv KV
		if err := rows.Scan(&kv.Label, &kv.Count); err != nil {
			return nil, err
		}
		result = append(result, kv)
	}
	return result, rows.Err()
}

// PeakHours returns total clicks grouped by hour-of-day (0–23) across all time.
func (r *AnalyticsRepo) PeakHours(ctx context.Context, code string) ([]PeakHour, error) {
	query := `
		SELECT EXTRACT(HOUR FROM bucket_hour)::INT AS hour, COALESCE(SUM(clicks), 0)::BIGINT AS clicks
		FROM agg_hourly_clicks
		WHERE short_code=$1
		GROUP BY hour
		ORDER BY hour`

	rows, err := r.pool.Query(ctx, query, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PeakHour
	for rows.Next() {
		var ph PeakHour
		if err := rows.Scan(&ph.Hour, &ph.Clicks); err != nil {
			return nil, err
		}
		result = append(result, ph)
	}
	return result, rows.Err()
}

// RecentClicks returns the last N click events.
func (r *AnalyticsRepo) RecentClicks(ctx context.Context, code string, limit int) ([]RecentClick, error) {
	query := `
		SELECT clicked_at, COALESCE(ip_address,''), COALESCE(country,''),
		       COALESCE(city,''), COALESCE(device_type,''), COALESCE(os,''), COALESCE(browser,'')
		FROM click_events
		WHERE short_code=$1
		ORDER BY clicked_at DESC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, query, code, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []RecentClick
	for rows.Next() {
		var rc RecentClick
		if err := rows.Scan(&rc.ClickedAt, &rc.IPAddress, &rc.Country, &rc.City, &rc.DeviceType, &rc.OS, &rc.Browser); err != nil {
			return nil, err
		}
		result = append(result, rc)
	}
	return result, rows.Err()
}
