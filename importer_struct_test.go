package sqlexec

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type RawSql string

func (r RawSql) SqlExpr() string {
	return string(r)
}

func TestResolveValue(t *testing.T) {
	assert.Equal(t, "'test'", resolveValue(reflect.ValueOf("test")))
	assert.Equal(t, "1", resolveValue(reflect.ValueOf(1)))
	assert.Equal(t, "1", resolveValue(reflect.ValueOf(int(1))))
	assert.Equal(t, "1", resolveValue(reflect.ValueOf(int8(1))))
	assert.Equal(t, "1", resolveValue(reflect.ValueOf(int16(1))))
	assert.Equal(t, "1", resolveValue(reflect.ValueOf(int32(1))))
	assert.Equal(t, "1", resolveValue(reflect.ValueOf(int64(1))))
	assert.Equal(t, "UNIX_TIMESTAMP()", resolveValue(reflect.ValueOf(RawSql("UNIX_TIMESTAMP()"))))
	assert.Equal(t, "NULL", resolveValue(reflect.ValueOf(sql.NullString{Valid: false})))
	assert.Equal(t, "NULL", resolveValue(reflect.ValueOf(nil)))
	assert.Equal(t, "NULL", resolveValue(reflect.ValueOf((*string)(nil))))
}

type TestUserImporter struct {
	Id           int64
	Name         string
	Age          int
	notUsedField string
}

func TestToSQL(t *testing.T) {
	user := TestUserImporter{Id: 1, Name: "test", Age: 20, notUsedField: "notUsed"}
	sql := structToSQL(user)
	assert.Equal(t, "INSERT INTO test_user (id, name, age) VALUES (1, 'test', 20);", sql)
}

func TestSourceImporter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO test_user \(id, name, age\) VALUES \(1, 'test1', 10\);`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO test_user \(id, name, age\) VALUES \(2, 'test2', 20\);`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = Exec(db, SourceStructImporter(func() []interface{} {
		return []interface{}{
			TestUserImporter{Id: 1, Name: "test1", Age: 10},
			TestUserImporter{Id: 2, Name: "test2", Age: 20},
		}
	}))
	assert.Nil(t, err)
}
