package dme

import (
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func StripQuotes(word string) string {
	if strings.HasPrefix(word, "\"") && strings.HasSuffix(word, "\"") {
		return strings.TrimPrefix(strings.TrimSuffix(word, "\""), "\"")
	} else if word == "{}" {
		word = ""
		return word
	}
	return word
}

func toListOfString(configured interface{}) []string {
	vs := make([]string, 0, 1)
	log.Println(configured.([]interface{}))
	for _, value := range configured.([]interface{}) {
		vs = append(vs, value.(string))
	}
	return vs
}

func toListOfInterface(name interface{}) []interface{} {
	nameList := make([]interface{}, 0)
	nameList = append(nameList, name)
	return nameList
}

// setIntField parses s as an integer and sets it on d. Silently no-ops on
// empty or unparseable values, matching how d.Set behaves for zero values.
func setIntField(d *schema.ResourceData, key string, s string) {
	if s == "" {
		return
	}
	if v, err := strconv.Atoi(s); err == nil {
		d.Set(key, v)
	}
}
