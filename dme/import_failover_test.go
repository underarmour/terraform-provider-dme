package dme

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// dme_failover's Read uses d.Get("record_id"), not d.Id(). The importer
// must populate record_id from the import ID so Read can issue its
// monitor lookup.
//
// Import format: "<record_id>"
// Example: tofu import dme_failover.www 227177880

func TestFailoverImporter_StateTransform(t *testing.T) {
	res := resourceDMEFailover()
	if res.Importer == nil || res.Importer.State == nil {
		t.Fatal("expected Importer.State on dme_failover")
	}

	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"record_id": "",
	})
	d.SetId("227177880")

	out, err := res.Importer.State(d, nil)
	if err != nil {
		t.Fatalf("Importer.State error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 ResourceData, got %d", len(out))
	}
	if rid, _ := out[0].Get("record_id").(string); rid != "227177880" {
		t.Errorf("record_id = %q, want %q", rid, "227177880")
	}
}

func TestFailoverImporter_EmptyID(t *testing.T) {
	res := resourceDMEFailover()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("")
	if _, err := res.Importer.State(d, nil); err == nil {
		t.Error("expected error for empty import ID, got nil")
	}
}
