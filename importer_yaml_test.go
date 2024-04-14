package sqlexec

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestYamlToSQLs(t *testing.T) {
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
}
