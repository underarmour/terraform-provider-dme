package dme

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// TestSuppressCaseInsensitiveDNSValue_MX covers DSO-3497 Cat 1.
//
// Scenario: user authors an MX target hostname in mixed/upper case
// (e.g. "ALT1.ASPMX.L.GOOGLE.COM."), DME returns it canonicalized to
// lower case ("alt1.aspmx.l.google.com.") on subsequent refresh, and
// the plan must show 0 changes — because DNS names are case-
// insensitive per RFC 1035 §2.3.3.
//
// The fix is a schema-level DiffSuppressFunc on the `value` attribute
// that case-folds when the record type is one whose value is a DNS
// hostname (MX, CNAME, NS, ANAME). Values for case-sensitive types
// (A, AAAA, TXT, SPF, CAA, HTTPRED) must not be folded.
func TestSuppressCaseInsensitiveDNSValue_MX(t *testing.T) {
	cases := []struct {
		name    string
		recType string
		old     string // value as read back from DME (state)
		new     string // value as authored by the user (config)
		want    bool   // true = suppress diff
	}{
		{
			name:    "MX target case-only diff is suppressed",
			recType: "MX",
			old:     "alt1.aspmx.l.google.com.",
			new:     "ALT1.ASPMX.L.GOOGLE.COM.",
			want:    true,
		},
		{
			name:    "MX target real change is not suppressed",
			recType: "MX",
			old:     "alt1.aspmx.l.google.com.",
			new:     "alt2.aspmx.l.google.com.",
			want:    false,
		},
		{
			name:    "CNAME target case-only diff is suppressed",
			recType: "CNAME",
			old:     "target.example.com.",
			new:     "Target.Example.Com.",
			want:    true,
		},
		{
			name:    "NS target case-only diff is suppressed",
			recType: "NS",
			old:     "ns1.example.com.",
			new:     "NS1.example.com.",
			want:    true,
		},
		{
			name:    "ANAME target case-only diff is suppressed",
			recType: "ANAME",
			old:     "n.sni.global.fastly.net.",
			new:     "N.SNI.GLOBAL.FASTLY.NET.",
			want:    true,
		},
		{
			name:    "TXT value case difference is NOT suppressed (TXT is byte-exact)",
			recType: "TXT",
			old:     "v=spf1 a mx",
			new:     "V=SPF1 A MX",
			want:    false,
		},
		{
			name:    "A value case difference is NOT suppressed (IPv4 has no case)",
			recType: "A",
			old:     "1.2.3.4",
			new:     "1.2.3.5",
			want:    false,
		},
		{
			name:    "HTTPRED value case-only diff is NOT suppressed (URLs are case-sensitive)",
			recType: "HTTPRED",
			old:     "https://example.com/path",
			new:     "https://example.com/PATH",
			want:    false,
		},
	}

	res := resourceManagedDNSRecordActions()
	valueSchema := res.Schema["value"]
	if valueSchema.DiffSuppressFunc == nil {
		t.Fatal("expected DiffSuppressFunc on resourceManagedDNSRecordActions().Schema[\"value\"]; got nil")
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
				"domain_id": "12345",
				"name":      "test",
				"type":      tc.recType,
				"value":     tc.new,
				"ttl":       "300",
			})
			got := valueSchema.DiffSuppressFunc("value", tc.old, tc.new, d)
			if got != tc.want {
				t.Errorf("DiffSuppressFunc(type=%s, old=%q, new=%q) = %v; want %v",
					tc.recType, tc.old, tc.new, got, tc.want)
			}
		})
	}
}

// TestSuppressTXTValueOuterQuotes covers the remaining edge of
// DSO-3497 Cat 4 that the read-side strip alone cannot solve.
//
// After normalizeValueOnRead strips DME's wire-format outer `"…"`,
// state contains the unwrapped value. But some consumers author TXT
// values WITH literal outer quotes in their config (e.g. YAML
// "\"MS=ms51334084\"" yielding the Go string `"MS=ms51334084"`).
// State=`MS=ms51334084` ≠ config=`"MS=ms51334084"` — still drift.
//
// The fix is a DiffSuppressFunc that treats `X` and `"X"` as
// equivalent for TXT/SPF/CAA values. Either authoring form is then
// accepted as a no-op against the same DME-stored record.
func TestSuppressTXTValueOuterQuotes(t *testing.T) {
	cases := []struct {
		name    string
		recType string
		old     string
		new     string
		want    bool
	}{
		{name: "TXT bare state vs quoted config is suppressed", recType: "TXT", old: "MS=ms51334084", new: `"MS=ms51334084"`, want: true},
		{name: "TXT quoted state vs bare config is suppressed (symmetric)", recType: "TXT", old: `"MS=ms51334084"`, new: "MS=ms51334084", want: true},
		{name: "TXT identical bare values suppress (Terraform won't even call this case, but the result is benign)", recType: "TXT", old: "bar", new: "bar", want: true},
		{name: "TXT real value change is not suppressed", recType: "TXT", old: "foo", new: "bar", want: false},
		{name: "TXT bare-vs-quoted-with-DIFFERENT-content is not suppressed", recType: "TXT", old: "foo", new: `"bar"`, want: false},
		{name: "SPF bare-vs-quoted is suppressed", recType: "SPF", old: "v=spf1 a mx", new: `"v=spf1 a mx"`, want: true},
		{name: "CAA bare-vs-quoted is suppressed", recType: "CAA", old: "letsencrypt.org", new: `"letsencrypt.org"`, want: true},
		{name: "HTTPRED is not TXT-like; outer quotes are not equivalent", recType: "HTTPRED", old: "https://example.com/", new: `"https://example.com/"`, want: false},
	}
	res := resourceManagedDNSRecordActions()
	valueSchema := res.Schema["value"]
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
				"domain_id": "12345", "name": "test", "type": tc.recType, "value": tc.new, "ttl": "300",
			})
			got := valueSchema.DiffSuppressFunc("value", tc.old, tc.new, d)
			if got != tc.want {
				t.Errorf("DiffSuppressFunc(type=%s, old=%q, new=%q) = %v; want %v", tc.recType, tc.old, tc.new, got, tc.want)
			}
		})
	}
}

