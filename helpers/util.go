package helpers

import (
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/andriyg76/go-hbars/runtime"
	"github.com/andriyg76/hexerr"
)

// GetArg returns the argument at index, or nil if out of bounds.
func GetArg(args []any, idx int) any {
	if idx < 0 || idx >= len(args) {
		return nil
	}
	return args[idx]
}

// GetStringArg returns the stringified argument at index, or empty string.
func GetStringArg(args []any, idx int) string {
	arg := GetArg(args, idx)
	if arg == nil {
		return ""
	}
	return runtime.Stringify(arg)
}

// GetNumberArg attempts to convert the argument at index to a float64.
func GetNumberArg(args []any, idx int) (float64, error) {
	arg := GetArg(args, idx)
	if arg == nil {
		return 0, nil
	}
	switch v := arg.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, err
		}
		return f, nil
	default:
		return 0, hexerr.Newf("cannot convert %T to number", arg)
	}
}

// IsEmpty checks if a value is empty (nil, empty string, empty collection, zero number).
// Returns (true, nil) for empty, (false, nil) for non-empty, or (false, error) when reflect is used and ReflectUsageLevel=ERROR.
func IsEmpty(v any) (bool, error) {
	if v == nil {
		return true, nil
	}
	switch t := v.(type) {
	case string:
		return t == "", nil
	case []any:
		return len(t) == 0, nil
	case []byte:
		return len(t) == 0, nil
	case map[string]any:
		return len(t) == 0, nil
	case map[any]any:
		return len(t) == 0, nil
	case int:
		return t == 0, nil
	case int8:
		return t == 0, nil
	case int16:
		return t == 0, nil
	case int32:
		return t == 0, nil
	case int64:
		return t == 0, nil
	case uint:
		return t == 0, nil
	case uint8:
		return t == 0, nil
	case uint16:
		return t == 0, nil
	case uint32:
		return t == 0, nil
	case uint64:
		return t == 0, nil
	case uintptr:
		return t == 0, nil
	case float32:
		return t == 0, nil
	case float64:
		return t == 0, nil
	}
	// Reflect path
	level := runtime.GetReflectUsageLevel()
	if level == runtime.ReflectError {
		return false, hexerr.Newf("reflect usage disabled by ReflectUsageLevel=ERROR (IsEmpty for type %T)", v)
	}
	if level == runtime.ReflectWarn {
		// log is in stdlib
		reflectWarn("IsEmpty", v)
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len() == 0, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0, nil
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0, nil
	}
	return false, nil
}

func reflectWarn(funcName string, v any) {
	log.Printf("[go-hbars] reflect used in %s for type %T", funcName, v)
}

// ParseTime attempts to parse a time string using common formats.
func ParseTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
	}
	for _, fmt := range formats {
		if t, err := time.Parse(fmt, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, hexerr.Newf("unable to parse time: %q", s)
}
