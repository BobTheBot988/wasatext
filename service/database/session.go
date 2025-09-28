package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) DoLogin(usrName string) (res sql.Result, err error) {

	tx, err := db.BeginTx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				err = nil
			}
		}
	}()

	// First check if user exists
	var userId int64
	checkQuery := "SELECT userId FROM User WHERE userName = $1"
	err = tx.QueryRow(checkQuery, usrName).Scan(&userId)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}
	return nil, tx.Commit()
}
