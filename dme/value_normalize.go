package dme

import "strings"

// txtLikeTypes are the record types whose `value` is stored by DME
// wrapped in outer `"…"` and, for values > 255 bytes, split into RFC
// 1035 §3.3.14 multi-string form with internal `""` junctions.
var txtLikeTypes = map[string]bool{
	"TXT": true,
	"SPF": true,
	"CAA": true,
}

// normalizeValueOnRead converts DME's on-the-wire `value` representation
// for TXT-like record types into the user-facing logical string:
//
//   - Strip outer `"…"` if present (DSO-3497 Cat 4)
//   - Replace internal `""` junctions with empty string, concatenating
//     RFC 1035 multi-string chunks back into the single logical value
//     (DSO-3497 Cat 5b)
//
// For non-TXT-like record types the value is returned unchanged.
func normalizeValueOnRead(recordType, raw string) string {
	if !txtLikeTypes[strings.ToUpper(recordType)] {
		return raw
	}
	v := raw
	if len(v) >= 2 && strings.HasPrefix(v, `"`) && strings.HasSuffix(v, `"`) {
		v = v[1 : len(v)-1]
	}
	// Internal `""` junctions collapse the multi-string back into the
	// single logical value.
	if strings.Contains(v, `""`) {
		v = strings.ReplaceAll(v, `""`, "")
	}
	return v
}
