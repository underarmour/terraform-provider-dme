package dme

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// importFailoverState handles dme_failover import. The resource's Read
// function uses d.Get("record_id") rather than d.Id() to build its API
// path, so the importer must populate that field from the import ID.
// The import ID is the DNS record ID that has a failover monitor attached.
func importFailoverState(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	if d.Id() == "" {
		return nil, fmt.Errorf("import ID must be the record_id of the failover monitor")
	}
	if err := d.Set("record_id", d.Id()); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
