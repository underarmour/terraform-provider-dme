package dme

import (
	"testing"

	"github.com/DNSMadeEasy/dme-go-client/container"
)

const sampleRecordsListJSON = `{
  "data": [
    {"id": 100, "name": "alpha", "type": "A",   "value": "1.1.1.1"},
    {"id": 200, "name": "beta",  "type": "TXT", "value": "\"v=spf1 -all\""},
    {"id": 300, "name": "mail",  "type": "MX",  "value": "mx.example.com"}
  ],
  "totalRecords": 3,
  "totalPages": 1
}`

func mustParseJSON(t *testing.T, body string) *container.Container {
	t.Helper()
	c, err := container.ParseJSON([]byte(body))
	if err != nil {
		t.Fatalf("ParseJSON: %v", err)
	}
	return c
}

func TestFindRecordByID_Found(t *testing.T) {
	con := mustParseJSON(t, sampleRecordsListJSON)
	got := findRecordByID(con, "200")
	if got == nil {
		t.Fatal("expected record id=200, got nil")
	}
	if v := extractField(got.S("name")); v != "beta" {
		t.Errorf("name = %q, want %q", v, "beta")
	}
}

func TestFindRecordByID_NotFound(t *testing.T) {
	con := mustParseJSON(t, sampleRecordsListJSON)
	if got := findRecordByID(con, "999"); got != nil {
		t.Error("expected nil for missing id")
	}
}

func TestFindRecordByID_EmptyData(t *testing.T) {
	con := mustParseJSON(t, `{"data":[],"totalRecords":0}`)
	if got := findRecordByID(con, "100"); got != nil {
		t.Error("expected nil on empty data")
	}
}

func TestFindRecordByID_NilContainer(t *testing.T) {
	if got := findRecordByID(nil, "100"); got != nil {
		t.Error("expected nil on nil container")
	}
}

func TestFindRecordByID_MissingDataKey(t *testing.T) {
	con := mustParseJSON(t, `{"totalRecords":0}`)
	if got := findRecordByID(con, "100"); got != nil {
		t.Error("expected nil when data key absent")
	}
}

func TestRecordIDMatches(t *testing.T) {
	cases := []struct {
		name   string
		val    interface{}
		target string
		want   bool
	}{
		{"float64 match", float64(227177880), "227177880", true},
		{"float64 miss", float64(100), "200", false},
		{"int match", 12345, "12345", true},
		{"int64 match", int64(98765), "98765", true},
		{"string match", "12345", "12345", true},
		{"string miss", "99999", "12345", false},
		{"unsupported type", true, "12345", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := recordIDMatches(tc.val, tc.target); got != tc.want {
				t.Errorf("recordIDMatches(%v, %q) = %v, want %v", tc.val, tc.target, got, tc.want)
			}
		})
	}
}
