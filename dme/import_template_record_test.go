package dme

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Import format: "<template_id>:<record_id>"
// Example: tofu import dme_template_record.mx 12345:67890

func TestParseTemplateRecordImportID_Valid(t *testing.T) {
	templateID, recordID, err := parseTemplateRecordImportID("12345:67890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if templateID != "12345" {
		t.Errorf("templateID = %q, want %q", templateID, "12345")
	}
	if recordID != "67890" {
		t.Errorf("recordID = %q, want %q", recordID, "67890")
	}
}

func TestParseTemplateRecordImportID_Errors(t *testing.T) {
	cases := []struct {
		name string
		id   string
	}{
		{"no colon", "67890"},
		{"empty", ""},
		{"empty template", ":67890"},
		{"empty record", "12345:"},
		{"only colon", ":"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := parseTemplateRecordImportID(tc.id); err == nil {
				t.Errorf("expected error for input %q, got nil", tc.id)
			}
		})
	}
}

func TestTemplateRecordImporter_StateTransform(t *testing.T) {
	res := resourceDMETemplateRecord()
	if res.Importer == nil || res.Importer.State == nil {
		t.Fatal("expected Importer.State on dme_template_record")
	}

	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"template_id": "", "name": "", "type": "", "value": "", "ttl": "",
	})
	d.SetId("12345:67890")

	out, err := res.Importer.State(d, nil)
	if err != nil {
		t.Fatalf("Importer.State error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 ResourceData, got %d", len(out))
	}
	if got := out[0].Id(); got != "67890" {
		t.Errorf("Id() = %q, want %q", got, "67890")
	}
	if tid, _ := out[0].Get("template_id").(string); tid != "12345" {
		t.Errorf("template_id = %q, want %q", tid, "12345")
	}
}
