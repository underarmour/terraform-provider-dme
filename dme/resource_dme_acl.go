package dme

import (
	"fmt"
	"log"
	"reflect"
	"sort"

	"github.com/DNSMadeEasy/dme-go-client/models"

	"github.com/DNSMadeEasy/dme-go-client/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDMEACL() *schema.Resource {
	return &schema.Resource{
		Create: resourceDMEACLCreate,
		Read:   resourceDMEACLRead,
		Update: resourceDMEACLUpdate,
		Delete: resourceDMEACLDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"ips": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDMEACLCreate(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	aclAttr := models.ACLAttribute{}

	if name, ok := d.GetOk("name"); ok {
		aclAttr.Name = name.(string)
	}

	if ips, ok := d.GetOk("ips"); ok {
		aclAttr.Ips = toListOfString(ips)
	}

	log.Println("Inside create: ", aclAttr)

	con, err := dmeClient.Save(&aclAttr, "dns/transferAcl/")
	if err != nil {
		return err
	}
	log.Println("Output containier create domain :", con.S("id"))
	d.SetId(fmt.Sprintf("%v", con.S("id")))

	return resourceDMEACLRead(d, m)

}

func resourceDMEACLRead(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	dn := d.Id()
	con, err := dmeClient.GetbyId("dns/transferAcl/" + dn)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", con.S("id")))
	d.Set("name", StripQuotes(con.S("name").String()))

	ips := con.S("ips").Data().([]interface{})
	listips := make([]string, 0)

	for _, id := range ips {
		listips = append(listips, id.(string))
	}

	// Preserve the user-specified ordering when the content matches what the
	// API returned; otherwise (including post-import when there is no prior
	// state) write the API-returned list directly.
	listget := make([]string, 0)
	if ips, ok := d.GetOk("ips"); ok {
		listget = toListOfString(ips)
	}

	sortedGet := append([]string(nil), listget...)
	sortedAPI := append([]string(nil), listips...)
	sort.Strings(sortedGet)
	sort.Strings(sortedAPI)

	if len(listget) > 0 && reflect.DeepEqual(sortedGet, sortedAPI) {
		d.Set("ips", listget)
	} else {
		d.Set("ips", listips)
	}
	return nil
}

func resourceDMEACLUpdate(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	aclAttr := &models.ACLAttribute{}

	if name, ok := d.GetOk("name"); ok {
		aclAttr.Name = name.(string)
	}

	if ips, ok := d.GetOk("ips"); ok {
		aclAttr.Ips = toListOfString(ips)
	}

	log.Println("domain structure is :", aclAttr)
	dn := d.Id()
	_, err := dmeClient.Update(aclAttr, "dns/transferAcl/"+dn)
	if err != nil {
		return err
	}

	return resourceDMEACLRead(d, m)
}

func resourceDMEACLDelete(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	dn := d.Id()

	err := dmeClient.Delete("dns/transferAcl/" + dn)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
