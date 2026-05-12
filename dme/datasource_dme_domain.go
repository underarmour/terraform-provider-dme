package dme

import (
	"fmt"

	"github.com/DNSMadeEasy/dme-go-client/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceDMEDomain() *schema.Resource {
	return &schema.Resource{
		Read: datasourceDMEDomainRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"gtd_enabled": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"soa_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"template_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"vanity_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"transfer_acl_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"folder_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"updated": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"created": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func datasourceDMEDomainRead(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	name := d.Get("name").(string)

	con, err := dmeClient.GetbyId("dns/managed/")
	if err != nil {
		return err
	}

	data := con.S("data").Data().([]interface{})
	var flag bool
	var cnt int
	for _, info := range data {
		val := info.(map[string]interface{})
		if StripQuotes(val["name"].(string)) == name {
			flag = true
			break
		}
		cnt = cnt + 1
	}
	if flag != true {
		return fmt.Errorf("Domain of specified name not found")
	}

	dataCon := con.S("data").Index(cnt)
	d.SetId(dataCon.S("id").String())
	d.Set("name", extractField(dataCon.S("name")))
	d.Set("gtd_enabled", extractField(dataCon.S("gtdEnabled")))
	d.Set("soa_id", extractField(dataCon.S("soaId")))
	d.Set("template_id", extractField(dataCon.S("templateId")))
	d.Set("vanity_id", extractField(dataCon.S("vanityId")))
	d.Set("transfer_acl_id", extractField(dataCon.S("transferAclId")))
	d.Set("folder_id", extractField(dataCon.S("folderId")))
	d.Set("updated", extractField(dataCon.S("updated")))
	d.Set("created", extractField(dataCon.S("created")))
	return nil
}
