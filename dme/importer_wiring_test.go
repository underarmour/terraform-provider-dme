package dme

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// TestAllResourcesHaveImporters asserts that every resource exposed by
// the provider has an Importer with a State function wired. A typed-nil
// *schema.ResourceImporter satisfies interface{} != nil, so we check
// the concrete pointer directly.
func TestAllResourcesHaveImporters(t *testing.T) {
	cases := []struct {
		name     string
		importer *schema.ResourceImporter
	}{
		{"dme_domain", resourceDMEDomain().Importer},
		{"dme_template", resourceDMETemplate().Importer},
		{"dme_contact_list", resourceDMEContactList().Importer},
		{"dme_transfer_acl", resourceDMEACL().Importer},
		{"dme_custom_soa_record", resourceDmeCustomSoaRecord().Importer},
		{"dme_vanity_nameserver_record", resourceDmeVanityNameserverRecord().Importer},
		{"dme_secondary_dns", resourceDMESecondaryDNS().Importer},
		{"dme_secondary_ip_set", resourceDMESecondaryIPSet().Importer},
		{"dme_folder_record", resourceDmeFolder().Importer},
		{"dme_failover", resourceDMEFailover().Importer},
		{"dme_template_record", resourceDMETemplateRecord().Importer},
		{"dme_dns_record", resourceManagedDNSRecordActions().Importer},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.importer == nil {
				t.Errorf("%s: Importer is nil — import support not wired", tc.name)
			} else if tc.importer.State == nil {
				t.Errorf("%s: Importer.State is nil — no State function wired", tc.name)
			}
		})
	}
}
