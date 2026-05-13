# Changelog

All notable changes to this provider are documented here.

Entries for v0.1.0 through v1.0.8 reflect upstream
`DNSMadeEasy/terraform-provider-dme` history. The primary source is the
upstream GitHub Release body for each tag; where the upstream Release
body was empty, a placeholder, or absent, the entry is reconstructed
from upstream git history and flagged. Dates are the upstream Release
`published_at` value where available, otherwise the tag commit date.

## Unreleased

### Added

- All optional provider configuration fields (`insecure`, `proxy_url`,
  `base_url`) can now be set via environment variables (`DME_INSECURE`,
  `DME_PROXY_URL`, `DME_BASE_URL`). The required fields (`api_key`,
  `secret_key`) already supported `DME_API_KEY` and `DME_SECRET_KEY`;
  this brings the optional fields in line with the same convention.
- Import support (`tofu import` / `terraform import`) for all twelve
  resources: `dme_domain`, `dme_dns_record`, `dme_template`,
  `dme_template_record`, `dme_contact_list`, `dme_transfer_acl`,
  `dme_custom_soa_record`, `dme_vanity_nameserver_record`,
  `dme_secondary_dns`, `dme_secondary_ip_set`, `dme_failover`,
  `dme_folder_record`. Import ID formats:
  - `dme_dns_record`: `<domain_id>:<record_id>`
  - `dme_template_record`: `<template_id>:<record_id>`
  - `dme_failover`: `<record_id>` (the DNS record ID with a failover
    monitor attached)
  - All others: the resource's numeric ID as shown in the DME console.

## v1.1.1 — 2026-05-12

### Upgrade notes

The fixes below correct what the provider reads back from the DME API.
If your configuration was written correctly (TXT values without outer
quotes, HTTPRED URLs with literal `&`), plans will simply go clean
after upgrading — no config changes needed.

If you worked around these bugs in config, you will see inverse drift
on first plan after upgrading:

- **TXT/SPF/CAA workaround:** if `value` was written as
  `"\"v=spf1 -all\""` (with escaped outer quotes) to match what the
  provider was reading back, remove the outer quotes. The correct
  form is `"v=spf1 -all"`.
- **HTTPRED workaround:** if `value` was written with `\u0026` instead
  of `&` to match what the provider was reading back, replace it with
  a literal `&`.

In both cases, run `terraform plan` after upgrading to identify
affected resources, update config to the correct form, and the plan
will go clean.

### Fixed

- `dme_dns_record` no longer reports spurious drift when the only
  difference between configuration and state is letter case on
  `name` or on MX/CNAME/NS/ANAME `value`. DME canonicalizes these
  to lowercase on storage (RFC 1035 §2.3.3 case-insensitivity);
  comparison is now case-insensitive via `DiffSuppressFunc`.
- `dme_dns_record` TXT/SPF/CAA `value` no longer carries the outer
  `"..."` wrapping or internal `""` multi-string junctions that DME
  adds on storage. Authoring the value with or without outer quotes
  is equivalent.
- `dme_dns_record` HTTPRED `value` no longer has `&` rewritten to
  `\u0026` on Read.
- `dme_dns_record` and `dme_template_record` **data sources** had the
  same TXT/SPF/CAA, MX casing, and HTTPRED escaping bugs as their
  resource counterparts. Both now use `extractField` and
  `normalizeValueOnRead` for consistent output. All other data sources
  also switched from the `json.Marshal`-based `StripQuotes` path to
  `extractField` to prevent HTML-entity leakage on fields containing
  `&`, `<`, or `>`.

## v1.1.0 — 2026-05-11

First release of the Under Armour fork. Drop-in replacement for
upstream `DNSMadeEasy/dme` v1.0.8: same resource types, same schema,
identical runtime behavior. The only consumer-side change is updating
the `source` declaration in `required_providers` from
`DNSMadeEasy/dme` to `underarmour/dme`.

### Added
- `NOTICE` file declaring fork modifications per MPL 2.0 §3.3.
- `FORK_NOTES.md` documenting upstream lineage, toolchain rationale,
  test-coverage reality, release pipeline, consumer wiring paths,
  and a pre-release checklist.
- README fork callout linking to FORK_NOTES.md.

### Changed
- **Module path renamed** from `github.com/terraform-providers/terraform-provider-dme`
  to `github.com/underarmour/terraform-provider-dme`. Contributors
  building from source must use the new import path.
- **Registry source** is `underarmour/dme` (upstream remains
  `DNSMadeEasy/dme`).
- Bumped minimum Go version to **1.26** (`go.mod` `go` directive).
  Go 1.23 went out of support when Go 1.25 shipped (Aug 2025);
  pinning to current keeps the project building on supported
  toolchains and removes ambiguity about which Go features the
  provider may use.
