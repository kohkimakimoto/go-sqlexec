package sqlexec

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestSourceString(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS post .*`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err = Exec(db, SourceString(`
CREATE TABLE IF NOT EXISTS post (
  id int NOT NULL,
  title text,
  body text,
  PRIMARY KEY(id)
);
`))
	assert.Nil(t, err)
}

func TestSourceDir(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS post .*`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS comment .*`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS user .*`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err = Exec(db, SourceDir(filepath.Join("testdata", "schema")))
	assert.Nil(t, err)
}

func TestSourceFile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS post .*`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS comment .*`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err = Exec(db, SourceFile(filepath.Join("testdata", "schema", "000_a.sql")))
	assert.Nil(t, err)
}
