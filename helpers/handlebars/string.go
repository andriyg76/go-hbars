package handlebars

import (
	"strings"

	"github.com/andriyg76/go-hbars/runtime"
)

// Upper converts a string to uppercase.
func Upper(args runtime.HelperArgs) (any, error) {
	s := args.GetString(0)
	return strings.ToUpper(s), nil
}

// Lower converts a string to lowercase.
func Lower(args runtime.HelperArgs) (any, error) {
	s := args.GetString(0)
	return strings.ToLower(s), nil
}

// Capitalize capitalizes the first letter of a string.
func Capitalize(args runtime.HelperArgs) (any, error) {
	s := args.GetString(0)
	if s == "" {
		return "", nil
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:]), nil
}

// CapitalizeAll capitalizes all words in a string.
func CapitalizeAll(args runtime.HelperArgs) (any, error) {
	s := args.GetString(0)
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " "), nil
}

// Truncate truncates a string to the specified length.
func Truncate(args runtime.HelperArgs) (any, error) {
	s := args.GetString(0)
	length := 30
	if len(args.Args) > 1 {
		if n, err := args.GetNumber(1); err == nil {
			length = int(n)
		}
	}
	if n, err := args.GetHashNumber("length"); err == nil && n != 0 {
		length = int(n)
	}
	if suffix := args.GetHashString("suffix"); suffix != "" {
		if len(s) > length {
			return s[:length] + suffix, nil
		}
		return s, nil
	}
	if len(s) > length {
		return s[:length] + "...", nil
	}
	return s, nil
}

// Reverse reverses a string.
func Reverse(args runtime.HelperArgs) (any, error) {
	s := args.GetString(0)
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes), nil
}

// Replace replaces occurrences of a substring.
func Replace(args runtime.HelperArgs) (any, error) {
	s := args.GetString(0)
	old := args.GetString(1)
	new := args.GetString(2)
	return strings.ReplaceAll(s, old, new), nil
}

// StripTags removes HTML tags from a string.
func StripTags(args runtime.HelperArgs) (any, error) {
	s := args.GetString(0)
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(r)
		}
	}
	return result.String(), nil
}

// StripQuotes removes quotes from a string.
func StripQuotes(args runtime.HelperArgs) (any, error) {
	s := args.GetString(0)
	s = strings.TrimPrefix(s, `"`)
	s = strings.TrimSuffix(s, `"`)
	s = strings.TrimPrefix(s, `'`)
	s = strings.TrimSuffix(s, `'`)
	return s, nil
}

// Join joins array elements with a separator.
func Join(args runtime.HelperArgs) (any, error) {
	arr := args.GetArg(0)
	sep := ", "
	if len(args.Args) > 1 {
		sep = args.GetString(1)
	}
	if s := args.GetHashString("separator"); s != "" {
		sep = s
	}
	var parts []string
	switch v := arr.(type) {
	case []any:
		for _, item := range v {
			parts = append(parts, runtime.Stringify(item))
		}
	case []string:
		parts = v
	default:
		return "", nil
	}
	return strings.Join(parts, sep), nil
}

// Split splits a string by a separator.
func Split(args runtime.HelperArgs) (any, error) {
	s := args.GetString(0)
	sep := ","
	if len(args.Args) > 1 {
		sep = args.GetString(1)
	}
	if separator := args.GetHashString("separator"); separator != "" {
		sep = separator
	}
	return strings.Split(s, sep), nil
}

