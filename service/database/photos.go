package database

import (
	"database/sql"
	"errors"

	"gitlab.com/mycompany8201046/myProject/service/api/model"
)

/* "Ah, my dear Tarnished, contemplate this:
In the grand tapestry of our world,
doth an image convey but a single message,
or doth it whisper myriad secrets untold? Reflect upon its depths,
as thou wandereth betwixt the stars and shadows." 🌌📸
*/

func (db *appdbimpl) GetPhoto(photoId int64) (val string, err error) {
	var path string

	tx, err := db.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				err = nil
			}
		}
	}()

	q := "SELECT path FROM Photo WHERE id = $1"
	err = tx.QueryRow(q, photoId).Scan(&path)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return "", err
	}

	if err = tx.Commit(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return "", err
	}
	return path, err
}

func (db *appdbimpl) InsertPhoto(pic model.Picture) (val int64, err error) {
	tx, err := db.BeginTx()
	if err != nil {
		return -1, err
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				err = nil
			}
		}
	}()

	q := `INSERT INTO Photo (path,size) VALUES($1,$2)`

	res, err := tx.Exec(q, "./"+pic.Path, pic.Size)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return -1, err
	}

	photoId, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	// Commit with hooks
	if err = tx.Commit(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return -1, err
	}

	return photoId, nil
}
