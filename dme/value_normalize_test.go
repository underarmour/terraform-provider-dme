package dme

import (
	"strings"
	"testing"
)

// TestNormalizeValueOnRead_Cat4_TXTQuoteStrip covers DSO-3497 Cat 4.
//
// DME wraps TXT/SPF/CAA `value` with outer `"…"` on storage (verified
// empirically: PUT `v=spf1 -all`, GET returns `"v=spf1 -all"`). The
// provider must strip the wire-format outer quotes on read so state
// contains the user-facing value.
func TestNormalizeValueOnRead_Cat4_TXTQuoteStrip(t *testing.T) {
	cases := []struct {
		name    string
		recType string
		raw     string
		want    string
	}{
		{name: "short TXT outer quotes stripped", recType: "TXT", raw: `"v=spf1 -all"`, want: "v=spf1 -all"},
		{name: "SPF outer quotes stripped", recType: "SPF", raw: `"v=spf1 a mx"`, want: "v=spf1 a mx"},
		{name: "CAA outer quotes stripped", recType: "CAA", raw: `"letsencrypt.org"`, want: "letsencrypt.org"},
		{name: "TXT with embedded equals preserved", recType: "TXT", raw: `"MS=ms51334084"`, want: "MS=ms51334084"},
		{name: "TXT unwrapped value is left alone", recType: "TXT", raw: `bare`, want: "bare"},
		{name: "empty TXT value", recType: "TXT", raw: ``, want: ""},
		{name: "two-char TXT '\"\"' is the empty string", recType: "TXT", raw: `""`, want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeValueOnRead(tc.recType, tc.raw)
			if got != tc.want {
				t.Errorf("normalizeValueOnRead(%q, %q) = %q; want %q", tc.recType, tc.raw, got, tc.want)
			}
		})
	}
}

// TestNormalizeValueOnRead_Cat5b_LongTXTConcatenation covers DSO-3497
// Cat 5b. DME stores TXT values > 255 bytes in RFC 1035 §3.3.14
// multi-string wire format: each ≤255-byte chunk wrapped in `"…"`,
// concatenated. Example: `"<255 chars>""<remaining chars>"`. The user-
// facing value is the logical concatenation; the internal `""`
// junctions are a wire-encoding detail.
func TestNormalizeValueOnRead_Cat5b_LongTXTConcatenation(t *testing.T) {
	// Mirror the actual probe value: 150 'A's + 150 'b's = 300 chars.
	// DME splits at 255, producing 150 'A' + 105 'b' in chunk 1 and 45 'b' in chunk 2.
	part1 := strings.Repeat("A", 150) + strings.Repeat("b", 105)
	part2 := strings.Repeat("b", 45)
	wireForm := `"` + part1 + `""` + part2 + `"`
	logicalForm := strings.Repeat("A", 150) + strings.Repeat("b", 150)

	got := normalizeValueOnRead("TXT", wireForm)
	if got != logicalForm {
		t.Errorf("normalizeValueOnRead long-TXT did not concatenate multi-string chunks\n got length=%d, want length=%d\n got first 40=%q\n got around split (245-265)=%q",
			len(got), len(logicalForm), got[:40], got[245:265])
	}
}

// Real-world DKIM record shape: multi-string wire form with an internal
// `""` junction at the 255-byte split, confirmed via direct-REST probe
// against api.dnsmadeeasy.com. After normalize, no internal `""` should
// remain and no outer `"` should bookend the value.
func TestNormalizeValueOnRead_Cat5b_DKIMProductionShape(t *testing.T) {
	// Truncated for test compactness; structure preserved.
	wire := `"v=DKIM1; k=rsa; p=AAAA""BBBB"`
	got := normalizeValueOnRead("TXT", wire)
	want := "v=DKIM1; k=rsa; p=AAAABBBB"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	if strings.Contains(got, `""`) {
		t.Errorf("normalized value still contains internal '\"\"' junction: %q", got)
	}
}

// Non-string record types (MX, A, CNAME, NS, HTTPRED, ANAME) must
// pass through unchanged.
func TestNormalizeValueOnRead_NonTXTPassthrough(t *testing.T) {
	cases := []struct {
		recType string
		raw     string
	}{
		{"MX", "alt1.aspmx.l.google.com."},
		{"A", "1.2.3.4"},
		{"AAAA", "2001:db8::1"},
		{"CNAME", "target.example.com."},
		{"NS", "ns1.example.com."},
		{"ANAME", "n.sni.global.fastly.net."},
		{"HTTPRED", "https://example.com/?a=1&b=2"},
	}
	for _, tc := range cases {
		t.Run(tc.recType, func(t *testing.T) {
			got := normalizeValueOnRead(tc.recType, tc.raw)
			if got != tc.raw {
				t.Errorf("normalizeValueOnRead(%q, %q) modified value to %q; expected passthrough",
					tc.recType, tc.raw, got)
			}
		})
	}
}
