package dme

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// parseDNSRecordImportID splits a composite import ID of the form
// "<domain_id>:<record_id>" into its parts. Both parts must be non-empty.
// DME does not expose a record-by-id endpoint independent of domain
// context, so the domain ID is required to drive Read.
func parseDNSRecordImportID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf(
			"invalid import ID %q: expected format <domain_id>:<record_id>", id)
	}
	return parts[0], parts[1], nil
}

func importDNSRecordState(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	domainID, recordID, err := parseDNSRecordImportID(d.Id())
	if err != nil {
		return nil, err
	}
	if err := d.Set("domain_id", domainID); err != nil {
		return nil, err
	}
	d.SetId(recordID)
	return []*schema.ResourceData{d}, nil
}
