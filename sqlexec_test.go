package sqlexec

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExec(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Error(err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(`CREATE TABLE post .*`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err = Exec(db, SourceString(`
CREATE TABLE post (
  id int NOT NULL,
  title text,
  body text,
  PRIMARY KEY(id)
);
`))
	assert.Nil(t, err)
}
