package dme

import (
	"strconv"

	"github.com/DNSMadeEasy/dme-go-client/container"
)

// extractField returns the string representation of a gabs container's
// underlying value WITHOUT routing through json.Marshal.
//
// The upstream provider used `StripQuotes(c.S("field").String())`,
// which calls gabs Container.String → Container.Bytes → json.Marshal.
// Go's encoding/json defaults to HTML-safe escaping, converting `&` to
// the literal 6-byte sequence `\u0026` (and `<` to `\u003c`, `>` to
// `\u003c`). For fields like HTTPRED `value` that legitimately contain
// `&` in URL query strings, that mangles state. (DSO-3497 Cat 2.)
//
// This helper extracts the underlying interface{} directly:
//   - string: returned as-is
//   - float64: formatted as the equivalent integer or float string
//     (DME returns ttl, mxLevel, priority, weight, port as JSON numbers
//     but the provider schema represents them as TypeString)
//   - bool: "true"/"false" (issuerCritical, hardLink, etc.)
//   - nil: empty string
//
// Falls back to the StripQuotes(String()) path only for shapes we
// don't expect, so behavior degrades gracefully rather than panicking.
func extractField(c *container.Container) string {
	if c == nil {
		return ""
	}
	switch v := c.Data().(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		// Render integers without a decimal point so ttl=3600 doesn't
		// become "3600.000000". Strconv handles either case.
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return StripQuotes(c.String())
	}
}
