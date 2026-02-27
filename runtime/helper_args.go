package runtime

import (
	"strconv"

	"github.com/andriyg76/hexerr"
)

// GetArg returns the positional argument at index, or nil if out of bounds.
func (a HelperArgs) GetArg(idx int) any {
	if a.Args == nil || idx < 0 || idx >= len(a.Args) {
		return nil
	}
	return a.Args[idx]
}

// GetString returns the stringified argument at index, or empty string.
func (a HelperArgs) GetString(idx int) string {
	arg := a.GetArg(idx)
	if arg == nil {
		return ""
	}
	return Stringify(arg)
}

// GetHash returns the hash argument for key, or nil if missing or HashArgs is nil.
func (a HelperArgs) GetHash(key string) any {
	if a.HashArgs == nil {
		return nil
	}
	return a.HashArgs[key]
}

// GetHashString returns the stringified hash value for key, or empty string.
func (a HelperArgs) GetHashString(key string) string {
	v := a.GetHash(key)
	if v == nil {
		return ""
	}
	return Stringify(v)
}

// GetHashNumber converts the hash value for key to float64. Returns (0, nil) if missing or nil.
func (a HelperArgs) GetHashNumber(key string) (float64, error) {
	v := a.GetHash(key)
	if v == nil {
		return 0, nil
	}
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, hexerr.Newf("cannot convert hash %q value %T to number", key, v)
	}
}

// GetNumber converts the argument at index to float64. Returns (0, nil) for nil or out of bounds.
func (a HelperArgs) GetNumber(idx int) (float64, error) {
	arg := a.GetArg(idx)
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
