# go-sqlexec

[![test](https://github.com/kohkimakimoto/go-sqlexec/actions/workflows/test.yml/badge.svg)](https://github.com/kohkimakimoto/go-sqlexec/actions/workflows/test.yml)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/kohkimakimoto/go-sqlexec/blob/master/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/kohkimakimoto/go-sqlexec.svg)](https://pkg.go.dev/github.com/kohkimakimoto/go-sqlexec)

go-sqlexec is a library for executing SQL queries from various sources such as files, directories and strings.

It is a straightforward library that features an `Exec` function. You can use this function with a database instance and `SqlSource` options to run the SQL queries you need. That's it!

`SqlSource` is a simple and flexible SQL data source mechanism. go-sqlexec includes several built-in `SqlSource` implementations. See [Built-in SqlSources](https://github.com/kohkimakimoto/go-sqlexec#built-in-sqlsources). You can also write your own `SqlSource` to suit your specific need.

Please see the usage below. It is a complete example.

## Usage

```golang
package sqlexec_test

import (
	"database/sql"
	"fmt"
	"github.com/kohkimakimoto/go-sqlexec"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"path/filepath"
)

func ExampleExec() {
	db, err := sql.Open("sqlite3", "./example.db")
	if err != nil {
		log.Fatal(err)
	}

	// Example of SourceDir
	// Use SQLs that are from *.sql files under the testdata/schema directory.
	schema := sqlexec.SourceDir(filepath.Join("testdata", "schema"))

	// Example of SourceString
	// Use SQLs that are written as strings.
	data := sqlexec.SourceString(
		"DELETE FROM post",
		"INSERT INTO post values (1, 'aaa', 'HelloWorld!')",
		"INSERT INTO post values (2, 'bbb', 'foobar')",
	)

	// Example of SourceFile
	// Use SQLs that are from a specific file
	data2 := sqlexec.SourceFile(filepath.Join("testdata", "data.sql"))

	// Example of custom SqlSource
	// You can use your custom SqlSource
	custom := CustomSqlSource()

	// Execute a set of SQL sources.
	// Each SqlSource is executed in an isolated transaction.
	if err := sqlexec.Exec(db, schema, data, data2, custom); err != nil {
		log.Fatal(err)
	}

	// Output:
	//1 aaa HelloWorld!
	//2 bbb foobar
	//3 ccc abcdefg
	//4 ddd abcdefg
}

// CustomSqlSource is an example of Custom SqlSource.
func CustomSqlSource() sqlexec.SqlSource {
	return func(tx *sql.Tx) ([]string, error) {
		// You can write arbitrary logic in the database transaction.
		rows, err := tx.Query("SELECT * FROM post ORDER BY id ASC")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var title string
			var body string
			if err := rows.Scan(&id, &title, &body); err != nil {
				return nil, err
			}

			// output current data
			fmt.Printf("%d %s %s\n", id, title, body)
		}

		// return SQLs
		return []string{
			"UPDATE post SET title = 'updated' where id = 1",
			"UPDATE post SET title = 'updated' where id = 2",
		}, nil
	}
}
```

## Built-in SqlSources

- `SourceString`: The SQLs are provided directly as strings.
- `SourceFile`: The SQLs are provided from file(s).
- `SourceDir`: The SQLs are provided from '*.sql' file(s) under the specific directory. The SQL files are executed in lexical order.
- `SourceImporters`: The SQLs are provided from structs.

## Author

Kohki Makimoto <kohki.makimoto@gmail.com>

## License

The MIT License (MIT)
