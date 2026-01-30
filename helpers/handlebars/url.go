package handlebars

import (
	"net/url"

	"github.com/andriyg76/go-hbars/helpers"
)

// EncodeURI encodes a URI component.
func EncodeURI(args []any) (any, error) {
	s := helpers.GetStringArg(args, 0)
	return url.QueryEscape(s), nil
}

// DecodeURI decodes a URI component.
func DecodeURI(args []any) (any, error) {
	s := helpers.GetStringArg(args, 0)
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return s, nil
	}
	return decoded, nil
}

// StripProtocol strips the protocol from a URL.
func StripProtocol(args []any) (any, error) {
	s := helpers.GetStringArg(args, 0)
	u, err := url.Parse(s)
	if err != nil {
		return s, nil
	}
	u.Scheme = ""
	if u.Opaque != "" {
		return u.Opaque + u.Path + u.RawQuery + u.Fragment, nil
	}
	return u.Host + u.Path + u.RawQuery + u.Fragment, nil
}

// StripQuerystring strips the query string from a URL.
func StripQuerystring(args []any) (any, error) {
	s := helpers.GetStringArg(args, 0)
	u, err := url.Parse(s)
	if err != nil {
		return s, nil
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

