package sqlexec

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

type ImporterFunc func() []interface{}

// SourceImporter is a SqlSource that generates SQLs from the structs returned by the ImporterFunc.
func SourceImporter(importerFunc ImporterFunc) SqlSource {
	return func(tx *sql.Tx) ([]string, error) {
		importerStructs := importerFunc()
		if len(importerStructs) == 0 {
			return nil, nil
		}
		sqls := make([]string, 0, len(importerStructs))
		for _, im := range importerStructs {
			sqls = append(sqls, toSQL(im))
		}
		return sqls, nil
	}
}

func toSQL(obj interface{}) string {
	val := reflect.ValueOf(obj)
	typ := val.Type()

	name := typ.Name()
	name = strings.TrimSuffix(name, "Importer")
	tableName := toSnakeCase(name)

	fieldNames := []string{}
	values := []string{}

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath == "" { // Means the field is exported
			fieldName := toSnakeCase(field.Name)
			fieldNames = append(fieldNames, fieldName)
			valueString := resolveValue(val.Field(i))
			values = append(values, valueString)
		}
	}

	sqlStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);",
		tableName,
		strings.Join(fieldNames, ", "),
		strings.Join(values, ", "))

	return sqlStr
}

func resolveValue(fieldValue reflect.Value) string {
	// Check if the fieldValue is valid or a nil pointer
	if !fieldValue.IsValid() || (fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil()) {
		return "NULL"
	}

	// Check if the fieldValue is a pointer and handle nil
	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			return "NULL"
		}
		// Dereference the pointer
		fieldValue = fieldValue.Elem()
	}

	// Check if SqlExpr method exists
	method := fieldValue.MethodByName("SqlExpr")
	if method.IsValid() {
		results := method.Call(nil)
		if len(results) > 0 {
			// Expecting the method to return a single string
			if sqlExpr, ok := results[0].Interface().(string); ok {
				return sqlExpr
			}
		}
	}

	// Handle sql.NullString and similar types
	if fieldValue.Type().PkgPath() == "database/sql" {
		switch fieldValue.Interface().(type) {
		case sql.NullString:
			ns := fieldValue.Interface().(sql.NullString)
			if ns.Valid {
				return fmt.Sprintf("'%s'", escapeSQLString(ns.String))
			} else {
				return "NULL"
			}
		case sql.NullInt16:
			ni := fieldValue.Interface().(sql.NullInt16)
			if ni.Valid {
				return fmt.Sprintf("%d", ni.Int16)
			} else {
				return "NULL"
			}
		case sql.NullInt32:
			ni := fieldValue.Interface().(sql.NullInt32)
			if ni.Valid {
				return fmt.Sprintf("%d", ni.Int32)
			} else {
				return "NULL"
			}
		case sql.NullInt64:
			ni := fieldValue.Interface().(sql.NullInt64)
			if ni.Valid {
				return fmt.Sprintf("%d", ni.Int64)
			} else {
				return "NULL"
			}
		case sql.NullFloat64:
			nf := fieldValue.Interface().(sql.NullFloat64)
			if nf.Valid {
				return fmt.Sprintf("%f", nf.Float64)
			} else {
				return "NULL"
			}
		case sql.NullTime:
			nt := fieldValue.Interface().(sql.NullTime)
			if nt.Valid {
				return fmt.Sprintf("'%s'", nt.Time.Format("2006-01-02 15:04:05"))
			} else {
				return "NULL"
			}
		}
	}

	switch fieldValue.Kind() {
	case reflect.String:
		// Escape the string and wrap in single quotes
		return fmt.Sprintf("'%s'", escapeSQLString(fieldValue.String()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Directly use the integer value
		return fmt.Sprintf("%d", fieldValue.Int())
	case reflect.Float32, reflect.Float64:
		// Directly use the float value
		return fmt.Sprintf("%f", fieldValue.Float())
	default:
		// Default to using string formatting, wrapped in single quotes
		return fmt.Sprintf("'%v'", fieldValue)
	}
}

// escapeSQLString escapes single quotes in strings for SQL queries
func escapeSQLString(str string) string {

	return strings.ReplaceAll(str, "'", "''")
}

func toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
