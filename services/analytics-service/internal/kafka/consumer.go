package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"os"
	"time"

	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/geo"
	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/models"
	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/repository"
	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/uaparser"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"go.uber.org/zap"
)

const (
	TopicClickEvents = "click-events"
	ConsumerGroupID  = "analytics-service"
)

type rawClickEvent struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	Referer     string    `json:"referer"`
	ClickedAt   time.Time `json:"clicked_at"`
}

type Consumer struct {
	client *kgo.Client
	repo   *repository.AnalyticsRepo
	logger *zap.Logger
}

func NewConsumer(brokers []string, repo *repository.AnalyticsRepo, logger *zap.Logger) (*Consumer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(ConsumerGroupID),
		kgo.ConsumeTopics(TopicClickEvents),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		kgo.DisableAutoCommit(),
		kgo.FetchMaxWait(500 * time.Millisecond),
	}

	user := os.Getenv("KAFKA_USERNAME")
	pass := os.Getenv("KAFKA_PASSWORD")

	if user != "" && pass != "" {
		// Redpanda Cloud / Upstash require SASL SCRAM-SHA-256 over TLS
		opts = append(opts,
			kgo.DialTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}),
			kgo.SASL(scram.Auth{
				User: user,
				Pass: pass,
			}.AsSha256Mechanism()),
		)
		logger.Info("kafka consumer: SASL_SSL (SCRAM-SHA-256) enabled")
	} else {
		logger.Info("kafka consumer: plain-text mode (no SASL)")
	}

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		logger.Error("kafka consumer: failed to create client", zap.Error(err))
		return nil, err
	}
	return &Consumer{client: cl, repo: repo, logger: logger}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	c.logger.Info("kafka consumer: starting", zap.String("topic", TopicClickEvents))

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("kafka consumer: shutting down")
			c.client.Close()
			return
		default:
		}

		fetches := c.client.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, e := range errs {
				c.logger.Error("kafka fetch error: " + e.Err.Error(),
					zap.String("topic", e.Topic),
					zap.Int32("partition", e.Partition),
				)
			}
		}

		var events []models.EnrichedClickEvent

		fetches.EachRecord(func(r *kgo.Record) {
			var raw rawClickEvent
			if err := json.Unmarshal(r.Value, &raw); err != nil {
				c.logger.Error("kafka: failed to unmarshal click event", zap.Error(err))
				return
			}

			parsed := uaparser.Parse(raw.UserAgent)
			geoInfo := geo.Lookup(raw.IP)

			events = append(events, models.EnrichedClickEvent{
				ShortCode:   raw.ShortCode,
				OriginalURL: raw.OriginalURL,
				IPAddress:   raw.IP,
				Country:     geoInfo.Country,
				CountryCode: geoInfo.CountryCode,
				City:        geoInfo.City,
				DeviceType:  parsed.DeviceType,
				OS:          parsed.OS,
				Browser:     parsed.Browser,
				Referer:     raw.Referer,
				ClickedAt:   raw.ClickedAt,
			})
		})

		if len(events) > 0 {
			if err := c.repo.InsertClickBatch(ctx, events); err != nil {
				c.logger.Error("kafka: failed to persist batch of click events", zap.Error(err), zap.Int("batch_size", len(events)))
			} else {
				c.logger.Info("kafka: successfully processed batch", zap.Int("batch_size", len(events)))
			}
		}

		if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
			c.logger.Warn("kafka commit error", zap.Error(err))
		}
	}
}


