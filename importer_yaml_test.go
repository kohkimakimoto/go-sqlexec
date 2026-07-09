package sqlexec

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestYamlToSQLs(t *testing.T) {
	t.Run("general", func(t *testing.T) {
		stmts, err := yamlToSQLs([]byte(`
employees:
  - employee_id: 1
    name: "田中一郎"
    age: 34
    department_id: 11

  - employee_id: 2
    name: "佐藤恵子"
    age: 28
    department_id: 1

departments:
  - department_id: 1
    name: "営業部"

  - department_id: 2
    name: "技術部"
`))
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"INSERT INTO employees (employee_id, name, age, department_id) VALUES (1, '田中一郎', 34, 11), (2, '佐藤恵子', 28, 1);",
			"INSERT INTO departments (department_id, name) VALUES (1, '営業部'), (2, '技術部');",
		}, stmts)
	})

	t.Run("empty", func(t *testing.T) {
		stmts, err := yamlToSQLs([]byte(""))
		assert.NoError(t, err)
		assert.Equal(t, []string{}, stmts)
	})

	t.Run("null values", func(t *testing.T) {
		stmts, err := yamlToSQLs([]byte(`
employees:
  - employee_id: 1
    name: null
    age: 34
    department_id: 11
`))
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"INSERT INTO employees (employee_id, name, age, department_id) VALUES (1, null, 34, 11);",
		}, stmts)
	})
}

func TestYamlToSQLsWithOptions(t *testing.T) {
	t.Run("backtick identifiers", func(t *testing.T) {
		stmts, err := yamlToSQLsWithOptions([]byte(`
select:
  - from: "users"
    group: "admin"
`), SourceYamlImporterOptions{IdentifierQuote: IdentifierQuoteBacktick})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"INSERT INTO `select` (`from`, `group`) VALUES ('users', 'admin');",
		}, stmts)
	})

	t.Run("double quote identifiers", func(t *testing.T) {
		stmts, err := yamlToSQLsWithOptions([]byte(`
select:
  - from: "users"
    group: "admin"
`), SourceYamlImporterOptions{IdentifierQuote: IdentifierQuoteDouble})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			`INSERT INTO "select" ("from", "group") VALUES ('users', 'admin');`,
		}, stmts)
	})

	t.Run("escaped identifier quote", func(t *testing.T) {
		stmts, err := yamlToSQLsWithOptions([]byte("\"odd`table\":\n  - \"odd`column\": \"value\"\n"), SourceYamlImporterOptions{IdentifierQuote: IdentifierQuoteBacktick})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"INSERT INTO `odd``table` (`odd``column`) VALUES ('value');",
		}, stmts)
	})
}

func TestSourceYamlImporter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO employees \(employee_id, name, age, department_id\) VALUES \(1, '田中一郎', 34, 11\), \(2, '佐藤恵子', 28, 1\);`).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(`INSERT INTO departments \(department_id, name\) VALUES \(1, '営業部'\), \(2, '技術部'\);`).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()
	err = Exec(db, SourceYamlImporter(filepath.Join("testdata", "data.yml")))
	assert.Nil(t, err)
}

func TestSourceYamlStringImporter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO employees \(employee_id, name, age, department_id\) VALUES \(1, '田中一郎', 34, 11\), \(2, '佐藤恵子', 28, 1\);`).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(`INSERT INTO departments \(department_id, name\) VALUES \(1, '営業部'\), \(2, '技術部'\);`).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()
	err = Exec(db, SourceYamlStringImporter(`
employees:
  - employee_id: 1
    name: "田中一郎"
    age: 34
    department_id: 11

  - employee_id: 2
    name: "佐藤恵子"
    age: 28
    department_id: 1

departments:
  - department_id: 1
    name: "営業部"

  - department_id: 2
    name: "技術部"
`))
	assert.Nil(t, err)

}

func TestSourceYamlImporterWithOptions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `employees` \\(`employee_id`, `name`, `age`, `department_id`\\) VALUES \\(1, '田中一郎', 34, 11\\), \\(2, '佐藤恵子', 28, 1\\);").WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("INSERT INTO `departments` \\(`department_id`, `name`\\) VALUES \\(1, '営業部'\\), \\(2, '技術部'\\);").WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()
	err = Exec(db, SourceYamlImporterWithOptions(
		SourceYamlImporterOptions{IdentifierQuote: IdentifierQuoteBacktick},
		filepath.Join("testdata", "data.yml"),
	))
	assert.Nil(t, err)
}

func TestSourceYamlStringImporterWithOptions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `select` \\(`from`, `group`\\) VALUES \\('users', 'admin'\\);").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = Exec(db, SourceYamlStringImporterWithOptions(
		SourceYamlImporterOptions{IdentifierQuote: IdentifierQuoteBacktick},
		`
select:
  - from: "users"
    group: "admin"
`,
	))
	assert.Nil(t, err)
}
