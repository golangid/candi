package main

const (
	templateRepositorySQL = `package repository

import (
	"context"
	"database/sql"
	"fmt"

	"pkg.agungdwiprasetyo.com/candi/tracer"
)

// RepoSQL model
type RepoSQL struct {
	readDB, writeDB *sql.DB
	tx              *sql.Tx

	// add repository
}

// NewRepositorySQL constructor
func NewRepositorySQL(readDB, writeDB *sql.DB, tx *sql.Tx) *RepoSQL {

	return &RepoSQL{
		readDB: readDB, writeDB: writeDB, tx: tx,
	}
}

// WithTransaction run transaction for each repository with context, include handle canceled or timeout context
func (r *RepoSQL) WithTransaction(ctx context.Context, txFunc func(ctx context.Context, repo *RepoSQL) error) (err error) {
	trace := tracer.StartTrace(ctx, "RepoSQL-Transaction")
	defer trace.Finish()
	ctx = trace.Context()

	tx, errInit := r.writeDB.Begin()
	if errInit != nil {
		return errInit
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}

		if err != nil {
			tx.Rollback()
			trace.SetError(err)
		} else {
			tx.Commit()
		}
	}()

	// reinit new repository in different memory address with tx value
	manager := NewRepositorySQL(r.readDB, r.writeDB, tx)
	defer manager.free()

	errChan := make(chan error)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic: %v", r)
			}
			close(errChan)
		}()

		if err := txFunc(ctx, manager); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("Canceled or timeout: %v", ctx.Err())
	case e := <-errChan:
		return e
	}
}

func (r *RepoSQL) free() {
	// make nil all repository
}	
`
)
