# Fork Notes

Working notes for the Under Armour fork of `terraform-provider-dme`.
This document captures the *why* behind divergences from upstream,
the base commits we forked from, and what consumers of the fork
need to know. User-facing release notes live in
[`CHANGELOG.md`](CHANGELOG.md); this file is for fork maintainers
and contributors.

## Status

This fork exists because upstream `DNSMadeEasy/terraform-provider-dme`
carries read-path bugs that cause spurious plan drift on TXT, HTTPRED,
and MX records — bugs that are not safe to work around in consumer
config. Upstream has been active as recently as July 2025, but several
high-value fixes have not yet been accepted. It also has no import
support, blocking adoption of `terraform import` for existing DNS
infrastructure. Fixes here are intended for upstream contribution once
stabilized.

Changes from upstream v1.0.8:
- **Read-path drift fixed** for all 12 resources and all 12 data
  sources: TXT/SPF/CAA outer-quote stripping, long-TXT multi-string
  junction collapse, HTTPRED `&` HTML-escape correction, MX/CNAME
  case-insensitive comparison.
- **Import support** added for all 12 resources via `terraform import`.
- **43 unit tests** added; upstream shipped zero.

The fork has not yet been tagged or published to a registry. Consumers
use the filesystem mirror path described in the Consumer wiring section.

## Upstream lineage

- Upstream: https://github.com/DNSMadeEasy/terraform-provider-dme
- Forked from: upstream `master` at commit `c46a0e4` (upstream
  `v1.0.8` tag, 2025-06-02).
- License: Mozilla Public License 2.0. Original upstream MPL-licensed
  files retain MPL when modified; new files we add can be any
  compatible license. See [`LICENSE`](LICENSE).
- Source-availability obligation per MPL §3 triggers on **external**
  distribution only. Internal use within Under Armour is not
  "distribution" under MPL §1.4 / §3.2. A `NOTICE` file describing
  modifications per MPL §3.3 will land before any external release.

## Divergences from upstream

### Toolchain

- Bumped `go.mod` `go` directive from 1.23 to **1.26**. Upstream
  pinned Go 1.23, which went out of support when Go 1.25 shipped
  (Aug 2025). Pinning to current keeps the project building only on
  supported toolchains.
- Did **not** introduce a `toolchain` directive. Upstream removed
  theirs deliberately in v1.0.7 (commit `a1c0e50`) to avoid silent
  auto-downloads of newer Go from `proxy.golang.org`; we agree with
  that choice. Consumers building from source get a clear error if
  their installed Go is too old, with no surprise network fetches.
- Bumped `actions/setup-go` from `v2` (GitHub Actions deprecated,
  Node-12-based) to `v5` in the release workflow. Required for the
  workflow to reliably consume the new `go.mod` directive.

### Read-path bug fixes (resources and data sources)

Upstream uses `json.Marshal` via `container.String()` to extract field
values, then strips surrounding quotes with `StripQuotes`. This path
HTML-escapes `&` → `\u0026`, `<` → `\u003c`, `>` → `\u003e` in any
field value that contains those characters (HTTPRED URLs being the
common case), and handles TXT/SPF/CAA values incorrectly.

Fixes applied to all 12 resource Read functions and all 12 data source
Read functions:

- **HTTPRED `value`**: replaced `StripQuotes(x.String())` with
  `extractField(x)`, which bypasses `json.Marshal` and returns the raw
  Go string. `&` in redirect URLs no longer becomes `\u0026`.
- **TXT/SPF/CAA `value`**: DME wraps values in outer `"…"` on storage
  and splits values longer than 255 bytes with internal `""` junctions
  (RFC 1035 §3.3.14 multi-string form). The upstream code assumed a
  specific pre-2023 encoding that DME no longer produces. Replaced with
  `normalizeValueOnRead`, which strips the outer quotes and collapses
  `""` junctions into a single clean string regardless of encoding
  variant.
- **MX/CNAME/NS/ANAME `value` and record `name`**: DME canonicalizes
  these to lowercase on storage. Upstream compared them with `==`,
  causing spurious drift on any mixed-case input. Fixed with
  `DiffSuppressFunc` using `strings.EqualFold`.

The root cause of the data source bugs being missed in the initial fix:
resource and data source Read functions duplicate the same
field-population logic with no shared helper. A `populateXxxFromContainer`
refactor is tracked under Known limitations.

### Import support

Upstream never implemented `Importer` on any resource. `terraform import`
was entirely unsupported; attempting it returned an error.

All 12 resources now have import wired:

- **9 resources** (`dme_domain`, `dme_template`, `dme_contact_list`,
  `dme_transfer_acl`, `dme_custom_soa_record`,
  `dme_vanity_nameserver_record`, `dme_secondary_dns`,
  `dme_secondary_ip_set`, `dme_folder_record`): passthrough import —
  import ID is the resource's numeric DME ID.
