package dme

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"dme": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("DME_API_KEY"); v == "" {
		t.Fatal("DME_API_KEY must be set for acceptance tests")
	}

	if v := os.Getenv("DME_SECRET_KEY"); v == "" {
		t.Fatal("DME_SECRET_KEY must be set for acceptance tests")
	}
}

// testAccSkipIfSandbox skips the calling test when running against the DME
// sandbox. Some features (Failover, SecondaryIPSet) and tests that rely on
// production-specific resource IDs are not usable in the sandbox environment.
func testAccSkipIfSandbox(t *testing.T) {
	t.Helper()
	if strings.Contains(os.Getenv("DME_BASE_URL"), "sandbox") {
		t.Skip("skipping: test not supported against DME sandbox")
	}
}
