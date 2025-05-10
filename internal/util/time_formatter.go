package util

import (
	"sync"
	"time"
)

const (
	// RFC3339Nano is the most precise and widely supported time format
	DefaultTimeFormat = time.RFC3339Nano
)

var (
	formatter     *TimeFormatter
	formatterOnce sync.Once
)

// TimeFormatter is a singleton that handles time formatting
type TimeFormatter struct {
	location *time.Location
}

func NewTimeFormatter() *TimeFormatter {
	formatterOnce.Do(func() {
		loc, err := time.LoadLocation("UTC")
		if err != nil {
			// Fallback to local time if UTC is not available
			loc = time.Local
		}
		formatter = &TimeFormatter{
			location: loc,
		}
	})
	return formatter
}

// FormatTime formats a time.Time to string using the DefaultTimeFormat
func (f *TimeFormatter) FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(f.location).Format(DefaultTimeFormat)
}

// ParseTime parses a string to time.Time using the DefaultTimeFormat
func (f *TimeFormatter) ParseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.ParseInLocation(DefaultTimeFormat, s, f.location)
}
