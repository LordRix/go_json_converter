package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type Person struct {
	FirstName  string    `json:"first_name" jsonout:"firstName"`
	LastName   string    `json:"last_name" jsonout:"lastName"`
	Age        int       `json:"age" jsonout:"ageYears"`
	Address    Address   `json:"address" jsonout:"homeAddress"`
	CreateDate time.Time `json:"create_date" jsonout:"createDate" iso8601_utc:"true"`
	CreateBy   User      `json:"create_by" jsonout:"createBy" flatten:"external_id"`
}

type Address struct {
	City     string `json:"city" jsonout:"city"`
	ZipCode  string `json:"zip_code" jsonout:"zipCode"`
	CreateBy User   `json:"create_by" jsonout:"createBy" flatten:"external_id"`
}

type User struct {
	ExternalID string `json:"external_id" jsonout:"externalId"`
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

		if iso8601Tag := fieldType.Tag.Get("iso8601_utc"); iso8601Tag == "true" {
			if t, ok := field.Interface().(time.Time); ok && !t.IsZero() {
				outputMap[jsonOutTag] = t.Format(time.RFC3339)
			} else {
				outputMap[jsonOutTag] = nil
			}
			continue
		}

		if flattenField := fieldType.Tag.Get("flatten"); flattenField != "" {
			nestedVal := reflect.ValueOf(field.Interface())
			nestedTyp := reflect.TypeOf(field.Interface())

			for j := 0; j < nestedTyp.NumField(); j++ {
				nestedField := nestedVal.Field(j)
				nestedFieldType := nestedTyp.Field(j)

				if nestedFieldType.Tag.Get("json") == flattenField && nestedField.CanInterface() {
					if nestedField.Kind() == reflect.Struct {
						nestedData, err := MarshalWithJsonOut(nestedField.Interface())
						if err != nil {
							fmt.Printf("Error flattening nested field %s: %v\n", flattenField, err)
							continue
						}
						var nestedMap map[string]interface{}
						if err := json.Unmarshal(nestedData, &nestedMap); err != nil {
							fmt.Printf("Error unmarshalling flattened data for %s: %v\n", flattenField, err)
							continue
						}
						for k, v := range nestedMap {
							outputMap[k] = v
						}
					} else {
						outputMap[jsonOutTag] = nestedField.Interface()
					}
					break
				}
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

	return json.MarshalIndent(outputMap, "", "    ")
}

func main() {
	jsonInput := `{
        "first_name": "John",
        "last_name": "Doe",
        "age": 30,
        "address": {
            "city": "New York",
            "zip_code": "10001",
            "create_by": {
                "external_id": "Deb"
            }
        },
        "create_date": "2025-01-06T16:00:00Z",
        "create_by": {
            "external_id": "Rix"
        }
    }`

	var someObject Person
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

// Version 2.0
