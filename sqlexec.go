package sqlexec

import (
	"database/sql"
	"errors"
)

// Executor is a client to execute SQLs.
type Executor struct {
	DB      *sql.DB
	Sources []SqlSource
}

// New creates a new Executor.
func New(db *sql.DB, sources ...SqlSource) *Executor {
	return &Executor{
		DB:      db,
		Sources: sources,
	}
}

func (e *Executor) Exec() error {
	if len(e.Sources) == 0 {
		return errors.New("no sources to be executed")
	}

	// execute SqlSource
	for _, s := range e.Sources {
		if err := e.execute(s); err != nil {
			return err
		}
	}

	return nil
}

func (e *Executor) execute(source SqlSource) error {
	tx, err := e.DB.Begin()
	if err != nil {
		return err
	}

	stmts, err := source(tx)
	if err != nil {
		return err
	}

	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Exec is a shortcut function that runs Executor.
func Exec(db *sql.DB, sources ...SqlSource) error {
	return New(db, sources...).Exec()
}
