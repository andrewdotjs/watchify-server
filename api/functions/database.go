package functions

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

func DBCreate(database *sql.DB, tableName string, infoStruct *interface{}) *error {
	var objectValue reflect.Value = reflect.ValueOf(&infoStruct)
	var typeOfObject reflect.Type = objectValue.Type()
	var query string
	var columns []string
	var replacementString []string
	var values []interface{}

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

	// TODO: Raise error if length of columns = 0

	for i := 0; i < len(columns); i++ {
		replacementString = append(replacementString, "?")
	}

	query = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(columns, ", "), strings.Join(replacementString, ", "))
	if _, err := database.Exec(query, values...); err != nil {
		return &err
	}

	return nil
}

func DBReadSingle(database *sql.DB, tableName string, column string, value interface{}) *error {
	return nil
}

func DBReadMultiple(database *sql.DB, tableName string, columns []string, values []interface{}) *error {
	return nil
}

func DBUpdate(database *sql.DB, tableName string, infoStruct *interface{}) *error {
	var objectValue reflect.Value = reflect.ValueOf(&infoStruct)
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

	// TODO: Raise error if length of columns = 0

	values = append(values, objectId)
	query = fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", tableName, strings.Join(columns, ", "))

	if _, err := database.Exec(query, values...); err != nil {
		return &err
	}

	return nil
}

func DBDeleteSingle(database *sql.DB, tableName string, id string) *error {
	return nil
}

func DBDeleteMultiple(database *sql.DB, tableName string) *error {
	return nil
}
