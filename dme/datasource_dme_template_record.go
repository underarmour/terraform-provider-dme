package dme

import (
	"fmt"
	"log"
	"strconv"

	"github.com/DNSMadeEasy/dme-go-client/client"
	"github.com/DNSMadeEasy/dme-go-client/container"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceDMETemplateRecord() *schema.Resource {
	return &schema.Resource{
		Read: datasourceDMETemplateRecordRead,

		Schema: map[string]*schema.Schema{
			"template_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"value": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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
				Optional: true,
				Computed: true,
			},

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

func datasourceDMETemplateRecordRead(d *schema.ResourceData, m interface{}) error {
	dmeClient := m.(*client.Client)
	name := d.Get("name").(string)
	recordtype := d.Get("type").(string)

	con, err := dmeClient.GetbyId("dns/template/" + d.Get("template_id").(string) + "/records")
	if err != nil {
		return err
	}

	// data := con.S("data").Data().([]interface{})
	// var flag bool
	// var count int
	// for _, info := range data {
	// 	val := info.(map[string]interface{})
	// 	if StripQuotes(val["name"].(string)) == name && StripQuotes(val["type"].(string)) == recordtype {
	// 		flag = true
	// 		break
	// 	}
	// 	count = count + 1
	// }
	// if flag != true {
	// 	return fmt.Errorf("Record of specified name not found")
	// }

	pages, _ := strconv.Atoi(con.S("totalPages").String())
	log.Println("total pages value: ", pages)

	log.Println("name value from get: ", name)
	log.Println("type value from get: ", recordtype)
	var finalCount = 0
	cont1 := &container.Container{}
	for j := 1; j <= pages; j++ {
		pageValue := fmt.Sprintf("%v", j)
		log.Println("page value: ", pageValue)
		getCont, _ := dmeClient.GetbyId("dns/template/" + d.Get("template_id").(string) + "/records?page=" + pageValue)
		log.Println("value of getCont: ", getCont)
		count, _ := getCont.ArrayCount("data")
		log.Println("value of count container: ", count)

		for i := 0; i < count; i++ {
			tempCont, _ := getCont.ArrayElement(i, "data")
			nameCont := extractField(tempCont.S("name"))
			typeCont := extractField(tempCont.S("type"))
			log.Println("name value inside containeer", nameCont)
			log.Println("type value inside containeer", typeCont)
			if name == nameCont && recordtype == typeCont {
				log.Println("inside if:")
				cont1 = tempCont
				finalCount = 1
				break
			}
		}
		if finalCount == 1 {
			break
		}
	}

	if finalCount != 1 {
		return fmt.Errorf("Record with given name does not exist")
	}
	log.Println("finalContainer value: ", cont1)

	d.SetId(fmt.Sprintf("%v", cont1.S("id").String()))
	recordType := extractField(cont1.S("type"))
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
	d.Set("priority", extractField(cont1.S("priority")))
	d.Set("port", extractField(cont1.S("port")))
	d.Set("caa_type", extractField(cont1.S("caaType")))
	d.Set("issuer_critical", extractField(cont1.S("issuerCritical")))

	return nil

}
