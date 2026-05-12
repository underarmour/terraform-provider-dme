---
layout: "dme"
page_title: "dme: dme_transfer_acl"
sidebar_current: "docs-dme-datasource-dme_transfer_acl"
description: |-
    Use this data source to retrieve a transfer ACL from the account.
---

# dme_transfer_acl #
Use this data source to retrieve a transfer ACL from the account.

# Example Usage #
```hcl
data "dme_transfer_acl" "first" {
  name = "transferacl"
}

```

## Argument Reference ##
* `name` - (Required) ACL Identifiable name.

## Attribute Reference ##
* `ips` - (Optional) The list of IP addresses defined in the ACL.

