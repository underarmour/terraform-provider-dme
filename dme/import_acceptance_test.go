package dme

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Import acceptance tests. These require TF_ACC=1 and valid DME
// credentials (apikey / secretkey env vars). Each test creates a
// resource via the provider, then imports it by the documented ID
// format and verifies that the imported state matches the created state.
//
// For composite-ID resources (dme_dns_record, dme_template_record) an
// ImportStateIdFunc assembles the composite from live state. For
// dme_failover, the import ID is d.Id() which equals record_id, so no
// custom function is needed.

// ---------------------------------------------------------------------------
// Simple passthrough imports (numeric ID only)
// ---------------------------------------------------------------------------

func TestAccImport_Domain(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDMEDomainConfig_basic("import-test-domain.com", "false"),
			},
			{
				ResourceName:      "dme_domain.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccImport_Template(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: testAccCheckDMETemplateConfig_basic("import-test-template")},
			{
				ResourceName:      "dme_template.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccImport_ContactList(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: testAccCheckDMEContactListConfig_basic("import-test@example.com")},
			{
				ResourceName:      "dme_contact_list.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccImport_TransferACL(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: testAccCheckDMEACLConfig_basic("import-test-acl", "1.2.3.4")},
			{
				ResourceName:      "dme_transfer_acl.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccImport_CustomSOA(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: testAccCheckDmeSOAConfig_basic("import-test-soa")},
			{
				ResourceName:      "dme_custom_soa_record.soacheck",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccImport_VanityNameserver(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: testAccCheckDmeVNSConfig_basic("import-test-vanity")},
			{
				ResourceName:      "dme_vanity_nameserver_record.vanityrecord",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccImport_SecondaryDNS(t *testing.T) {
	testAccSkipIfSandbox(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: testAccCheckDMESecondaryDNSConfig_basic("12345")},
			{
				ResourceName:      "dme_secondary_dns.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccImport_SecondaryIPSet(t *testing.T) {
	testAccSkipIfSandbox(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: testAccCheckDMESecondaryIPSetConfig_basic("1.2.3.4")},
			{
				ResourceName:      "dme_secondary_ip_set.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccImport_FolderRecord(t *testing.T) {
	testAccSkipIfSandbox(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: testAccCheckDmeFolderConfig_basic("import-test-folder")},
			{
				ResourceName:      "dme_folder_record.folderrecord",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Composite ID imports
// ---------------------------------------------------------------------------

func TestAccImport_DNSRecord(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: testAccCheckDMERecordConfig_basic("86400")},
			{
				ResourceName:      "dme_dns_record.a1",
				ImportState:       true,
				ImportStateVerify: true,
				// Assemble the composite import ID from live state.
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					domain, ok := s.RootModule().Resources["dme_domain.domain1"]
					if !ok {
						return "", fmt.Errorf("dme_domain.domain1 not found in state")
					}
					record, ok := s.RootModule().Resources["dme_dns_record.a1"]
					if !ok {
						return "", fmt.Errorf("dme_dns_record.a1 not found in state")
					}
					return domain.Primary.ID + ":" + record.Primary.ID, nil
				},
				// Optional/Computed fields that may be set server-side to non-empty
				// defaults not present in the minimal test config. Add here if the
				// acceptance run surfaces mismatches on specific attributes.
				ImportStateVerifyIgnore: []string{"gtd_location"},
			},
		},
	})
}

func TestAccImport_TemplateRecord(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: testAccCheckDMETemplateRecordConfig_basic("86400")},
			{
				ResourceName:      "dme_template_record.a1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					tmpl, ok := s.RootModule().Resources["dme_template.template1"]
					if !ok {
						return "", fmt.Errorf("dme_template.template1 not found in state")
					}
					record, ok := s.RootModule().Resources["dme_template_record.a1"]
					if !ok {
						return "", fmt.Errorf("dme_template_record.a1 not found in state")
					}
					return tmpl.Primary.ID + ":" + record.Primary.ID, nil
				},
				ImportStateVerifyIgnore: []string{"gtd_location"},
			},
		},
	})
}

func TestAccImport_Failover(t *testing.T) {
	testAccSkipIfSandbox(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: testAccCheckDMEFailoverConfig_basic("1.2.3.4")},
			{
				ResourceName:      "dme_failover.a1",
				ImportState:       true,
				ImportStateVerify: true,
				// Import ID = d.Id() = recordId, which equals record_id in
				// state. No custom IdFunc needed; SDK uses d.Id() by default.
			},
		},
	})
}