// TestSuppressCaseInsensitiveDNSName_TXT covers DSO-3497 Cat 3 and Cat 5a.
//
// Scenario: user authors a TXT record `name` in mixed/upper case (e.g.
// "_DMARC.gr" or "TXTName.DNSIAC-FIXTURE"), DME canonicalizes to lower
// case on storage, refresh pulls the lower-case form into state, and
// a plan run shows a no-op `name` diff that — because `name` is
// ForceNew — manifests as `forces replacement`. The fix is a
// DiffSuppressFunc on the `name` attribute that case-folds.
//
// Cat 5a (long-TXT name) shares the same root cause and is covered by
// the same suppress hook; long-TXT differs from short-TXT only in
// value length, not in name handling.
func TestSuppressCaseInsensitiveDNSName_TXT(t *testing.T) {
	cases := []struct {
		name    string
		recType string
		old     string // state value
		new     string // config value
		want    bool
	}{
		{
			name:    "TXT name DMARC-style case-only diff is suppressed",
			recType: "TXT",
			old:     "_dmarc.gr",
			new:     "_DMARC.gr",
			want:    true,
		},
		{
			name:    "TXT name DKIM-style case-only diff is suppressed",
			recType: "TXT",
			old:     "6104caa4-e585-11eb-a222-fea9c8e4c65d._domainkey.info",
			new:     "6104CAA4-E585-11EB-A222-FEA9C8E4C65D._domainkey.info",
			want:    true,
		},
		{
			name:    "TXT name real change is not suppressed",
			recType: "TXT",
			old:     "_dmarc.gr",
			new:     "_dmarc.uk",
			want:    false,
		},
		{
			name:    "Long-TXT name case-only diff is suppressed (cat 5a)",
			recType: "TXT",
			old:     "whiteimage._domainkey.eshop",
			new:     "Whiteimage._domainkey.eshop",
			want:    true,
		},
		{
			name:    "MX name case-only diff is also suppressed",
			recType: "MX",
			old:     "mail.example.com",
			new:     "Mail.Example.Com",
			want:    true,
		},
	}

	res := resourceManagedDNSRecordActions()
	nameSchema := res.Schema["name"]
	if nameSchema.DiffSuppressFunc == nil {
		t.Fatal("expected DiffSuppressFunc on resourceManagedDNSRecordActions().Schema[\"name\"]; got nil")
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
				"domain_id": "12345",
				"name":      tc.new,
				"type":      tc.recType,
				"value":     "irrelevant",
				"ttl":       "300",
			})
			got := nameSchema.DiffSuppressFunc("name", tc.old, tc.new, d)
			if got != tc.want {
				t.Errorf("DiffSuppressFunc(type=%s, old=%q, new=%q) = %v; want %v",
					tc.recType, tc.old, tc.new, got, tc.want)
			}
		})
	}
}

// TestTXTNameCaseOnlyDiff_DoesNotForceReplacement is the integration
// test that proves DiffSuppressFunc short-circuits ForceNew evaluation
// for the cat 3 scenario: state has lowercase, config has uppercase,
// plan should show neither an in-place update nor a replacement.
func TestTXTNameCaseOnlyDiff_DoesNotForceReplacement(t *testing.T) {
	res := resourceManagedDNSRecordActions()

	// Construct prior state (what DME's lower-cased read returns) and
	// new config (mixed case as authored) for a TXT record.
	priorAttrs := map[string]string{
		"id":        "1",
		"domain_id": "12345",
		"name":      "6104caa4-e585-11eb-a222-fea9c8e4c65d._domainkey.info",
		"type":      "TXT",
		"value":     "v=DKIM1; k=rsa; p=AAAA",
		"ttl":       "3600",
	}
	priorState := &terraform.InstanceState{ID: "1", Attributes: priorAttrs}

	newConfig := terraform.NewResourceConfigRaw(map[string]interface{}{
		"domain_id": "12345",
		"name":      "6104CAA4-E585-11EB-A222-FEA9C8E4C65D._domainkey.info",
		"type":      "TXT",
		"value":     "v=DKIM1; k=rsa; p=AAAA",
		"ttl":       "3600",
	})

	diff, err := res.Diff(priorState, newConfig, nil)
	if err != nil {
		t.Fatalf("res.Diff returned error: %v", err)
	}
	if diff != nil && !diff.Empty() {
		t.Errorf("expected empty diff for case-only name change; got: %#v", diff.Attributes)
	}
	if diff != nil && diff.RequiresNew() {
		t.Errorf("case-only TXT name diff must not RequiresNew (would destroy/create DKIM record); diff: %#v", diff.Attributes)
	}
}