- **`dme_dns_record`**: composite import ID `domain_id:record_id`.
  Read refactored to handle post-import state where `name` and `type`
  are not yet populated; falls back to listing all records in the domain
  and locating by ID via `findRecordByID`.
- **`dme_template_record`**: composite import ID `template_id:record_id`.
- **`dme_failover`**: import ID is the monitored record's DME ID
  (same as `record_id` in config).

### Housekeeping (no behavior impact)

- README rewritten to remove HashiCorp template chrome (Gitter,
  mailing list, logo, GOPATH-era build instructions) and modernize
  build/usage instructions. Fixed the provider example to match the
  current schema (`api_key`/`secret_key`, modern HCL).
- CHANGELOG reconstructed from upstream tag history (v0.1.0 through
  v1.0.8) under Keep-a-Changelog structure. Dropped the stale
  `0.2.0 (Unreleased)` upstream header.
- Removed dead files:
  - `.travis.yml` (Travis OSS shut down 2021; pinned Go 1.13.8).
  - `scripts/gogetcookie.sh` (HashiCorp-internal googlesource
    cookie hack; only used by `.travis.yml`).
  - `scripts/changelog-links.sh` (hard-coded
    `terraform-provider-datadog` URL; copy-paste artifact, never
    functional against this repo).
  - `main.tf` (developer scratch file with hard-coded template IDs
    accidentally committed).

## Test coverage reality

Upstream ships 26 test functions across 13 `_test.go` files.
**All 26 are acceptance tests** that skip unless `TF_ACC=1` is set
(they require a live DNS Made Easy account). The repository has
**zero pure unit tests**.

This means `make test` reports green by skipping everything for
the upstream-inherited suite. The fork's bug-fix work introduced the
first pure unit tests, following a strict failing-then-passing TDD
loop. Current unit test inventory (no `TF_ACC` required):

| File                                    | Tests | Covers                                                        |
|-----------------------------------------|-------|---------------------------------------------------------------|
| `dme/diff_suppress_test.go`             | 4     | DiffSuppressFunc semantics for `name` and `value`.            |
| `dme/read_extract_test.go`              | 7     | `extractField` bypassing the `json.Marshal` HTML-escape path. |
| `dme/value_normalize_test.go`           | 4     | TXT/SPF/CAA outer-quote strip and long-TXT junction collapse. |
| `dme/datasource_dme_dns_records_test.go`| 6     | Data source value normalization for TXT, SPF, long-TXT,       |
|                                         |       | HTTPRED, MX, and A via the same extractField + normalizeValueOnRead path. |
| `dme/dns_record_lookup_test.go`         | 12    | `findRecordByID` and `recordIDMatches` helpers covering       |
|                                         |       | float64/int/string ID types and edge cases.                   |
| `dme/import_dns_record_test.go`         | 4     | `parseDNSRecordImportID` composite ID parsing.                |
| `dme/import_template_record_test.go`    | 3     | `parseTemplateRecordImportID` composite ID parsing.           |
| `dme/import_failover_test.go`           | 2     | `importFailoverState` single-ID import.                       |
| `dme/importer_wiring_test.go`           | 1     | Structural: all 12 resources have `Importer.State` wired.     |

Total: 43 unit tests. Additionally, 12 `TestAccImport_*` acceptance
tests in `dme/import_acceptance_test.go` exercise import end-to-end
against a live DME account (require `TF_ACC=1`).

To exercise the existing acceptance suite against a real DME
account:

```sh
TF_ACC=1 DME_API_KEY=... DME_SECRET_KEY=... make testacc
```

## Build

Requires Go 1.26+ installed locally.

```sh
make build      # compile and install binary to $(go env GOPATH)/bin
make test       # unit tests (currently a no-op; see above)
make testacc    # acceptance tests; requires DME credentials
make vet        # go vet
make fmt        # gofmt
```

## Release pipeline

The pipeline is inherited from upstream v1.0.7 and updated for the
fork's toolchain:

- `.goreleaser.yml` cross-compiles for linux, darwin, windows, and
  freebsd across amd64, arm64, arm, and 386 (minus darwin/386,
  explicitly ignored). Produces zip archives plus a signed
  `SHA256SUMS` file matching the layout the OpenTofu and Terraform
  Registries expect.
- `.github/workflows/release.yml` triggers on `v*` tags and runs
  GoReleaser. GPG signing handled via
  `crazy-max/ghaction-import-gpg`, consuming `GPG_PRIVATE_KEY` and
  `PASSPHRASE` repo secrets; the import step exposes a fingerprint
  to GoReleaser via `GPG_FINGERPRINT`. GPG key generation and
  secret provisioning are separate human gestures performed before
  cutting the first real tag.
