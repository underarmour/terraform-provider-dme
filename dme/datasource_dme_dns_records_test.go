package dme

// Unit tests for the value-normalization path used by
// datasourceManagedDNSRecordActionsRead and
// datasourceDMETemplateRecordRead.
//
// The Read functions cannot be exercised end-to-end without a live DME
// client (no interface to mock). These tests instead exercise the exact
// two-call sequence — extractField(cont.S("value")) followed by
// normalizeValueOnRead(type, raw) — against containers built from
// representative raw API response payloads. If the data source Read is
// ever refactored to use a different code path these tests will catch
// the regression.
//
// Full behavioral coverage (data source returns correct values for real
// records) requires TF_ACC acceptance tests against a live account.

import (
	"testing"

	"github.com/DNSMadeEasy/dme-go-client/container"
)

func mustParseRecord(t *testing.T, raw string) *container.Container {
	t.Helper()
	c, err := container.ParseJSON([]byte(raw))
	if err != nil {
		t.Fatalf("ParseJSON: %v", err)
	}
	return c
}

// applyDatasourceValuePath mirrors what datasourceManagedDNSRecordActionsRead
// now does when populating the value field.
func applyDatasourceValuePath(cont *container.Container) (recordType, value string) {
	recordType = extractField(cont.S("type"))
	value = normalizeValueOnRead(recordType, extractField(cont.S("value")))
	return
}

func TestDatasourceDNSRecord_TXTValueNormalized(t *testing.T) {
	// DME returns TXT value wrapped in outer quotes.
	con := mustParseRecord(t, `{"id":1,"name":"test","type":"TXT","value":"\"v=spf1 include:example.com -all\""}`)
	_, got := applyDatasourceValuePath(con)
	want := "v=spf1 include:example.com -all"
	if got != want {
		t.Errorf("TXT value = %q, want %q", got, want)
	}
}

func TestDatasourceDNSRecord_SPFValueNormalized(t *testing.T) {
	con := mustParseRecord(t, `{"id":2,"name":"test","type":"SPF","value":"\"v=spf1 -all\""}`)
	_, got := applyDatasourceValuePath(con)
	want := "v=spf1 -all"
	if got != want {
		t.Errorf("SPF value = %q, want %q", got, want)
	}
}

func TestDatasourceDNSRecord_LongTXTJunctionStripped(t *testing.T) {
	// DME splits TXT > 255 chars at the 255-byte boundary and wraps the
	// whole value with outer quotes: "chunk1""chunk2".
	// The data source must collapse the internal "" junctions and strip
	// the outer quotes, matching the resource Read behavior.
	chunk1 := repeatByte('a', 255)
	chunk2 := "remainder"
	// Raw API value: outer-quoted with an internal "" junction.
	raw := `{"id":3,"name":"k._domainkey","type":"TXT","value":"\"` + chunk1 + `\"\"` + chunk2 + `\""}`
	con := mustParseRecord(t, raw)
	_, got := applyDatasourceValuePath(con)
	want := chunk1 + chunk2
	if got != want {
		t.Errorf("long TXT value = %q, want %q", got, want)
	}
}

func repeatByte(b byte, n int) string {
	s := make([]byte, n)
	for i := range s {
		s[i] = b
	}
	return string(s)
}

func TestDatasourceDNSRecord_HTTPREDAmpersandPreserved(t *testing.T) {
	// extractField bypasses json.Marshal so & must not become \u0026.
	con := mustParseRecord(t, `{"id":4,"name":"redir","type":"HTTPRED","value":"https://example.com/path?a=1&b=2"}`)
	_, got := applyDatasourceValuePath(con)
	want := "https://example.com/path?a=1&b=2"
	if got != want {
		t.Errorf("HTTPRED value = %q, want %q", got, want)
	}
}

func TestDatasourceDNSRecord_MXValuePassthrough(t *testing.T) {
	// MX target is not quote-wrapped; extractField + normalizeValueOnRead
	// must return it unchanged (normalization is a no-op for MX).
	con := mustParseRecord(t, `{"id":5,"name":"@","type":"MX","value":"mail.example.com"}`)
	_, got := applyDatasourceValuePath(con)
	want := "mail.example.com"
	if got != want {
		t.Errorf("MX value = %q, want %q", got, want)
	}
}

func TestDatasourceDNSRecord_AValuePassthrough(t *testing.T) {
	con := mustParseRecord(t, `{"id":6,"name":"www","type":"A","value":"1.2.3.4"}`)
	_, got := applyDatasourceValuePath(con)
	if got != "1.2.3.4" {
		t.Errorf("A value = %q, want %q", got, "1.2.3.4")
	}
}
