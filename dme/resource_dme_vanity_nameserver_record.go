package dme

import (
	"fmt"
	"log"

	"github.com/DNSMadeEasy/dme-go-client/client"
	"github.com/DNSMadeEasy/dme-go-client/models"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDmeVanityNameserverRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceDmeVanityNameserverCreate,
		Read:   resourceDmeVanityNameserverRead,
		Update: resourceDmeVanityNameserverUpdate,
		Delete: resourceDmeVanityNameserverDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"servers": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"public_config": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"default_config": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"name_server_group_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
		},
	}
}

func resourceDmeVanityNameserverCreate(d *schema.ResourceData, m interface{}) error {
	dmeConnect := m.(*client.Client)

	vanityAttr := &models.Vanity{}

	if value, ok := d.GetOk("name"); ok {
		vanityAttr.Name = value.(string)
	}

	if value, ok := d.GetOk("servers"); ok {
		vanityAttr.Servers = toListOfString(value)
	}

	if value, ok := d.GetOk("public"); ok {
		vanityAttr.Public = value.(bool)
	}

	if value, ok := d.GetOk("default"); ok {
		vanityAttr.Default = value.(bool)
	}

	if value, ok := d.GetOk("name_server_group_id"); ok {
		vanityAttr.NameServerGroupID = value.(int)
	}

	cont, err := dmeConnect.Save(vanityAttr, "dns/vanity/")

	if err != nil {
		log.Println("Error returned: ", err)
		return err
	}

	log.Println("Value of container: ", cont)
	id := cont.S("id")
	log.Println("Id value: ", id)
	d.SetId(fmt.Sprintf("%v", id))
	return resourceDmeVanityNameserverRead(d, m)
}

func resourceDmeVanityNameserverRead(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	dn := d.Id()

	con, err := dmeClient.GetbyId("dns/vanity/" + dn)
	if err != nil {
		return err
	}

	d.Set("name", StripQuotes(con.S("name").String()))

	if raw := con.S("servers").Data(); raw != nil {
		if slice, ok := raw.([]interface{}); ok {
			servers := make([]string, len(slice))
			for i, s := range slice {
				servers[i] = s.(string)
			}
			d.Set("servers", servers)
		}
	}

	if b, ok := con.S("public").Data().(bool); ok {
		d.Set("public_config", b)
	} else {
		d.Set("public_config", StripQuotes(con.S("public").String()) == "true")
	}

	if b, ok := con.S("default").Data().(bool); ok {
		d.Set("default_config", b)
	} else {
		d.Set("default_config", StripQuotes(con.S("default").String()) == "true")
	}

	setIntField(d, "name_server_group_id", StripQuotes(con.S("nameServerGroupId").String()))

	return nil
}

func resourceDmeVanityNameserverUpdate(d *schema.ResourceData, m interface{}) error {
	dmeConnect := m.(*client.Client)

	vanityAttr := &models.Vanity{}

	if value, ok := d.GetOk("name"); ok {
		vanityAttr.Name = value.(string)
	}

	if value, ok := d.GetOk("servers"); ok {
		vanityAttr.Servers = toListOfString(value)
	}

	if value, ok := d.GetOk("public"); ok {
		vanityAttr.Public = value.(bool)
	}

	if value, ok := d.GetOk("default"); ok {
		vanityAttr.Default = value.(bool)
	}

	if value, ok := d.GetOk("name_server_group_id"); ok {
		vanityAttr.NameServerGroupID = value.(int)
	}

	log.Println("VNS structure is :", vanityAttr)
	dn := d.Id()

	_, err := dmeConnect.Update(vanityAttr, "dns/vanity/"+dn)
	if err != nil {
		return err
	}
	return resourceDmeVanityNameserverRead(d, m)
}

func resourceDmeVanityNameserverDelete(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	dn := d.Id()

	err := dmeClient.Delete("dns/vanity/" + dn)
	if err != nil {
		return nil
	}

	d.SetId("")
	return nil
}
