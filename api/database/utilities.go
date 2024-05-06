package database

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// Destination must be a pointer to a struct
func RowsToStructs(rows *sql.Rows, destination interface{}) error {
	destinationClone := reflect.ValueOf(destination).Elem()

	arguments := make([]interface{}, destinationClone.Type().Elem().NumField())

	for rows.Next() {
		rowp := reflect.New(destinationClone.Type().Elem())
		rowv := rowp.Elem()

		for i := 0; i < rowv.NumField(); i++ {
			arguments[i] = rowv.Field(i).Addr().Interface()
		}

		if err := rows.Scan(arguments...); err != nil {
			return err
		}

		destinationClone.Set(reflect.Append(destinationClone, rowv))
	}
	return nil
}

// Destination must be a pointer to a struct
func RowToStructs(row *sql.Row, destination interface{}) error {
	destinationClone := reflect.ValueOf(destination).Elem()

	arguments := make([]interface{}, destinationClone.Type().NumField())

	rowp := reflect.New(destinationClone.Type())
	rowv := rowp.Elem()

	for i := 0; i < rowv.NumField(); i++ {
		arguments[i] = rowv.Field(i).Addr().Interface()
	}

	if err := row.Scan(arguments...); err != nil {
		return err
	}

	destinationClone.Set(rowv)
	return nil
}

func UpdateQueryBuilder(tableName string, decodedRequestObject any) (string, []any) {
	var values []any
	var id string
	valueReflect := reflect.ValueOf(decodedRequestObject)
	typeOfS := valueReflect.Type()

	setString := ""
	count := 1

	for i := 0; i < valueReflect.NumField(); i++ {
		field := strings.Split(typeOfS.Field(i).Tag.Get("json"), ",")[0]
		value := valueReflect.Field(i).Interface()

		if !reflect.ValueOf(value).IsZero() || reflect.TypeOf(value).String() == "bool" {
			if count > 2 {
				setString += ", "
			}

			if field == "id" {
				id = fmt.Sprint(value)
			} else {
				setString += fmt.Sprintf("%v = ?", field)
				values = append(values, value)
			}
			count++
		}
	}

	values = append(values, id)

	statement := fmt.Sprintf(
		`
			UPDATE
				%v
			SET
				%v
			WHERE
				id = ?
		`,
		tableName,
		setString,
	)

	return statement, values
}

func InsertQueryBuilder(tableName string, decodedRequestObject any) (string, []any) {
	var values []any

	valueString := ""
	count := 1

	valueReflect := reflect.ValueOf(decodedRequestObject)
	typeOfS := valueReflect.Type()

	for i := 0; i < valueReflect.NumField(); i++ {
		field := strings.Split(typeOfS.Field(i).Tag.Get("json"), ",")[0]
		value := valueReflect.Field(i).Interface()

		if count > 1 {
			valueString += ", "
		}

		// Swap to go-idiomatic switch statement if this gets big and ugly
		if field == "password" {
			valueString += fmt.Sprintf("crypt($%v, gen_salt('bf'))", count)
		} else {
			valueString += fmt.Sprintf("$%v", count)
		}

		if !reflect.ValueOf(value).IsZero() || reflect.TypeOf(value).String() == "float32" {
			values = append(values, value)
		} else {
			values = append(values, nil)
		}

		count++
	}

	statement := fmt.Sprintf(
		`
			INSERT INTO
				%v
			VALUES (
				%v
			)
		`,
		tableName,
		valueString,
	)

	return statement, values
}