- Release workflow now uses `actions/setup-go@v5` with Go 1.26
  (upgraded from `@v2`, which is GitHub Actions deprecated).
- README rewritten to remove HashiCorp template chrome and modernize
  build/usage instructions; example provider block fixed to match
  current schema (`api_key`/`secret_key`, modern HCL).
- CHANGELOG built from upstream GitHub Release bodies for v0.1.0
  through v1.0.8, with git-history fallback flagged in-line for tags
  where upstream Release bodies were empty or absent. Keep-a-Changelog
  section structure throughout.

### Removed
- `.travis.yml` (Travis OSS shut down 2021; pinned obsolete tooling).
- `scripts/gogetcookie.sh` (HashiCorp-internal googlesource cookie
  hack; only used by `.travis.yml`).
- `scripts/changelog-links.sh` (hard-coded wrong provider URL;
  copy-paste artifact, never functional against this repo).
- `main.tf` (developer scratch file with hard-coded template IDs
  accidentally committed).

## v1.0.8 — 2025-06-26

_Source: upstream GitHub Release notes._

### Changed

- **Resource Behavior Updates:** Updated the `dme_domain` resource to
  improve user control and predictability for several optional fields:
  - Fields `gtd_enabled`, `soa_id`, `template_id`, `vanity_id`, and
    `transfer_acl_id` are no longer `Computed`. Their values are now
    determined solely by user configuration or explicit defaults.
  - `gtd_enabled` now defaults to `"false"` if omitted.
  - If optional IDs (`soa_id`, `template_id`, `vanity_id`,
    `transfer_acl_id`, `folder_id`) are not set, the provider will not
    assign custom values, resulting in default platform behavior
    (e.g., no custom SOA, template, or vanity nameservers).
  - During updates, these fields are always sent explicitly, ensuring
    that removal or resetting is reflected in DNS Made Easy.
- **Improved Consistency:** These changes ensure that resource state in
  Terraform more reliably matches user configuration, reducing the
  potential for state drift or unexpected diffs caused by external
  changes.

### Docs

- Enhanced documentation for `dme_domain` to clarify the default
  behavior for all optional arguments.

### Migration Note

- **Potential Configuration Churn:** Users upgrading from previous
  versions may notice that Terraform now manages these fields more
  strictly. If a field was previously managed implicitly (via remote
  state or platform defaults), users may need to review and update
  their configurations to ensure intended behavior.

Related PR: [DNSMadeEasy/terraform-provider-dme#47](https://github.com/DNSMadeEasy/terraform-provider-dme/pull/47).

## v1.0.7 — 2025-02-05

_Source: upstream GitHub Release notes._

### Added

- Optional config parameter to specify a custom API server base URL
  (DNS Made Easy sandbox or customer-specific endpoint), used instead
  of the default API server.

### Changed

- Improved validation and plumbing of other config parameters.
- Improved existing documentation terminology and formatting.

## v1.0.6 — 2022-10-04

_Source: git history; upstream Release body was a tag-name placeholder._

### Changed

- Dependency update; bumped `dme-go-client` (v1.11.2 → v1.11.3).
- Improved rate-limiter behavior.
- Adopted Terraform provider environment variable conventions
  (`DME_API_KEY`, `DME_SECRET_KEY`) — fixes #19. (#40)

## v1.0.5 — 2022-03-28

_Source: upstream GitHub Release notes._

### Fixed

- Compatibility issue with M1 MacBooks (`darwin_arm64`).

## v1.0.4 — 2021-11-16

_Source: upstream GitHub Release notes._

### Fixed

- API rate-limit issue.

## v1.0.3 / v0.1.3 — 2021-10-26 / 2021-03-10

_Source: git history; upstream Release bodies were empty for both tags.
`v0.1.3` and `v1.0.3` point to the same commit (PR #28); upstream
retagged 7 months later when the version scheme switched to `v1.x`._

### Fixed

- `dme_dns_record` create failing when `name` was empty (apex record).
  (#28)

## v0.1.2 — 2020-09-18

_Source: git history; upstream Release body was a tag-name placeholder._

### Added

- `.goreleaser.yml` and GitHub release workflow.

## v0.1.1 — 2020-07-06

_Source: git history; no upstream GitHub Release was published for this tag._

### Changed

- Provider recreated against the modern Terraform plugin SDK with the
  standard implementation layout.
- Vendored dependencies refreshed.
- Acceptance tests added.

## v0.1.0 — 2017-06-20

_Source: git history; no upstream GitHub Release was published for this tag._

### Added

- Initial standalone release. Same functionality as the in-tree
  `terraform-providers/dme` provider in Terraform 0.9.8, repackaged per
  the
  [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/).
