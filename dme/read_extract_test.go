package dme

import (
	"testing"

	"github.com/DNSMadeEasy/dme-go-client/container"
)

// TestExtractField_AmpersandPreserved covers DSO-3497 Cat 2.
//
// gabs Container.String() → json.Marshal which HTML-escapes `&` to the
// literal 6-byte sequence `\u0026` (also `<` → `\u003c`, `>` → `\u003e`).
// HTTPRED `value` strings legitimately contain `&` in URL query strings;
// the provider read path must not mangle them.
//
// Empirical basis: cat2-roundtrip-2026-05-07T13-40-09.json — direct-REST
// roundtrip confirmed DME stores `&` verbatim. The mangling is in the
// provider's interpretation of the response, not on the wire.
func TestExtractField_AmpersandPreserved(t *testing.T) {
	raw := []byte(`{"value":"https://example.com/sso?fixture=dnsiac&category=cat2&binding=POST"}`)
	c, err := container.ParseJSON(raw)
	if err != nil {
		t.Fatalf("ParseJSON: %v", err)
	}

	got := extractField(c.S("value"))
	want := "https://example.com/sso?fixture=dnsiac&category=cat2&binding=POST"

	if got != want {
		t.Errorf("extractField returned %q\nwant %q", got, want)
	}
}

func TestExtractField_AngleBracketsPreserved(t *testing.T) {
	// Go's json.Marshal HTML-escapes `<` and `>` too. Same root cause as
	// the `&` escaping; same fix applies.
	raw := []byte(`{"value":"a<b>c"}`)
	c, err := container.ParseJSON(raw)
	if err != nil {
		t.Fatalf("ParseJSON: %v", err)
	}

	got := extractField(c.S("value"))
	want := "a<b>c"

	if got != want {
		t.Errorf("extractField returned %q\nwant %q", got, want)
	}
}

func TestExtractField_PlainStringRoundtrip(t *testing.T) {
	raw := []byte(`{"value":"plain string"}`)
	c, _ := container.ParseJSON(raw)
	if got, want := extractField(c.S("value")), "plain string"; got != want {
		t.Errorf("extractField returned %q, want %q", got, want)
	}
}

func TestExtractField_NumberAsString(t *testing.T) {
	// DME returns numeric fields (ttl, mxLevel, priority, etc.) as JSON
	// numbers; provider schema represents them as TypeString. The
	// extractor must coerce.
	raw := []byte(`{"ttl":3600,"mxLevel":10}`)
	c, _ := container.ParseJSON(raw)
	if got, want := extractField(c.S("ttl")), "3600"; got != want {
		t.Errorf("extractField(ttl) = %q, want %q", got, want)
	}
	if got, want := extractField(c.S("mxLevel")), "10"; got != want {
		t.Errorf("extractField(mxLevel) = %q, want %q", got, want)
	}
}

func TestExtractField_BoolAsString(t *testing.T) {
	raw := []byte(`{"issuerCritical":true,"hardLink":false}`)
	c, _ := container.ParseJSON(raw)
	if got, want := extractField(c.S("issuerCritical")), "true"; got != want {
		t.Errorf("extractField(issuerCritical) = %q, want %q", got, want)
	}
	if got, want := extractField(c.S("hardLink")), "false"; got != want {
		t.Errorf("extractField(hardLink) = %q, want %q", got, want)
	}
}

func TestExtractField_NilContainer(t *testing.T) {
	if got, want := extractField(nil), ""; got != want {
		t.Errorf("extractField(nil) = %q, want %q", got, want)
	}
}

func TestExtractField_NullValue(t *testing.T) {
	raw := []byte(`{"value":null}`)
	c, _ := container.ParseJSON(raw)
	if got, want := extractField(c.S("value")), ""; got != want {
		t.Errorf("extractField(null) = %q, want %q", got, want)
	}
}
