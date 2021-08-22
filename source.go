package sqlexec

import (
	"bufio"
	"bytes"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SqlSource is a type that provides SQL to the Executor.
// It is a function that returns SQL strings. Each SqlSource will be executed in an isolated transaction.
type SqlSource func(tx *sql.Tx) ([]string, error)

// SourceString is the simplest SqlSource.
// It returns SQL strings that passed as statements.
func SourceString(statements ...string) SqlSource {
	return func(tx *sql.Tx) ([]string, error) {
		return statements, nil
	}
}

// SourceFile is a SqlSource that loads SQLs from one or multiple files.
func SourceFile(filenames ...string) SqlSource {
	return func(tx *sql.Tx) ([]string, error) {
		var retStmts []string
		for _, filename := range filenames {
			f, err := os.Open(filename)
			if err != nil {
				return nil, err
			}
			// parse source file
			stmts, err := ParseSQLStatements(f)
			if err != nil {
				return nil, err
			}

			retStmts = append(retStmts, stmts...)
		}
		return retStmts, nil
	}
}

// SourceDir is a SqlSource that loads SQls from the files under a directory.
func SourceDir(dirname string) SqlSource {
	return func(tx *sql.Tx) ([]string, error) {
		var retStmts []string
		if err := filepath.Walk(dirname, func(name string, info os.FileInfo, err error) error {
			// skip directory
			if info.IsDir() {
				return nil
			}

			// collect .sql file only
			base := filepath.Base(name)
			if ext := filepath.Ext(base); ext != ".sql" {
				return nil
			}

			f, err := os.Open(name)
			if err != nil {
				return err
			}
			// parse SQL source file
			stmts, err := ParseSQLStatements(f)
			if err != nil {
				return err
			}

			retStmts = append(retStmts, stmts...)
			return nil
		}); err != nil {
			return nil, err
		}

		return retStmts, nil
	}
}

func ParseSQLStatements(r io.Reader) ([]string, error) {
	var stmts []string
	scanner := bufio.NewScanner(r)

	var buf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()

		// ignore comment except beginning with '-- '
		if strings.HasPrefix(line, "-- ") {
			continue
		}

		if _, err := buf.WriteString(line + "\n"); err != nil {
			return nil, err
		}

		if endsWithSemicolon(line) {
			stmts = append(stmts, buf.String())
			buf.Reset()
		}
	}

	return stmts, nil
}

// Checks the line to see if the line has a statement-ending semicolon
// or if the line contains a double-dash comment.
func endsWithSemicolon(line string) bool {
	prev := ""
	scanner := bufio.NewScanner(strings.NewReader(line))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()
		if strings.HasPrefix(word, "--") {
			break
		}
		prev = word
	}

	return strings.HasSuffix(prev, ";")
}
