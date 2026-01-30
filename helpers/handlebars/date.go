package handlebars

import (
	"fmt"
	"time"

	"github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/runtime"
)

// FormatDate formats a date using the specified format.
func FormatDate(args []any) (any, error) {
	arg := helpers.GetArg(args, 0)
	if arg == nil {
		return "", nil
	}
	
	var t time.Time
	switch v := arg.(type) {
	case time.Time:
		t = v
	case string:
		var err error
		t, err = helpers.ParseTime(v)
		if err != nil {
			return "", nil
		}
	case int64:
		t = time.Unix(v, 0)
	case float64:
		t = time.Unix(int64(v), 0)
	default:
		return "", nil
	}
	
	format := "2006-01-02"
	hash, _ := runtime.HashArg(args)
	if hash != nil {
		if f, ok := hash["format"].(string); ok {
			format = f
		}
	}
	
	// Convert Go time format to a more readable format if needed
	// For now, use the format as-is (Go time format)
	return t.Format(format), nil
}

// Now returns the current time.
func Now(args []any) (any, error) {
	format := time.RFC3339
	hash, _ := runtime.HashArg(args)
	if hash != nil {
		if f, ok := hash["format"].(string); ok {
			format = f
		}
	}
	return time.Now().Format(format), nil
}

// Ago returns a human-readable time ago string.
func Ago(args []any) (any, error) {
	arg := helpers.GetArg(args, 0)
	if arg == nil {
		return "", nil
	}
	
	var t time.Time
	switch v := arg.(type) {
	case time.Time:
		t = v
	case string:
		var err error
		t, err = helpers.ParseTime(v)
		if err != nil {
			return "", nil
		}
	case int64:
		t = time.Unix(v, 0)
	case float64:
		t = time.Unix(int64(v), 0)
	default:
		return "", nil
	}
	
	duration := time.Since(t)
	if duration < time.Minute {
		return "just now", nil
	}
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago", nil
		}
		return fmt.Sprintf("%d minutes ago", minutes), nil
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago", nil
		}
		return fmt.Sprintf("%d hours ago", hours), nil
	}
	if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago", nil
		}
		return fmt.Sprintf("%d days ago", days), nil
	}
	if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago", nil
		}
		return fmt.Sprintf("%d months ago", months), nil
	}
	years := int(duration.Hours() / (24 * 365))
	if years == 1 {
		return "1 year ago", nil
	}
	return fmt.Sprintf("%d years ago", years), nil
}

