package runtime

import (
	"encoding/json"
	"log"
	"reflect"

	"github.com/andriyg76/hexerr"
)

// IsNumericZero reports whether v is a numeric type with value zero.
// It handles json.Number (from decoding JSON with UseNumber()), Go numeric types,
// and numeric kinds via reflection when ReflectUsageLevel allows.
func IsNumericZero(v any) (bool, error) {
	if v == nil {
		return false, nil
	}
	switch n := v.(type) {
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return i == 0, nil
		}
		if f, err := n.Float64(); err == nil {
			return f == 0, nil
		}
		return false, nil
	}
	// Reflect path
	level := GetReflectUsageLevel()
	if level == ReflectError {
		return false, hexerr.New("reflect usage disabled by ReflectUsageLevel=ERROR (IsNumericZero)")
	}
	if level == ReflectWarn {
		log.Printf("[go-hbars] reflect used in IsNumericZero for type %T", v)
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Interface || rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return false, nil
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0, nil
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0, nil
	default:
		return false, nil
	}
}

// IncludeZeroTruthy reports whether v should be treated as truthy when
// the includeZero option is set: it returns true if v is numeric zero
// or if IsTruthy(v) is true. Used by {{#if}} / {{#unless}} with includeZero=true.
func IncludeZeroTruthy(v any) (bool, error) {
	zero, err := IsNumericZero(v)
	if err != nil {
		return false, err
	}
	if zero {
		return true, nil
	}
	return IsTruthy(v)
}

// IsTruthy reports whether a value should be treated as true in block helpers.
func IsTruthy(value any) (bool, error) {
	if value == nil {
		return false, nil
	}
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return v != "", nil
	case []byte:
		return len(v) > 0, nil
	case []any:
		return len(v) > 0, nil
	case map[string]any:
		return len(v) > 0, nil
	case map[any]any:
		return len(v) > 0, nil
	case SafeString:
		return v != "", nil
	case int:
		return v != 0, nil
	case int8:
		return v != 0, nil
	case int16:
		return v != 0, nil
	case int32:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case uint:
		return v != 0, nil
	case uint8:
		return v != 0, nil
	case uint16:
		return v != 0, nil
	case uint32:
		return v != 0, nil
	case uint64:
		return v != 0, nil
	case uintptr:
		return v != 0, nil
	case float32:
		return v != 0, nil
	case float64:
		return v != 0, nil
	}

	// Reflect path for unknown types
	level := GetReflectUsageLevel()
	if level == ReflectError {
		return false, hexerr.Newf("reflect usage disabled by ReflectUsageLevel=ERROR (IsTruthy for type %T)", value)
	}
	if level == ReflectWarn {
		log.Printf("[go-hbars] reflect used in IsTruthy for type %T", value)
	}
	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Interface || rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return false, nil
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool(), nil
	case reflect.String:
		return rv.Len() > 0, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() != 0, nil
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0, nil
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() > 0, nil
	default:
		return true, nil
	}
}
