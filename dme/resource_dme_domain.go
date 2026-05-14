package dme

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/DNSMadeEasy/dme-go-client/models"

	"github.com/DNSMadeEasy/dme-go-client/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDMEDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceDMEDomainCreate,
		Read:   resourceDMEDomainRead,
		Update: resourceDMEDomainUpdate,
		Delete: resourceDMEDomainDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"gtd_enabled": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
				Default:  "false",
			},

			"soa_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
			},

			"template_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
			},

			"vanity_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
			},

			"transfer_acl_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
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

func resourceDMEDomainCreate(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	domainAttr := &models.DomainAttribute{}

	if name, ok := d.GetOk("name"); ok {
		domainAttr.Name = name.(string)
	}

	if gtdEnabled, ok := d.GetOk("gtd_enabled"); ok {
		domainAttr.GtdEnabled = gtdEnabled.(string)
	}

	if soa, ok := d.GetOk("soa_id"); ok {
		domainAttr.SOAID = soa.(string)
	}

	if template, ok := d.GetOk("template_id"); ok {
		domainAttr.TemplateID = template.(string)
	}

	if vanity, ok := d.GetOk("vanity_id"); ok {
		domainAttr.VanityID = vanity.(string)
	}

	if transferaci, ok := d.GetOk("transfer_acl_id"); ok {
		domainAttr.TransferAClID = transferaci.(string)
	}

	if folder, ok := d.GetOk("folder_id"); ok {
		domainAttr.FolderID = folder.(string)
	}

	if updated, ok := d.GetOk("updated"); ok {
		domainAttr.Updated = updated.(string)
	}

	if created, ok := d.GetOk("created"); ok {
		domainAttr.Created = created.(string)
	}

	log.Println("domain structure is :", domainAttr)
	con, err := dmeClient.Save(domainAttr, "dns/managed/")
	if err != nil {
		return err
	}
	log.Println("Output containier create domain :", con.S("id"))
	d.SetId(fmt.Sprintf("%v", con.S("id")))
	return resourceDMEDomainRead(d, m)
}

func resourceDMEDomainUpdate(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	domainAttr := &models.DomainAttribute{}

	domainAttr.GtdEnabled = d.Get("gtd_enabled").(string)
	domainAttr.SOAID = d.Get("soa_id").(string)
	domainAttr.TemplateID = d.Get("template_id").(string)
	domainAttr.VanityID = d.Get("vanity_id").(string)
	domainAttr.TransferAClID = d.Get("transfer_acl_id").(string)

	if d.HasChange("folder_id") {
		domainAttr.FolderID = d.Get("folder_id").(string)
	}

	if d.HasChange("updated") {
		domainAttr.Updated = d.Get("updated").(string)
	}

	if d.HasChange("created") {
		domainAttr.Created = d.Get("created").(string)
	}

	log.Println("domain structure is :", domainAttr)
	dn := d.Id()
	_, err := dmeClient.Update(domainAttr, "dns/managed/"+dn)
	if err != nil {
		return err
	}
	return resourceDMEDomainRead(d, m)
}

func resourceDMEDomainRead(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	dn := d.Id()

	con, err := dmeClient.GetbyId("dns/managed/" + dn)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", con.S("id")))
	d.Set("name", StripQuotes(con.S("name").String()))
	d.Set("gtd_enabled", StripQuotes(con.S("gtdEnabled").String()))
	d.Set("soa_id", StripQuotes(con.S("soaId").String()))
	d.Set("template_id", StripQuotes(con.S("templateId").String()))
	d.Set("vanity_id", StripQuotes(con.S("vanityId").String()))
	d.Set("transfer_acl_id", StripQuotes(con.S("transferAclId").String()))
	d.Set("folder_id", StripQuotes(con.S("folderId").String()))
	d.Set("updated", StripQuotes(con.S("updated").String()))
	d.Set("created", StripQuotes(con.S("created").String()))

	return nil
}

func resourceDMEDomainDelete(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	dn := d.Id()

	// DME domains enter a "pending" state after create/delete operations and
	// cannot be deleted until they exit it. Poll with exponential backoff
	// (5s→10s→20s→40s, capped at 60s) to catch fast cases quickly while
	// staying well within the 150 req/5 min rate limit across parallel tests.
	// 30 attempts covers up to ~28 minutes in the worst case.
	var err error
	sleep := 5 * time.Second
	const maxSleep = 60 * time.Second
	const maxAttempts = 30
	for i := 0; i < maxAttempts; i++ {
		err = dmeClient.Delete("dns/managed/" + dn)
		if err == nil {
			break
		}
		if !strings.Contains(err.Error(), "pending") {
			break
		}
		log.Printf("[DEBUG] domain %s pending, retrying delete in %s (attempt %d/%d)", dn, sleep, i+1, maxAttempts)
		time.Sleep(sleep)
		if sleep < maxSleep {
			sleep *= 2
			if sleep > maxSleep {
				sleep = maxSleep
			}
		}
	}
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
