package dme

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// DNS names are case-insensitive per RFC 1035 §2.3.3. DME canonicalizes
// the host-shaped fields (record name, MX/CNAME/NS/ANAME target) to
// lower case on storage; users routinely author them in mixed case. We
// suppress case-only diffs so a refresh against lower-cased state does
// not produce a no-op plan against mixed-case config.
//
// Empirical basis: direct-REST PUT+GET probes against api.dnsmadeeasy.com
// confirming server-side lowercasing of name and MX/CNAME/NS/ANAME value
// fields. Evidence captured in an internal DNS IaC repository.

// dnsHostnameValueTypes is the set of record types whose `value`
// attribute holds a DNS hostname (case-insensitive per RFC 1035).
// Other record types keep byte-exact comparison on `value`.
var dnsHostnameValueTypes = map[string]bool{
	"MX":    true,
	"CNAME": true,
	"NS":    true,
	"ANAME": true,
}

// suppressCaseInsensitiveDNSValue suppresses semantically-equivalent
// diffs on the `value` attribute:
//
//   - DNS-hostname target types (MX, CNAME, NS, ANAME): case-only diff
//     suppressed (RFC 1035 §2.3.3; DME canonicalizes to lowercase
//     server-side). Covers DSO-3497 Cat 1.
//   - TXT-like types (TXT, SPF, CAA): the outer wrapping `"…"` that
//     DME's wire format adds may or may not be present in the user's
//     config. Suppress when the two values match after stripping any
//     single set of outer quotes. Covers DSO-3497 Cat 4.
func suppressCaseInsensitiveDNSValue(k, old, new string, d *schema.ResourceData) bool {
	rt := strings.ToUpper(toString(d.Get("type")))
	switch {
	case dnsHostnameValueTypes[rt]:
		return strings.EqualFold(old, new)
	case txtLikeTypes[rt]:
		return stripOuterDoubleQuotes(old) == stripOuterDoubleQuotes(new)
	}
	return false
}

func toString(v interface{}) string {
	s, _ := v.(string)
	return s
}

// stripOuterDoubleQuotes removes a single pair of outer `"…"` from s,
// or returns s unchanged if not so wrapped. Used by the cat 4
// DiffSuppressFunc to treat `X` and `"X"` as equivalent.
func stripOuterDoubleQuotes(s string) string {
	if len(s) >= 2 && strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		return s[1 : len(s)-1]
	}
	return s
}

// suppressCaseInsensitiveDNSName suppresses case-only diffs on the
// `name` attribute. DNS record names are always case-insensitive.
func suppressCaseInsensitiveDNSName(k, old, new string, d *schema.ResourceData) bool {
	return strings.EqualFold(old, new)
}
