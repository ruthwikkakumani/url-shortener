// Package geo resolves an IP address to country/city using the free ip-api.com service.
// Private / loopback IPs are returned with empty geo fields.
package geo

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

var (
	httpClient = &http.Client{Timeout: 3 * time.Second}
	ipCache    *lru.Cache[string, GeoInfo] // Production-grade LRU cache
)

func init() {
	var err error
	// Set a hard limit of 10,000 unique IPs. 
	// If it hits the limit, it safely deletes the oldest, least-used IP.
	ipCache, err = lru.New[string, GeoInfo](10000)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize IP LRU cache: %v", err))
	}
}

type GeoInfo struct {
	Country     string
	CountryCode string
	City        string
}

type ipAPIResponse struct {
	Status      string `json:"status"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	City        string `json:"city"`
}

// isPrivate returns true for RFC-1918 / loopback addresses.
func isPrivate(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return true
	}
	if parsed.IsLoopback() || parsed.IsLinkLocalUnicast() || parsed.IsPrivate() {
		return true
	}
	return false
}

// Lookup resolves an IP to geo info. Returns empty GeoInfo for private/invalid IPs.
func Lookup(ip string) GeoInfo {
	if isPrivate(ip) {
		return GeoInfo{}
	}

	// 1. Check LRU cache first
	if val, ok := ipCache.Get(ip); ok {
		return val
	}

	// 2. Not in cache, perform HTTP lookup
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,countryCode,city", ip)
	resp, err := httpClient.Get(url)
	if err != nil {
		return GeoInfo{}
	}
	defer resp.Body.Close()

	var data ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return GeoInfo{}
	}

	if data.Status != "success" {
		// Cache the empty result so we don't spam invalid IPs and hit rate limits
		ipCache.Add(ip, GeoInfo{})
		return GeoInfo{}
	}

	info := GeoInfo{
		Country:     data.Country,
		CountryCode: data.CountryCode,
		City:        data.City,
	}

	// 3. Store successful result in LRU cache
	ipCache.Add(ip, info)

	return info
}
