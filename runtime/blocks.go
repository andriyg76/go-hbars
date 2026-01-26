package runtime

import (
	"reflect"
	"sort"
)

// IsTruthy reports whether a value should be treated as true in block helpers.
func IsTruthy(value any) bool {
	if value == nil {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case []byte:
		return len(v) > 0
	case SafeString:
		return v != ""
	case int:
		return v != 0
	case int8:
		return v != 0
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint16:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	case uintptr:
		return v != 0
	case float32:
		return v != 0
	case float64:
		return v != 0
	}

	v := reflect.ValueOf(value)
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.String:
		return v.Len() > 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() > 0
	default:
		return true
	}
}

// IterItem represents a single {{#each}} iteration value.
type IterItem struct {
	Value any
	Key   string
	Index int
}

// Iterate returns items for {{#each}} blocks, or nil when not iterable.
func Iterate(value any) []IterItem {
	if value == nil {
		return nil
	}
	v := reflect.ValueOf(value)
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return nil
		}
		items := make([]IterItem, v.Len())
		for i := 0; i < v.Len(); i++ {
			items[i] = IterItem{Value: v.Index(i).Interface(), Index: i}
		}
		return items
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return nil
		}
		keys := v.MapKeys()
		if len(keys) == 0 {
			return nil
		}
		names := make([]string, len(keys))
		for i, key := range keys {
			names[i] = key.String()
		}
		sort.Strings(names)
		items := make([]IterItem, len(names))
		for i, key := range names {
			items[i] = IterItem{
				Value: v.MapIndex(reflect.ValueOf(key)).Interface(),
				Key:   key,
				Index: i,
			}
		}
		return items
	default:
		return nil
	}
}
