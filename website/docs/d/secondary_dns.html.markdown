---
layout: "dme"
page_title: "dme: dme_secondary_dns"
sidebar_current: "docs-dme-datasource-dme_secondary_dns"
description: |-
  Use this data source to retrieve a secondary DNS configuration from the account.
---

# dme_secondary_dns #
Use this data source to retrieve a secondary DNS configuration from the account.

## Example Usage ##

```hcl
data "dme_secondary_dns" "example" {
  name        = "example.com"
}

```

## Argument Reference ##
* `name` - (Required) Name of Secondary DNS action. Name should be unique.

## Attribute Reference ##
* `name` - (Required) Name of domain action. Name should be unique.
* `ipset_id` - id of the Secondary Ip set which is associated with the secondary DNS.
* `folder_id` - id of the Folder record which is associated with the secondary DNS.
* `id` - Set to the dme calculated id of secondary DNS action.