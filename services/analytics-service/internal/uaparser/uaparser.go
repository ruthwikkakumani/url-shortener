package uaparser

import (
	ua "github.com/mileusna/useragent"
)

type ParsedUA struct {
	DeviceType string // desktop | mobile | tablet
	OS         string
	Browser    string
}

// Parse extracts device type, OS, and browser from a User-Agent string.
func Parse(userAgent string) ParsedUA {
	parsed := ua.Parse(userAgent)

	deviceType := "desktop"
	if parsed.Mobile {
		deviceType = "mobile"
	} else if parsed.Tablet {
		deviceType = "tablet"
	}

	os := normalise(parsed.OS)
	browser := normalise(parsed.Name)

	return ParsedUA{
		DeviceType: deviceType,
		OS:         os,
		Browser:    browser,
	}
}

func normalise(s string) string {
	if s == "" {
		return "Other"
	}
	return s
}
