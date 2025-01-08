package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type AccountContainer struct {
	Accounts []Account `json:"records" jsonout:"accounts"`
}
type Account struct {
	Id       string           `json:"Id" jsonout:"primaryKey"`
	Name     string           `json:"Name" jsonout:"name"`
	Contacts ContactContainer `json:"Contacts" jsonout:"contactsContainer"`
}
type ContactContainer struct {
	Contacts []Contact `json:"records" jsonout:"contacts"`
}
type Contact struct {
	ID          string `json:"Id" jsonout:"id"`
	FirstName   string `json:"FirstName" jsonout:"firstName"`
	LastName    string `json:"LastName" jsonout:"lastName"`
	Email       string `json:"Email" jsonout:"email"`
	CreatedDate string `json:"CreatedDate" jsonout:"createdDate" iso8601_utc:"true"`
	CreatedBy   User   `json:"CreatedBy" flatten:"External_Id" jsonout:"externalId"`
}
type User struct {
	External_Id string `json:"External_Id"`
}

func resolveFieldName(fieldType reflect.StructField) string {
	jsonOutTag := fieldType.Tag.Get("jsonout")
	jsonTag := fieldType.Tag.Get("json")
	if jsonOutTag != "" {
		return strings.Split(jsonOutTag, ",")[0]
	}
	if jsonTag != "" {
		return strings.Split(jsonTag, ",")[0]
	}
	return fieldType.Name
}

func MarshalWithJsonOut(input interface{}) ([]byte, error) {
	val := reflect.ValueOf(input)
	typ := reflect.TypeOf(input)

	outputMap := make(map[string]interface{})

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		jsonOutTag := resolveFieldName(fieldType)

		if !field.CanInterface() {
			continue
		}

		if flattenField := fieldType.Tag.Get("flatten"); flattenField != "" {
			nestedVal := reflect.ValueOf(field.Interface())
			nestedTyp := reflect.TypeOf(field.Interface())

			for j := 0; j < nestedTyp.NumField(); j++ {
				nestedField := nestedVal.Field(j)
				nestedFieldType := nestedTyp.Field(j)

				if nestedFieldType.Name == flattenField && nestedField.CanInterface() {
					outputMap[jsonOutTag] = nestedField.Interface()
					break
				}
			}
			continue
		}

		if iso8601Tag := fieldType.Tag.Get("iso8601_utc"); iso8601Tag == "true" {
			if str, ok := field.Interface().(string); ok && str != "" {
				const iso8601Layout = "2006-01-02T15:04:05.999-0700"
				parsedTime, err := time.Parse(iso8601Layout, str)
				if err != nil {
					fmt.Printf("Error parsing date for %s: %v\n", jsonOutTag, err)
					outputMap[jsonOutTag] = str
				} else {
					outputMap[jsonOutTag] = parsedTime.UTC().Format(time.RFC3339)
				}
			} else {
				outputMap[jsonOutTag] = nil
			}
			continue
		}

		if field.Kind() == reflect.Slice {
			sliceData := []interface{}{}
			for j := 0; j < field.Len(); j++ {
				nestedData, err := MarshalWithJsonOut(field.Index(j).Interface())
				if err != nil {
					fmt.Printf("Error marshaling slice element %s: %v\n", jsonOutTag, err)
					continue
				}
				var nestedMap map[string]interface{}
				if err := json.Unmarshal(nestedData, &nestedMap); err != nil {
					fmt.Printf("Error unmarshalling slice element data for %s: %v\n", jsonOutTag, err)
					continue
				}
				sliceData = append(sliceData, nestedMap)
			}
			outputMap[jsonOutTag] = sliceData
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

	return json.MarshalIndent(outputMap, "", "    ")
}

func main() {
	jsonInput := `{
        "totalSize": 1,
        "done": true,
        "records": [
            {
                "attributes": {
                    "type": "Account",
                    "url": "/services/data/v60.0/sobjects/Account/0012x00001ABC123"
                },
                "Id": "0012x00001ABC123",
                "Name": "TechCorp Inc.",
                "Contacts": {
                    "totalSize": 2,
                    "done": true,
                    "records": [
                        {
                            "attributes": {
                                "type": "Contact",
                                "url": "/services/data/v60.0/sobjects/Contact/0032x00001XYZ123"
                            },
                            "Id": "0032x00001XYZ123",
                            "FirstName": "John",
                            "LastName": "Doe",
                            "Email": "john.doe@techcorp.com",
                            "CreatedDate": "2025-01-06T16:15:39.452+0100",
                            "CreatedBy": {
                              "External_id": "Rix"
                            }
                        },
                        {
                            "attributes": {
                                "type": "Contact",
                                "url": "/services/data/v60.0/sobjects/Contact/0032x00001XYZ124"
                            },
                            "Id": "0032x00001XYZ124",
                            "FirstName": "Mary",
                            "LastName": "Doe",
                            "Email": "mary.doe@techcorp.com",
                            "CreatedDate": "2025-01-06T12:11:49.452+0100",
                            "CreatedBy": {
                              "External_id": "Deb"
                            }
                        }
                    ]
                }
            }
        ]
    }`

	var someObject AccountContainer
	err := json.Unmarshal([]byte(jsonInput), &someObject)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	jsonData, err := MarshalWithJsonOut(someObject)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	fmt.Println(string(jsonData))
}

// Version 2.7
