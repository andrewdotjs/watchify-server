package functions

import (
	"fmt"
	"reflect"
	"strings"
)

func BuildInsertQuery(tableName string, object interface{}) (string, []any) {
	var objectValue reflect.Value = reflect.ValueOf(object)
	var typeOfObject reflect.Type = objectValue.Type()
	var query string
	var columns []string
	var replacementString []string
	var values []any

	// Iterate through object fields and values
	for i := 0; i < objectValue.NumField(); i++ {
		var fieldValue = objectValue.Field(i).Interface()
		var fieldName = strings.Replace(typeOfObject.Field(i).Tag.Get("json"), ",omitempty", "", 1)

		// Retrieve modified values
		if fieldValue != "" && fieldName != "cover" {
			columns = append(columns, fieldName)
			values = append(values, fieldValue)
		}
	}

	for i := 0; i < len(columns); i++ {
		replacementString = append(replacementString, "?")
	}

	query = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(columns, ", "), strings.Join(replacementString, ", "))

	return query, values
}

func BuildUpdateQuery(tableName string, object interface{}) (string, []any) {
	var objectValue reflect.Value = reflect.ValueOf(object)
	var typeOfObject reflect.Type = objectValue.Type()
	var objectId string
	var query string
	var columns []string
	var values []any

	// Iterate through object fields and values
	for i := 0; i < objectValue.NumField(); i++ {
		var fieldValue = objectValue.Field(i).Interface()
		var fieldName = strings.Replace(typeOfObject.Field(i).Tag.Get("json"), ",omitempty", "", 1)

		if fieldName == "id" {
			objectId = fmt.Sprint(fieldValue)
		}

		// Retrieve modified values
		if fieldValue != "" && fieldValue != false && fieldName != "id" && fieldName != "cover" {
			columns = append(columns, fmt.Sprintf("%s = ?", fieldName))
			values = append(values, fieldValue)
		}
	}

	values = append(values, objectId)
	query = fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", tableName, strings.Join(columns, ", "))

	return query, values
}