- Pipeline validation: `goreleaser release --snapshot --clean
  --skip=publish,sign` produces the full per-OS/arch artifact set
  in `dist/` locally for inspection without requiring a tag or
  signing key. Last dry-run-validated on the toolchain-bump
  commit; 13 archives + SHA256SUMS produced cleanly.

## Consumer wiring

Two distribution paths are supported by the same release pipeline;
the choice between them happens at tag time, not in the build.

**Public Registry path.** Once the fork repo is public and tagged,
register `underarmour/dme` in the OpenTofu Registry (and/or the
Terraform Registry) via the Registry's standard registration
process. From then on, new tags get indexed automatically.
Consumers set:

```hcl
terraform {
  required_providers {
    dme = {
      source  = "underarmour/dme"
      version = "~> 1.1"
    }
  }
}
```

**Filesystem mirror path** (for private / pre-public use). Same
release artifacts, but the consumer's CI downloads the right
OS/arch zip from the fork's GitHub Releases (with a PAT for
private-repo access) and lays it down at:

```
~/.terraform.d/plugins/registry.opentofu.org/underarmour/dme/<version>/<os>_<arch>/
```

OpenTofu's filesystem-mirror discovery resolves `source =
"underarmour/dme"` locally before reaching for the Registry. Same
`required_providers` declaration on the consumer side; the only
difference is the CI step that primes the mirror.

## Empirical DME API behavior

Settled by direct-REST PUT-then-GET probes against `api.dnsmadeeasy.com`
(no Terraform provider in the loop). Evidence captured in an internal
DNS IaC repository (`work/recon/` probe snapshots).

| Field                              | DME server-side behavior                                                                  |
|------------------------------------|-------------------------------------------------------------------------------------------|
| Record `name` (any type)           | Canonicalized to **lowercase** on storage. Mixed case in, lower case back.                |
| MX/CNAME/NS/ANAME `value` (target) | Canonicalized to **lowercase** on storage.                                                |
| TXT `value`                        | Wrapped with outer `"…"` on storage. Sent `v=spf1 -all`, GET returns `"v=spf1 -all"`.     |
| TXT `value` > 255 chars            | Split at 255-byte boundary with internal `""` junction, all wrapped in outer `"…"`. RFC 1035 §3.3.14 multi-string form. |
| HTTPRED `value`                    | Stored verbatim. `&` round-trips as `&`; literal `\u0026` round-trips as `\u0026`.        |

**Implication for the fork's v2.0.0 fixes:**

- DNS-name case "drift" is not data corruption; it's RFC 1035 §2.3.3
  case-insensitivity surfacing through canonicalization. The right fix
  is schema-level `DiffSuppressFunc` using `strings.EqualFold` on the
  affected attributes — not a read-path patch that "preserves" case the
  server didn't store.
- The original `docs/internal/fork-spec.md` framing ("DME stores mixed
  case") was unsupported by the recon evidence the spec itself cited.
  Corrected framing: lowercase return is canonicalization, the fix is
  case-insensitive comparison.
- Cat 2 (HTTPRED `&` → `\u0026`) and cat 4/5b (TXT outer `"…"` and
  multi-string junctions) remain unambiguous provider-side read-path
  bugs and are patched directly.

## Known limitations and deferred work

### Populate-helper refactor (aspirational)

Resource Read functions and data source Read functions duplicate the
same field-population logic. A change to `normalizeValueOnRead` or
`extractField` must currently be applied in two places, which is how
the original drift-fix (resources only, not data sources) went
unnoticed. The right fix is a shared `populateXxxFromContainer`
function called by both paths, matching the pattern used in the
Terraform AWS provider's `flattenXxx` helpers. Deferred intentionally:
a structural refactor of this kind increases the diff against upstream,
making the fixes harder to evaluate and merge. If upstream accepts the
bug fixes first, this refactor is more appropriate as a follow-on PR
against their tree.

## v2.0.0 preparation reminders

Items to revisit when prepping the v2.0.0 release (post-bug-fix work).
Captured here so they don't slip through the cracks at release time.

- **README "drop-in replacement" language.** The fork callout at the
  top of `README.md` asserts drop-in compatibility with upstream.
  That claim is correct for v1.1.0 but breaks the moment v2.0.0
  ships read-path behavior changes. At v2.0.0 cut, rewrite the
  callout to describe v2.x as a corrected-behavior replacement (not
  drop-in) with a pointer to v1.1.x for consumers who want the
  unchanged-behavior line.
- **CHANGELOG breaking-change section.** v2.0.0 needs a clear
  `### Breaking changes` subsection listing every read-path
  semantic that shifted, plus any schema-shape changes from
  adding import support.
- **State upgrader.** If any read-path fix shifts the on-disk state
  shape (most likely for long-TXT records, where the multi-string
  wire encoding may normalize differently in state), schema version
  + state upgrader required so consumers don't see destructive
  plans on first apply post-upgrade.
