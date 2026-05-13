package dme

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func parseTemplateRecordImportID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf(
			"invalid import ID %q: expected format <template_id>:<record_id>", id)
	}
	return parts[0], parts[1], nil
}

func importTemplateRecordState(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	templateID, recordID, err := parseTemplateRecordImportID(d.Id())
	if err != nil {
		return nil, err
	}
	if err := d.Set("template_id", templateID); err != nil {
		return nil, err
	}
	d.SetId(recordID)
	return []*schema.ResourceData{d}, nil
}
