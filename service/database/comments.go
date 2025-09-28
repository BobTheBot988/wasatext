package database

import "database/sql"

func (db *appdbimpl) GetComments(convId int64, messageId int64) (*sql.Rows, error) {
	query := "SELECT commentId,content,msgId,convId,userId,userName FROM Comment WHERE msgId = $1 AND convId = $2"
	rows, err := db.c.Query(query, messageId, convId)
	if err != nil {
		return nil, err
	}
	return rows, err
}
