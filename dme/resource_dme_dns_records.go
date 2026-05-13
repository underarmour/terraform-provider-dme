package dme

import (
	"fmt"
	"log"
	"strings"

	"github.com/DNSMadeEasy/dme-go-client/client"
	"github.com/DNSMadeEasy/dme-go-client/models"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	// "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	// "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceManagedDNSRecordActions() *schema.Resource {
	return &schema.Resource{
		Create: resourceManagedDNSRecordActionsCreate,
		Update: resourceManagedDNSRecordActionsUpdate,
		Read:   resourceManagedDNSRecordActionsRead,
		Delete: resourceManagedDNSRecordActionsDelete,

		Importer: &schema.ResourceImporter{
			State: importDNSRecordState,
		},

		Schema: map[string]*schema.Schema{
			"domain_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"value": &schema.Schema{
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressCaseInsensitiveDNSValue,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressCaseInsensitiveDNSName,
			},

			"dynamic_dns": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"password": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"ttl": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			// "monitor": &schema.Schema{
			// 	Type:     schema.TypeBool,
			// 	Optional: true,
			// 	Computed: true,
			// },

			// "failover": &schema.Schema{
			// 	Type:     schema.TypeBool,
			// 	Optional: true,
			// 	Computed: true,
			// },

			// "failed": &schema.Schema{
			// 	Type:     schema.TypeBool,
			// 	Optional: true,
			// 	Computed: true,
			// },

			"gtd_location": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"caa_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"issuer_critical": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"keywords": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"title": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"redirect_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"hardlink": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"mx_level": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"weight": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"priority": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"port": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceManagedDNSRecordActionsCreate(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)

	recordAttr := models.ManagedDNSRecordActions{}

	recordAttr.Name = d.Get("name").(string)
	// if name, ok := d.GetOk("name"); ok {
	// 	recordAttr.Name = name.(string)
	// }

	if value, ok := d.GetOk("value"); ok {
		recordAttr.Value = value.(string)
	}

	if Type, ok := d.GetOk("type"); ok {
		recordAttr.Type = Type.(string)
	}

	if dynamicdns, ok := d.GetOk("dynamic_dns"); ok {
		recordAttr.DynamicDNS = dynamicdns.(string)
	}

	if password, ok := d.GetOk("password"); ok {
		recordAttr.Password = password.(string)
	}

	if ttl, ok := d.GetOk("ttl"); ok {
		recordAttr.Ttl = ttl.(string)
	}

	if gtdlocation, ok := d.GetOk("gtd_location"); ok {
		recordAttr.GtdLocation = gtdlocation.(string)
	}

	if description, ok := d.GetOk("description"); ok {
		recordAttr.Description = description.(string)
	}

	if keywords, ok := d.GetOk("keywords"); ok {
		recordAttr.Keywords = keywords.(string)
	}

	if title, ok := d.GetOk("title"); ok {
		recordAttr.Title = title.(string)
	}

	if redirecttype, ok := d.GetOk("redirect_type"); ok {
		recordAttr.RedirectType = redirecttype.(string)
	}

	if hardlink, ok := d.GetOk("hardlink"); ok {
		recordAttr.HardLink = hardlink.(string)
	}

	if mxlevel, ok := d.GetOk("mx_level"); ok {
		recordAttr.MxLevel = mxlevel.(string)
	}

	if weight, ok := d.GetOk("weight"); ok {
		recordAttr.Weight = weight.(string)
	}

	if priority, ok := d.GetOk("priority"); ok {
		recordAttr.Priority = priority.(string)
	}

	if port, ok := d.GetOk("port"); ok {
		recordAttr.Port = port.(string)
	}

	if caatype, ok := d.GetOk("caa_type"); ok {
		recordAttr.CaaType = caatype.(string)
	}
	if issuer, ok := d.GetOk("issuer_critical"); ok {
		recordAttr.IssuerCritical = issuer.(string)
	}
	log.Println("Value of recordAttr: ", &recordAttr)

	cont, err := dmeClient.Save(&recordAttr, "dns/managed/"+d.Get("domain_id").(string)+"/records/")

	if err != nil {
		log.Println("Error returned: ", err)
		return err
	}

	log.Println("Value of container: ", cont)
	idname := cont.S("name").String()
	if strings.HasPrefix(idname, "\"") && strings.HasSuffix(idname, "\"") {
		idname = strings.TrimSuffix(strings.TrimPrefix(idname, "\""), "\"")
	}
	log.Println("Idname value inside create: ", idname)
	log.Println("Id valueinside create: ", cont.S("id"))
	d.Set("name", fmt.Sprintf("%v", idname))
	d.SetId(fmt.Sprintf("%v", cont.S("id")))

	return resourceManagedDNSRecordActionsRead(d, m)
}

func resourceManagedDNSRecordActionsRead(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	dnsId := d.Id()
	domainID := d.Get("domain_id").(string)
	name := d.Get("name").(string)
	recordType := d.Get("type").(string)

	// When name and type are empty the resource was just imported and the
	// Importer only had the composite ID to work with. Fall back to listing
	// all records in the domain and locating by numeric ID.
	endpoint := "dns/managed/" + domainID + "/records"
	if name != "" && recordType != "" {
		endpoint += "?recordName=" + name + "&type=" + recordType
	}

	con, err := dmeClient.GetbyId(endpoint)
	if err != nil {
		return err
	}

	cont1 := findRecordByID(con, dnsId)
	if cont1 == nil {
		log.Printf("[WARN] DME record %s not found in domain %s; removing from state", dnsId, domainID)
		d.SetId("")
		return nil
	}

	d.SetId(fmt.Sprintf("%v", cont1.S("id").String()))
	recordType = extractField(cont1.S("type"))
	d.Set("name", extractField(cont1.S("name")))
	d.Set("value", normalizeValueOnRead(recordType, extractField(cont1.S("value"))))
	d.Set("type", recordType)
	d.Set("dynamic_dns", extractField(cont1.S("dynamicDns")))
	d.Set("password", extractField(cont1.S("password")))
	d.Set("ttl", extractField(cont1.S("ttl")))
	d.Set("gtd_location", extractField(cont1.S("gtdLocation")))
	d.Set("description", extractField(cont1.S("description")))
	d.Set("keywords", extractField(cont1.S("keywords")))
	d.Set("title", extractField(cont1.S("title")))
	d.Set("redirect_type", extractField(cont1.S("redirectType")))
	d.Set("hardlink", extractField(cont1.S("hardLink")))
	d.Set("mx_level", extractField(cont1.S("mxLevel")))
	d.Set("weight", extractField(cont1.S("weight")))
	d.Set("port", extractField(cont1.S("port")))
	d.Set("priority", extractField(cont1.S("priority")))
	d.Set("caa_type", extractField(cont1.S("caaType")))
	d.Set("issuer_critical", extractField(cont1.S("issuerCritical")))

	return nil
}

func resourceManagedDNSRecordActionsUpdate(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	recordAttr := models.ManagedDNSRecordActions{}

	if name, ok := d.GetOk("name"); ok {
		recordAttr.Name = name.(string)
	}

	if value, ok := d.GetOk("value"); ok {
		recordAttr.Value = value.(string)
	}

	if Type, ok := d.GetOk("type"); ok {
		recordAttr.Type = Type.(string)
	}

	if dynamicdns, ok := d.GetOk("dynamic_dns"); ok {
		recordAttr.DynamicDNS = dynamicdns.(string)
	}

	if password, ok := d.GetOk("password"); ok {
		recordAttr.Password = password.(string)
	}

	if ttl, ok := d.GetOk("ttl"); ok {
		recordAttr.Ttl = ttl.(string)
	}

	if gtdlocation, ok := d.GetOk("gtd_location"); ok {
		recordAttr.GtdLocation = gtdlocation.(string)
	}

	if description, ok := d.GetOk("description"); ok {
		recordAttr.Description = description.(string)
	}

	if keywords, ok := d.GetOk("keywords"); ok {
		recordAttr.Keywords = keywords.(string)
	}

	if title, ok := d.GetOk("title"); ok {
		recordAttr.Title = title.(string)
	}

	if redirecttype, ok := d.GetOk("redirect_type"); ok {
		recordAttr.RedirectType = redirecttype.(string)
	}

	if hardlink, ok := d.GetOk("hardlink"); ok {
		recordAttr.HardLink = hardlink.(string)
	}

	if mxlevel, ok := d.GetOk("mx_level"); ok {
		recordAttr.MxLevel = mxlevel.(string)
	}

	if weight, ok := d.GetOk("weight"); ok {
		recordAttr.Weight = weight.(string)
	}

	if priority, ok := d.GetOk("priority"); ok {
		recordAttr.Priority = priority.(string)
	}

	if port, ok := d.GetOk("port"); ok {
		recordAttr.Port = port.(string)
	}

	if caatype, ok := d.GetOk("caa_type"); ok {
		recordAttr.CaaType = caatype.(string)
	}

	if issuer, ok := d.GetOk("issuer_critical"); ok {
		recordAttr.IssuerCritical = issuer.(string)
	}

	log.Println("Inside update method: recordattr: ", recordAttr)
	recordId := d.Id()

	recordAttr.IdUpdate = recordId
	_, err := dmeClient.Update(&recordAttr, "dns/managed/"+d.Get("domain_id").(string)+"/records/"+recordId)
	if err != nil {
		return err
	}

	return resourceManagedDNSRecordActionsRead(d, m)
}

func resourceManagedDNSRecordActionsDelete(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	dn := d.Id()

	err := dmeClient.Delete("dns/managed/" + d.Get("domain_id").(string) + "/records/" + dn)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
