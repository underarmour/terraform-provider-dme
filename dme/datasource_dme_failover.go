package dme

import (
	"fmt"
	"log"

	"github.com/DNSMadeEasy/dme-go-client/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceDMEFailover() *schema.Resource {
	return &schema.Resource{
		Read: datasourceDMEFailoverRead,

		Schema: map[string]*schema.Schema{
			"record_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"monitor": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"system_description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"max_emails": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"sensitivity": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"protocol_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"port": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"failover": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"auto_failover": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"ip1": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"ip2": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"ip3": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"ip4": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"ip5": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"contact_list": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"http_fqdn": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"http_file": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"send_string": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"timeout": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"dns_timeout": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"dns_fqdn": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"http_query_string": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func datasourceDMEFailoverRead(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	con, err := dmeClient.GetbyId("monitor/" + d.Get("record_id").(string))
	if err != nil {
		return err
	}
	log.Println("Inside read, container value: ", con)

	d.SetId(fmt.Sprintf("%v", con.S("recordId")))
	d.Set("monitor", extractField(con.S("monitor")))
	d.Set("system_description", extractField(con.S("systemDescription")))
	d.Set("max_emails", extractField(con.S("maxEmails")))
	d.Set("sensitivity", extractField(con.S("sensitivity")))
	d.Set("protocol_id", extractField(con.S("protocolId")))
	d.Set("port", extractField(con.S("port")))
	d.Set("failover", extractField(con.S("failover")))
	d.Set("auto_failover", extractField(con.S("autoFailover")))
	d.Set("ip1", extractField(con.S("ip1")))
	d.Set("ip2", extractField(con.S("ip2")))
	d.Set("ip3", extractField(con.S("ip3")))
	d.Set("ip4", extractField(con.S("ip4")))
	d.Set("ip5", extractField(con.S("ip5")))
	d.Set("contact_list", extractField(con.S("contactListId")))
	d.Set("http_fqdn", extractField(con.S("httpFqdn")))
	d.Set("http_file", extractField(con.S("httpFile")))
	d.Set("http_query_string", extractField(con.S("httpQueryString")))
	d.Set("send_string", extractField(con.S("sendString")))
	d.Set("timeout", extractField(con.S("timeout")))
	// d.Set("dns_fqdn", extractField(con.S("dnsFqdn")))
	// d.Set("dns_timeout", extractField(con.S("dnsTimeout")))

	return nil

}
