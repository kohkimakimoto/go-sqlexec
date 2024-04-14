package sqlexec

import (
	"database/sql"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

func SourceYamlImporter(filenames ...string) SqlSource {
	return func(tx *sql.Tx) ([]string, error) {
		var retStmts []string
		for _, filename := range filenames {
			b, err := os.ReadFile(filename)
			if err != nil {
				return nil, err
			}
			stmts, err := yamlToSQLs(b)
			if err != nil {
				return nil, err
			}
			retStmts = append(retStmts, stmts...)
		}
		return retStmts, nil
	}
}

func yamlToSQLs(data []byte) ([]string, error) {
	var parsedData yaml.Node
	if err := yaml.Unmarshal(data, &parsedData); err != nil {
		return nil, err
	}

	var stmts []string
	for i := 0; i < len(parsedData.Content[0].Content); i += 2 {
		table := parsedData.Content[0].Content[i].Value
		records := parsedData.Content[0].Content[i+1]
		columnNames := []string{}
		values := []string{}

		for _, record := range records.Content {
			rowValues := []string{}
			for j := 0; j < len(record.Content); j += 2 {
				col := record.Content[j].Value
				val := record.Content[j+1]

				if len(columnNames) < len(record.Content)/2 {
					columnNames = append(columnNames, col)
				}

				valStr := ""
				switch val.Kind {
				case yaml.ScalarNode:
					if val.Tag == "!!str" {
						valStr = fmt.Sprintf("'%s'", escapeSQLString(val.Value))
					} else {
						valStr = val.Value // Non-string scalar (int, float, etc.)
					}
				default:
					valStr = fmt.Sprintf("'%s'", val.Value) // Default case for unknown types
				}

				rowValues = append(rowValues, valStr)
			}
			values = append(values, fmt.Sprintf("(%s)", strings.Join(rowValues, ", ")))
		}

		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s;", table, strings.Join(columnNames, ", "), strings.Join(values, ", "))
		stmts = append(stmts, stmt)
	}
	return stmts, nil
}
