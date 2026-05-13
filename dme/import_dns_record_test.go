package dme

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Import format: "<domain_id>:<record_id>"
// Example: tofu import dme_dns_record.www 8249868:227177880

func TestParseDNSRecordImportID_Valid(t *testing.T) {
	domainID, recordID, err := parseDNSRecordImportID("8249868:227177880")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if domainID != "8249868" {
		t.Errorf("domainID = %q, want %q", domainID, "8249868")
	}
	if recordID != "227177880" {
		t.Errorf("recordID = %q, want %q", recordID, "227177880")
	}
}

func TestParseDNSRecordImportID_Errors(t *testing.T) {
	cases := []struct {
		name string
		id   string
	}{
		{"no colon", "227177880"},
		{"empty", ""},
		{"empty domain", ":227177880"},
		{"empty record", "8249868:"},
		{"only colon", ":"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := parseDNSRecordImportID(tc.id); err == nil {
				t.Errorf("expected error for input %q, got nil", tc.id)
			}
		})
	}
}

func TestDNSRecordImporter_StateTransform(t *testing.T) {
	res := resourceManagedDNSRecordActions()
	if res.Importer == nil || res.Importer.State == nil {
		t.Fatal("expected Importer.State on dme_dns_record")
	}

	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"domain_id": "", "name": "", "type": "", "value": "", "ttl": "",
	})
	d.SetId("8249868:227177880")

	out, err := res.Importer.State(d, nil)
	if err != nil {
		t.Fatalf("Importer.State error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 ResourceData, got %d", len(out))
	}
	if got := out[0].Id(); got != "227177880" {
		t.Errorf("Id() = %q, want %q", got, "227177880")
	}
	if dom, _ := out[0].Get("domain_id").(string); dom != "8249868" {
		t.Errorf("domain_id = %q, want %q", dom, "8249868")
	}
}

func TestDNSRecordImporter_InvalidID(t *testing.T) {
	res := resourceManagedDNSRecordActions()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("not-composite")
	if _, err := res.Importer.State(d, nil); err == nil {
		t.Error("expected error for invalid import ID, got nil")
	}
}
