package dme

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	dmeClient "github.com/DNSMadeEasy/dme-go-client/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

// testRunSuffix is a 6-character random hex string unique to this test
// invocation. All domain-creating fixtures use it to avoid name collisions
// between concurrent or back-to-back runs.
var testRunSuffix string

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"dme": testAccProvider,
	}
}

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano()) //nolint:staticcheck
	testRunSuffix = fmt.Sprintf("%06x", rand.Intn(0xffffff))

	if os.Getenv("TF_ACC") != "" {
		if os.Getenv("DME_SKIP_DOMAIN_TESTS") == "" {
			fmt.Println()
			fmt.Println("+-----------------------------------------------------------------+")
			fmt.Println("|  DOMAIN TESTS ENABLED                                           |")
			fmt.Println("|                                                                 |")
			fmt.Println("|  DME does not guarantee timing for domain create/delete.        |")
			fmt.Println("|  Domain tests may take 15-30 minutes to clean up after          |")
			fmt.Println("|  themselves. This is normal; do not interrupt mid-run.          |")
			fmt.Println("|                                                                 |")
			fmt.Println("|  Set DME_SKIP_DOMAIN_TESTS=1 to skip these tests.              |")
			fmt.Println("+-----------------------------------------------------------------+")
			fmt.Println()
		}
		deleteStaleDomains()
	}

	os.Exit(m.Run())
}

// testDomain returns a domain name scoped to the current test run via
// testRunSuffix so that concurrent or back-to-back runs never collide.
// base should be a short label, e.g. "dom", "dns", "import-dom".
func testDomain(base string) string {
	return fmt.Sprintf("tf-acc-%s-%s.com", testRunSuffix, base)
}

// deleteStaleDomains lists all domains in the account and fires best-effort
// background deletes on any whose name starts with "tf-acc-". It does not
// block — DME domain deletes are async and may take up to 30 minutes.
// Domains still in a pending state will reject the delete; they will be
// retried on the next test run.
func deleteStaleDomains() {
	apiKey := os.Getenv("DME_API_KEY")
	secretKey := os.Getenv("DME_SECRET_KEY")
	if apiKey == "" || secretKey == "" {
		return
	}

	opts := []dmeClient.Option{}
	if baseURL := os.Getenv("DME_BASE_URL"); baseURL != "" {
		opts = append(opts, dmeClient.BaseURL(baseURL))
	}
	if os.Getenv("DME_INSECURE") == "true" {
		opts = append(opts, dmeClient.Insecure(true))
	}

	c := dmeClient.GetClient(apiKey, secretKey, opts...)

	con, err := c.GetbyId("dns/managed")
	if err != nil {
		log.Printf("[HYGIENE] failed to list domains: %v", err)
		return
	}

	domains, err := con.S("data").Children()
	if err != nil {
		log.Printf("[HYGIENE] failed to parse domain list: %v", err)
		return
	}

	var stale []string
	for _, d := range domains {
		name := StripQuotes(d.S("name").String())
		id := StripQuotes(d.S("id").String())
		if strings.HasPrefix(name, "tf-acc-") {
			stale = append(stale, id)
		}
	}

	if len(stale) == 0 {
		return
	}

	log.Printf("[HYGIENE] found %d stale test domain(s); firing background deletes", len(stale))
	for _, id := range stale {
		id := id
		go func() {
			if err := c.Delete("dns/managed/" + id); err != nil {
				log.Printf("[HYGIENE] delete domain %s: %v (may still be pending; will retry next run)", id, err)
			} else {
				log.Printf("[HYGIENE] delete domain %s: accepted", id)
			}
		}()
	}
	// Brief pause so goroutines can be scheduled before TestMain returns.
	time.Sleep(500 * time.Millisecond)
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

// testAccSkipIfDomainTestsDisabled skips the calling test when
// DME_SKIP_DOMAIN_TESTS is set. Domain create/delete in the DME API is
// asynchronous with no guaranteed timing (15-30 min observed in practice).
// CI sets this flag to keep the suite fast and reliable. Run the full suite
// manually or via the release-gate workflow before cutting a release.
func testAccSkipIfDomainTestsDisabled(t *testing.T) {
	t.Helper()
	if os.Getenv("DME_SKIP_DOMAIN_TESTS") != "" {
		t.Skip("skipping: DME_SKIP_DOMAIN_TESTS is set (domain lifecycle tests disabled)")
	}
}

// testAccSkipIfSandbox skips tests for features that are genuinely
// unsupported in the DME sandbox (returns HTML, requires production-only
// resource IDs). This is distinct from domain timing — these tests fail
// for API incompatibility reasons, not slow cleanup.
func testAccSkipIfSandbox(t *testing.T) {
	t.Helper()
	if strings.Contains(os.Getenv("DME_BASE_URL"), "sandbox") {
		t.Skip("skipping: test not supported against DME sandbox")
	}
}
