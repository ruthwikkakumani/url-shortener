package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	ips map[string]*ipEntry
	mu  sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	rl := &IPRateLimiter{
		ips: make(map[string]*ipEntry),
		r:   r,
		b:   b,
	}
	go rl.cleanupLoop()
	return rl
}

func (i *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-10 * time.Minute)
		i.mu.Lock()
		for ip, entry := range i.ips {
			if entry.lastSeen.Before(cutoff) {
				delete(i.ips, ip)
			}
		}
		i.mu.Unlock()
	}
}

func (i *IPRateLimiter) addIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()
	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = &ipEntry{limiter: limiter, lastSeen: time.Now()}
	return limiter
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	entry, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		return i.addIP(ip)
	}

	i.mu.Lock()
	entry.lastSeen = time.Now()
	i.mu.Unlock()

	return entry.limiter
}

func RateLimitMiddleware(requestsPerSecond float64, burstSize int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rate.Limit(requestsPerSecond), burstSize)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.GetLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			return
		}
		c.Next()
	}
}
