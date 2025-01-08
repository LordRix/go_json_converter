package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Root struct representing the entire API response
type AccountResponse struct {
	TotalSize int       `json:"totalSize"`
	Done      bool      `json:"done"`
	Records   []Account `json:"records"`
}

// Account struct representing the Account object and its relationships
type Account struct {
	Attributes    Attributes      `json:"attributes"`
	Id            string          `json:"Id"`
	Name          string          `json:"Name"`
	Contacts      ContactList     `json:"Contacts"`
	Opportunities OpportunityList `json:"Opportunities"`
}

// ContactList to handle nested Contact objects
type ContactList struct {
	TotalSize int       `json:"totalSize"`
	Done      bool      `json:"done"`
	Records   []Contact `json:"records"`
}

// OpportunityList to handle nested Opportunity objects
type OpportunityList struct {
	TotalSize int           `json:"totalSize"`
	Done      bool          `json:"done"`
	Records   []Opportunity `json:"records"`
}

// Contact struct for individual contacts
type Contact struct {
	Attributes Attributes `json:"attributes"`
	Id         string     `json:"Id"`
	FirstName  string     `json:"FirstName"`
	LastName   string     `json:"LastName"`
	Email      string     `json:"Email"`
}

// Opportunity struct for individual opportunities
type Opportunity struct {
	Attributes Attributes `json:"attributes"`
	Id         string     `json:"Id"`
	Name       string     `json:"Name"`
	StageName  string     `json:"StageName"`
	CloseDate  string     `json:"CloseDate"`
}

// Attributes struct for metadata
type Attributes struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func parseTag(tag string, suffix string) (string, bool) {
	if strings.HasSuffix(tag, suffix) {
		return strings.TrimSuffix(tag, suffix), true
	}
	return tag, false
}

func MarshalWithJsonOut(input interface{}) ([]byte, error) {
	val := reflect.ValueOf(input)
	typ := reflect.TypeOf(input)

	outputMap := make(map[string]interface{})

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		jsonOutTag := fieldType.Tag.Get("jsonout")
		jsonTag := fieldType.Tag.Get("json")

		if jsonOutTag == "" {
			if jsonTag != "" {
				jsonOutTag = strings.Split(jsonTag, ",")[0]
			} else {
				jsonOutTag = fieldType.Name
			}
		}

		if tag, flatten := parseTag(jsonOutTag, ",flatten"); flatten {
			jsonOutTag = tag
			if field.Kind() == reflect.Struct {
				for j := 0; j < field.NumField(); j++ {
					nestedField := field.Field(j)
					nestedType := field.Type().Field(j)
					nestedJsonOutTag := nestedType.Tag.Get("jsonout")
					nestedJsonTag := nestedType.Tag.Get("json")
					if nestedJsonOutTag == "" {
						if nestedJsonTag != "" {
							nestedJsonOutTag = strings.Split(nestedJsonTag, ",")[0]
						} else {
							nestedJsonOutTag = nestedType.Name
						}
					}
					outputMap[nestedJsonOutTag] = nestedField.Interface()
				}
			}
			continue
		}

		if tag, isDate := parseTag(jsonOutTag, ",date"); isDate {
			jsonOutTag = tag
			if t, ok := field.Interface().(time.Time); ok {
				outputMap[jsonOutTag] = t.Format(time.RFC3339)
			}
			continue
		}

		if field.Kind() == reflect.Struct {
			nestedData, err := MarshalWithJsonOut(field.Interface())
			if err != nil {
				fmt.Printf("Error marshaling nested field %s: %v\n", jsonOutTag, err)
				continue
			}
			var nestedMap map[string]interface{}
			if err := json.Unmarshal(nestedData, &nestedMap); err != nil {
				fmt.Printf("Error unmarshalling nested data for %s: %v\n", jsonOutTag, err)
				continue
			}
			outputMap[jsonOutTag] = nestedMap
		} else {
			outputMap[jsonOutTag] = field.Interface()
		}
	}

	return json.Marshal(outputMap)
}

func main() {
	jsonInput := `{
        "totalSize": 1,
        "done": true,
        "records": [{
            "attributes": {"type": "Account", "url": "/services/data/v60.0/sobjects/Account/0012x00001ABC123"},
            "Id": "0012x00001ABC123",
            "Name": "TechCorp Inc.",
            "Contacts": {
                "totalSize": 1,
                "done": true,
                "records": [{
                    "attributes": {"type": "Contact", "url": "/services/data/v60.0/sobjects/Contact/0032x00001XYZ123"},
                    "Id": "0032x00001XYZ123",
                    "FirstName": "John",
                    "LastName": "Doe",
                    "Email": "john.doe@techcorp.com"
                }]
            },
            "Opportunities": {
                "totalSize": 1,
                "done": true,
                "records": [{
                    "attributes": {"type": "Opportunity", "url": "/services/data/v60.0/sobjects/Opportunity/0062x00001OPP123"},
                    "Id": "0062x00001OPP123",
                    "Name": "Enterprise Sale",
                    "StageName": "Closed Won",
                    "CloseDate": "2024-06-30"
                }]
            }
        }]
    }`

	var person AccountResponse
	err := json.Unmarshal([]byte(jsonInput), &person)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	jsonData, err := MarshalWithJsonOut(person)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	fmt.Println(string(jsonData))
}
