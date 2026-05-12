# DNSMadeEasy Terraform Provider (Under Armour fork)

Terraform / OpenTofu provider for managing DNS Made Easy resources via the
[DME REST API v2.0](https://api-docs.dnsmadeeasy.com/).

This is an Under Armour maintained fork of
[`DNSMadeEasy/terraform-provider-dme`](https://github.com/DNSMadeEasy/terraform-provider-dme).
It fixes long-standing read-path drift bugs and adds import support
for all resources, while maintaining the same resource types, schema,
and API behavior as upstream. Switching forks requires
exactly one HCL change: `source = "underarmour/dme"`. See
[`FORK_NOTES.md`](FORK_NOTES.md) for details on what diverges from
upstream and why.

## Requirements

- Terraform 0.12+ or OpenTofu 1.x
- Go 1.26+ (for building from source)
- A DNS Made Easy account with API access (an API key and secret key)

## Building

```sh
git clone https://github.com/underarmour/terraform-provider-dme.git
cd terraform-provider-dme
make build
```

The binary is installed to `$(go env GOPATH)/bin/terraform-provider-dme`.

## Usage

```hcl
terraform {
  required_providers {
    dme = {
      source = "underarmour/dme"
    }
  }
}

provider "dme" {
  api_key    = var.dme_api_key
  secret_key = var.dme_secret_key

  # Optional
  # insecure  = false
  # proxy_url = "https://proxy.example.com:8080"
  # base_url  = "https://api.sandbox.dnsmadeeasy.com/V2.0"
}

resource "dme_domain" "example" {
  name = "example.com"
}

resource "dme_dns_record" "www" {
  domain_id = dme_domain.example.id
  name      = "www"
  type      = "A"
  value     = "192.0.2.1"
  ttl       = 300
}
```

Credentials may also be supplied via environment variables:

| Variable         | Purpose                  |
|------------------|--------------------------|
| `DME_API_KEY`    | DNS Made Easy API key    |
| `DME_SECRET_KEY` | DNS Made Easy secret key |

### Rate limiting

The DME API enforces a 150-request / 5-minute sliding window. The upstream
provider has no retry/backoff logic, so concurrent applies routinely exhaust
the budget. The standard workaround is to run with `-parallelism=1`:

```sh
terraform plan  -parallelism=1
terraform apply -parallelism=1
```

## Resources and data sources

Resources: `dme_domain`, `dme_dns_record`, `dme_custom_soa_record`,
`dme_template`, `dme_template_record`, `dme_vanity_nameserver_record`,
`dme_transfer_acl`, `dme_secondary_dns`, `dme_secondary_ip_set`,
`dme_failover`, `dme_folder_record`, `dme_contact_list`.

Each resource is also exposed as a data source of the same name.

## Development

Standard Go module project — clone anywhere on disk, no `$GOPATH/src/...`
layout required.

```sh
make build      # compile and install the provider binary
make test       # unit tests
make testacc    # acceptance tests (hits the live DME API; requires credentials)
make fmt        # gofmt
make vet        # go vet
```

Acceptance tests require `DME_API_KEY` and `DME_SECRET_KEY` to be set and
will create and destroy real records against the configured account.

## License

Mozilla Public License 2.0. See [LICENSE](LICENSE).
