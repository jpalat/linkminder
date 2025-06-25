package database

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"
)

// TagsToJSON converts a slice of tags to JSON string
func TagsToJSON(tags []string) string {
	if tags == nil || len(tags) == 0 {
		return "[]"
	}
	jsonData, err := json.Marshal(tags)
	if err != nil {
		return "[]"
	}
	return string(jsonData)
}

// TagsFromJSON converts a JSON string to slice of tags
func TagsFromJSON(jsonStr string) []string {
	if jsonStr == "" {
		return []string{}
	}
	var tags []string
	err := json.Unmarshal([]byte(jsonStr), &tags)
	if err != nil {
		return []string{}
	}
	return tags
}

// CustomPropsToJSON converts a map of custom properties to JSON string
func CustomPropsToJSON(props map[string]string) string {
	if props == nil || len(props) == 0 {
		return "{}"
	}
	jsonData, err := json.Marshal(props)
	if err != nil {
		return "{}"
	}
	return string(jsonData)
}

// CustomPropsFromJSON converts a JSON string to map of custom properties
func CustomPropsFromJSON(jsonStr string) map[string]string {
	if jsonStr == "" {
		return map[string]string{}
	}
	var props map[string]string
	err := json.Unmarshal([]byte(jsonStr), &props)
	if err != nil {
		return map[string]string{}
	}
	return props
}

// ExtractDomain extracts domain from URL string
func ExtractDomain(urlStr string) string {
	if urlStr == "" {
		return ""
	}
	
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	
	return parsedURL.Hostname()
}

// CalculateAge calculates the age string from timestamp
func CalculateAge(timestamp string) string {
	// Parse the timestamp (assuming SQLite datetime format)
	t, err := time.Parse("2006-01-02 15:04:05", timestamp)
	if err != nil {
		return "unknown"
	}
	
	duration := time.Since(t)
	
	if duration.Hours() < 1 {
		return "now"
	} else if duration.Hours() < 24 {
		return time.Since(t).Truncate(time.Hour).String()
	} else if duration.Hours() < 24*7 {
		days := int(duration.Hours() / 24)
		return time.Duration(days*24) * time.Hour.String()
	} else {
		weeks := int(duration.Hours() / (24 * 7))
		return time.Duration(weeks*24*7) * time.Hour.String()
	}
}

// GetSuggestedAction suggests an action based on domain, title, and description
func GetSuggestedAction(domain, title, description string) string {
	domain = strings.ToLower(domain)
	title = strings.ToLower(title)
	description = strings.ToLower(description)
	
	// Documentation sites
	if strings.Contains(domain, "doc") || strings.Contains(domain, "manual") ||
		strings.Contains(title, "documentation") || strings.Contains(title, "docs") {
		return "working"
	}
	
	// Tutorial or learning content
	if strings.Contains(title, "tutorial") || strings.Contains(title, "guide") ||
		strings.Contains(title, "how to") || strings.Contains(title, "learn") {
		return "working"
	}
	
	// News or blog posts
	if strings.Contains(domain, "blog") || strings.Contains(domain, "news") ||
		strings.Contains(title, "announce") {
		return "read-later"
	}
	
	// Default suggestion
	return "read-later"
}