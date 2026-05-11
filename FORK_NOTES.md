# Fork Notes

Working notes for the Under Armour fork of `terraform-provider-dme`.
This document captures the *why* behind divergences from upstream,
the base commits we forked from, and what consumers of the fork
need to know. User-facing release notes live in
[`CHANGELOG.md`](CHANGELOG.md); this file is for fork maintainers
and contributors.

## Status

WIP. The fork has not yet been tagged or published. The first
intended release is **v1.1.0** — a drop-in replacement for upstream
v1.0.8 with housekeeping and toolchain modernization but no
behavior changes. Behavior changes (read-path bug fixes) ship in
**v2.0.0**.

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

This means `make test` reports green by skipping everything. The
fork's read-path bug-fix work will introduce the first pure unit
tests in this provider, with one failing-then-passing test per
bug as the standard for accepting a fix.

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

## Deferred / follow-on work

Populated as roadmap items get explicitly deferred during the fork
work. Each entry records the item, the deferral reason, and the
release it's expected to land in (or "follow-on, no committed
target").

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
