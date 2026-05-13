package dme

import (
	"fmt"

	"github.com/DNSMadeEasy/dme-go-client/container"
)

// findRecordByID locates a record in a DME records-list response by its
// stringified numeric id. Returns nil if the data array is missing,
// malformed, or contains no record with the requested id. Callers must
// treat nil as "record no longer exists" and clear state with SetId("").
func findRecordByID(con *container.Container, recordID string) *container.Container {
	if con == nil {
		return nil
	}
	raw := con.S("data").Data()
	data, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	for i, info := range data {
		val, ok := info.(map[string]interface{})
		if !ok {
			continue
		}
		if idVal, present := val["id"]; present && recordIDMatches(idVal, recordID) {
			return con.S("data").Index(i)
		}
	}
	return nil
}

// recordIDMatches compares a DME-returned id (decoded from JSON as
// float64) against a stringified record ID.
func recordIDMatches(idVal interface{}, recordID string) bool {
	switch v := idVal.(type) {
	case float64:
		return fmt.Sprintf("%d", int64(v)) == recordID
	case int:
		return fmt.Sprintf("%d", v) == recordID
	case int64:
		return fmt.Sprintf("%d", v) == recordID
	case string:
		return v == recordID
	}
	return false
}
