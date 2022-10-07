package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

const (
	TAG_JSON_ARRAY = "db_json[]"
	TAG_JSON       = "db_json"
	TAG_DEFAULT    = "db"
)

type query struct {
	db *sqlx.DB
}

type Query interface {
	InsertValues(anyStructure any) (columns string, values string, inputValues []any)
	UpdateValues(anyStructure any) (condition string, values []any)
}

func NewQuery(db *sqlx.DB) Query {
	return &query{
		db: db,
	}
}

func getIdColumn(structure reflect.Value) (tag string, value any) {
	if structure.Type().Kind() == reflect.Pointer && structure.IsNil() {
		return
	}
	if structure.Type().Kind() == reflect.Pointer {
		structure = reflect.Indirect(structure)
	}
	for i := 0; i < structure.NumField(); i++ {
		if structure.Type().Field(i).Type.Kind() == reflect.Struct {
			continue
		}
		if (structure.Type().Field(i).Name == "id" || structure.Type().Field(i).Name == "ID") && structure.Type().Field(i).Tag.Get("entity") != "" {
			tag = structure.Type().Field(i).Tag.Get("entity")
			value = structure.Field(i).Interface()
			return
		}
	}

	return
}

func (q *query) InsertValues(anyStructure any) (columns string, values string, inputValues []any) {
	var valuesBuilder, columnsBuilder []string
	currentStructure := reflect.ValueOf(anyStructure)

	if reflect.TypeOf(anyStructure).Kind() == reflect.Pointer {
		currentStructure = reflect.Indirect(currentStructure)
	}

	var iterator = 1
	for i := 0; i < currentStructure.NumField(); i++ {
		if currentStructure.Type().Field(i).Name == "id" || currentStructure.Type().Field(i).Name == "ID" {
			continue
		}

		dbTag := currentStructure.Type().Field(i).Tag.Get("db")
		fieldValue := currentStructure.Field(i).Interface()

		if currentStructure.Field(i).Kind() == reflect.Pointer || currentStructure.Field(i).Kind() == reflect.Struct {
			tag, value := getIdColumn(currentStructure.Field(i))
			if tag != "" && value != "" {
				dbTag = tag
				fieldValue = value
			}
		}

		if dbTag == "end" || dbTag == "start" {
			dbTag = fmt.Sprintf("%q", dbTag)
		}

		if dbTag != "" && fieldValue != "" {
			inputValues = append(inputValues, fieldValue)
			columnsBuilder = append(columnsBuilder, dbTag)
			valuesBuilder = append(valuesBuilder, fmt.Sprintf("$%d", iterator))
			iterator++
		}
	}
	columns = strings.Join(columnsBuilder, ", ")
	values = strings.Join(valuesBuilder, ", ")

	return
}

func (q *query) UpdateValues(anyStructure any) (condition string, values []any) {
	var conditionBuilder []string
	currentStructure := reflect.ValueOf(anyStructure)

	if reflect.TypeOf(anyStructure).Kind() == reflect.Pointer {
		currentStructure = reflect.Indirect(currentStructure)
	}

	var iterator = 1
	for i := 0; i < currentStructure.NumField(); i++ {
		isStruct := currentStructure.Field(i).Type().Kind() == reflect.Struct
		isNilPtr := currentStructure.Field(i).Type().Kind() == reflect.Pointer && currentStructure.Field(i).IsNil()
		isId := currentStructure.Type().Field(i).Name == "id" || currentStructure.Type().Field(i).Name == "ID"

		if isStruct || isId || isNilPtr {
			continue
		}
		dbTag := currentStructure.Type().Field(i).Tag.Get("db")
		fieldValue := currentStructure.Field(i).Interface()
		if currentStructure.Field(i).Kind() == reflect.Pointer {
			fieldValue = reflect.Indirect(currentStructure.Field(i)).Interface()
		}
		if dbTag == "end" || dbTag == "start" {
			dbTag = fmt.Sprintf("%q", dbTag)
		}
		if dbTag != "" {
			if currentStructure.Field(i).Kind() == reflect.Struct ||
				currentStructure.Field(i).Kind() == reflect.Slice ||
				currentStructure.Field(i).Kind() == reflect.Map {

				if (currentStructure.Field(i).Kind() == reflect.Slice ||
					currentStructure.Field(i).Kind() == reflect.Array ||
					currentStructure.Field(i).Kind() == reflect.Map) &&
					currentStructure.Field(i).Len() == 0 {
					continue
				}
				bytes, err := json.Marshal(fieldValue)
				if err != nil {
					panic(err)
				}
				values = append(values, string(bytes))
			} else {
				values = append(values, fieldValue)
			}
			conditionBuilder = append(conditionBuilder, fmt.Sprintf("%s = $%d", dbTag, iterator))
			iterator++
		}
	}
	condition = strings.Join(conditionBuilder, ", ")

	return
}
