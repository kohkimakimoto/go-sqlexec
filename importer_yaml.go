package sqlexec

import (
	"database/sql"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

const (
	// IdentifierQuoteNone disables identifier quoting.
	IdentifierQuoteNone = ""
	// IdentifierQuoteDouble uses double quotes for identifiers.
	IdentifierQuoteDouble = `"`
	// IdentifierQuoteBacktick uses backticks for identifiers.
	IdentifierQuoteBacktick = "`"
)

// SourceYamlImporterOptions configures SourceYamlImporterWithOptions and SourceYamlStringImporterWithOptions.
type SourceYamlImporterOptions struct {
	// IdentifierQuote is the quote string used for table and column names.
	// Empty string keeps identifiers unquoted.
	IdentifierQuote string
}

func SourceYamlImporter(filenames ...string) SqlSource {
	return SourceYamlImporterWithOptions(SourceYamlImporterOptions{}, filenames...)
}

func SourceYamlImporterWithOptions(options SourceYamlImporterOptions, filenames ...string) SqlSource {
	return func(tx *sql.Tx) ([]string, error) {
		var retStmts []string
		for _, filename := range filenames {
			b, err := os.ReadFile(filename)
			if err != nil {
				return nil, err
			}
			stmts, err := yamlToSQLsWithOptions(b, options)
			if err != nil {
				return nil, err
			}
			if len(stmts) > 0 {
				retStmts = append(retStmts, stmts...)
			}
		}
		return retStmts, nil
	}
}

func SourceYamlStringImporter(yamlStrings ...string) SqlSource {
	return SourceYamlStringImporterWithOptions(SourceYamlImporterOptions{}, yamlStrings...)
}

func SourceYamlStringImporterWithOptions(options SourceYamlImporterOptions, yamlStrings ...string) SqlSource {
	return func(tx *sql.Tx) ([]string, error) {
		var retStmts []string
		for _, yamlString := range yamlStrings {
			stmts, err := yamlToSQLsWithOptions([]byte(yamlString), options)
			if err != nil {
				return nil, err
			}
			if len(stmts) > 0 {
				retStmts = append(retStmts, stmts...)
			}
		}
		return retStmts, nil
	}
}

func yamlToSQLs(data []byte) ([]string, error) {
	return yamlToSQLsWithOptions(data, SourceYamlImporterOptions{})
}

func yamlToSQLsWithOptions(data []byte, options SourceYamlImporterOptions) ([]string, error) {
	if len(data) == 0 {
		return []string{}, nil
	}
	var parsedData yaml.Node
	if err := yaml.Unmarshal(data, &parsedData); err != nil {
		return nil, err
	}

	var stmts []string
	for i := 0; i < len(parsedData.Content[0].Content); i += 2 {
		table := quoteSQLIdentifier(parsedData.Content[0].Content[i].Value, options.IdentifierQuote)
		records := parsedData.Content[0].Content[i+1]
		columnNames := []string{}
		values := []string{}

		for _, record := range records.Content {
			rowValues := []string{}
			for j := 0; j < len(record.Content); j += 2 {
				col := quoteSQLIdentifier(record.Content[j].Value, options.IdentifierQuote)
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

func quoteSQLIdentifier(identifier, quote string) string {
	if quote == "" {
		return identifier
	}
	return fmt.Sprintf("%s%s%s", quote, strings.ReplaceAll(identifier, quote, quote+quote), quote)
}
