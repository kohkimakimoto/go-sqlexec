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
	costom := CustomSqlSource()

	// Execute a set of SQL sources.
	// Each SqlSource is executed in an isolated transaction.
	if err := sqlexec.Exec(db, schema, data, data2, costom); err != nil {
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
