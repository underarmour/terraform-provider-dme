package dme

import (
	"fmt"

	"github.com/DNSMadeEasy/dme-go-client/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceDmeCustomSoaRecord() *schema.Resource {
	return &schema.Resource{
		Read:          datasourceConstellixDomainRead,
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"email": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"comp": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"ttl": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"refresh": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"serial": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"retry": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"expire": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"negative_cache": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func datasourceConstellixDomainRead(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	name := d.Get("name").(string)
	con, err := dmeClient.GetbyId("dns/soa/")
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
		return fmt.Errorf("SOA Record of specified name not found")
	}

	dataCon := con.S("data").Index(cnt)
	d.SetId(dataCon.S("id").String())

	d.Set("name", extractField(dataCon.S("name")))
	d.Set("email", extractField(dataCon.S("email")))
	d.Set("comp", extractField(dataCon.S("comp")))
	d.Set("ttl", extractField(dataCon.S("ttl")))
	d.Set("retry", extractField(dataCon.S("retry")))
	d.Set("refresh", extractField(dataCon.S("refresh")))
	d.Set("expire", extractField(dataCon.S("expire")))
	d.Set("serial", extractField(dataCon.S("serial")))
	d.Set("negative_cache", extractField(dataCon.S("negativeCache")))
	return nil

}
