package database

import (
	"database/sql"
	"errors"

	"gitlab.com/mycompany8201046/myProject/service/api/model"
)

func (db *appdbimpl) GetMessage(msgId int64, convId int64) *sql.Row {
	q := "SELECT messageId,content,mtime,usrSenderId,convId,IFNULL(photoId,-1),IFNULL(repliedId,-1),IFNULL(repliedConvId,-1) FROM Message WHERE messageId=$1 AND convId=$2"
	res := db.c.QueryRow(q, msgId, convId)
	return res
}

func (db *appdbimpl) CreateMessage(message model.MessageInput, photoId int64, usrId int64, convId int64) (val int64, err error) {
	tx, err := db.BeginTx()
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return 0, err
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				err = nil
			}
		}
	}()
	// Will be ignored if commit succeeds

	// * Needed to emulate AUTOINCREMENT Feature ON Message TABLE
	var next_id int64
	var query string
	next_id_q := "SELECT COALESCE(MAX(messageId) + 1, 1) FROM Message WHERE convId = $1"
	err = tx.QueryRow(next_id_q, convId).Scan(&next_id)

	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return next_id, err
	}

	if photoId > 0 {
		query = "INSERT INTO Message (messageId,content,mtime,usrSenderId,convId,photoId,repliedId,repliedConvId) VALUES($1,$2,unixepoch(),$3,$4,$5,$6,$7);"
		_, err = tx.Exec(query, next_id, message.Content, usrId, convId, photoId, message.RepliedId, message.RepliedConvId)
	} else {
		query = "INSERT INTO Message (messageId,content,mtime,usrSenderId,convId,repliedId,repliedConvId) VALUES($1,$2,unixepoch(),$3,$4,$5,$6);"
		_, err = tx.Exec(query, next_id, message.Content, usrId, convId, message.RepliedId, message.RepliedConvId)

	}
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return 0, err
	}

	// Commit with hooks
	if err = tx.Commit(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return 0, err
	}

	return next_id, nil
}

func (db *appdbimpl) ForwardMessage(ogMessageId int64, ogConvId int64, usrId int64, newConvId int64) (res sql.Result, err error) {
	tx, err := db.BeginTx()
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				err = nil
			}
		}
	}()

	// Here we take the contents of the original message and we copy them over to the new message
	var messageContent string
	var photoId int64
	q := "SELECT content,IFNULL(photoId,-1) FROM Message WHERE messageId = $1 and convId = $2"
	e := tx.QueryRow(q, ogMessageId, ogConvId).Scan(&messageContent, &photoId)
	if e != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, e
	}
	if newConvId < 0 {
		var userArray = []int64{usrId, -newConvId}

		newConvId, err = db.createConversationWithTx(tx, userArray)
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			return nil, err
		}
	}

	// We create the new message using transaction
	_, err = db.createMessageWithTx(tx, messageContent, photoId, usrId, newConvId, 0, 0)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return nil, nil
}

// Helper method to create message within existing transaction
func (db *appdbimpl) createMessageWithTx(tx *HookedTx, messageContent string, photoId int64, usrId int64, convId int64, replId int64, replConvId int64) (int64, error) {
	var next_id int64
	var query string
	next_id_q := "SELECT COALESCE(MAX(messageId) + 1, 1) FROM Message WHERE convId = $1"
	err := tx.QueryRow(next_id_q, convId).Scan(&next_id)

	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return next_id, err
	}

	if photoId > 0 {
		query = "INSERT INTO Message (messageId,content,mtime,usrSenderId,convId,photoId,repliedId,repliedConvId) VALUES($1,$2,unixepoch(),$3,$4,$5,$6,$7);"
		_, err = tx.Exec(query, next_id, messageContent, usrId, convId, photoId, replId, replConvId)

	} else {
		query = "INSERT INTO Message (messageId,content,mtime,usrSenderId,convId,repliedId,repliedConvId) VALUES($1,$2,unixepoch(),$3,$4,$5,$6);"
		_, err = tx.Exec(query, next_id, messageContent, usrId, convId, replId, replConvId)

	}
	if errors.Is(err, sql.ErrTxDone) {
		err = nil
	}
	return next_id, err
}

func (db *appdbimpl) HasMessageBeenRead(msgId int64, convId int64, userId int64) (bool, error) {
	var hasBeenread uint8
	query := `SELECT 
            (SELECT COUNT(DISTINCT "userId") FROM MessageReadStatus 
             WHERE messageId = $1 AND convId = $2) = 
            (SELECT COUNT(DISTINCT "usrId") FROM Conv_User 
             WHERE convId = $2 AND usrId NOT IN 
                (SELECT usrSenderId FROM Message 
                 WHERE messageId = $1 AND convId = $2))`
	err := db.c.QueryRow(query, msgId, convId).Scan(&hasBeenread)
	if err != nil {
		return false, err
	}
	return hasBeenread != 0, nil
}

func (db *appdbimpl) WhoHasReadMessage(msgId int64, convId int64, userId int64) (*sql.Rows, error) {
	query := `SELECT userId FROM MessageReadStatus WHERE messageId = $1 AND convId = $2`
	rows, err := db.c.Query(query, msgId, convId)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (db *appdbimpl) WhoHasNotReadMessage(msgId int64, convId int64, userId int64) (*sql.Rows, error) {
	query := `SELECT "usrId" FROM "Conv_User"
        WHERE "convId" = $1
        AND "usrId" NOT IN (
            -- Users who have read the message
            SELECT "userId" FROM "MessageReadStatus"
            WHERE "messageId" = $2 AND "convId" = $1
        )
        AND "usrId" != (
            -- The message sender
            SELECT "usrSenderId" FROM "Message"
            WHERE "messageId" = $2 AND "convId" = $1
        )`
	rows, err := db.c.Query(query, convId, msgId)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (db *appdbimpl) ReadMessage(msgId int64, convId int64, userId int64) error {
	tx, err := db.BeginTx()
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return err
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				err = nil
			}
		}
	}()

	query := "INSERT OR IGNORE INTO MessageReadStatus (messageId,convId,userId,readTime) VALUES($1,$2,$3,unixepoch())"
	_, err = tx.Exec(query, msgId, convId, userId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return err
	}

	return tx.Commit()
}

func (db *appdbimpl) DeleteMessage(msgId int64, convId int64) (res sql.Result, err error) {
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

	query := "DELETE FROM Message WHERE messageId=$1 AND convId = $2;"
	result, err := tx.Exec(query, msgId, convId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	if err = tx.Commit(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	return result, nil
}

func CommentCheckUser(db *appdbimpl, msgId int64, convId int64, userId int64) (val int64, a bool, err error) {
	var commentId int64
	q := "SELECT commentId FROM Comment WHERE msgId= $1 AND convId = $2 AND userId = $3"
	err = db.c.QueryRow(q, msgId, convId, userId).Scan(&commentId)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	} else if err != nil {
		return 0, false, err
	}

	return commentId, commentId != 0, nil
}

func UpdateComment(db *appdbimpl, userId int64, msgId int64, convId int64, commentContent string, commentId int64) (err error) {
	tx, err := db.BeginTx()
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return err
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				err = nil
			}
		}
	}()

	q := `UPDATE Comment AS C
		  SET content = $1
		  WHERE  C.userId = $2 AND C.convId=$3 AND C.msgId = $4 AND C.commentId = $5`
	_, err = tx.Exec(q, commentContent, userId, convId, msgId, commentId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return err
	}

	return tx.Commit()
}

func (db *appdbimpl) RemoveComment(userId int64, convId int64, messageId int64) (err error) {
	tx, err := db.BeginTx()
	if err != nil {
		return err
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				err = nil
			}
		}
	}()

	q := `DELETE FROM Comment AS C
		  WHERE C.userId = $1 AND C.convId=$2 AND C.msgId = $3 `
	_, err = tx.Exec(q, userId, convId, messageId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return err
	}

	return tx.Commit()
}

func (db *appdbimpl) GetFinalComment(msgId int64, convId int64) ([]model.Comment, error) {
	var commentList []model.Comment
	var tmp model.Comment

	q := "SELECT * FROM Comment WHERE msgId= $1 AND convId = $2"
	rows, err := db.c.Query(q, msgId, convId)
	if errors.Is(err, sql.ErrNoRows) {
		return commentList, nil
	} else if err != nil {
		return commentList, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&tmp.Id, &tmp.MessageId, &tmp.ConversationId, &tmp.UserId, &tmp.UserName, &tmp.Content)
		if err != nil {
			return commentList, err
		}
		commentList = append(commentList, tmp)
	}

	if err = rows.Err(); err != nil {
		return commentList, err
	}
	return commentList, nil
}

func (db *appdbimpl) CommentMessage(msgId int64, convId int64, usrId int64, commentContent string) (val int64, err error) {
	tx, err := db.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				err = nil
			}
		}
	}()

	var ogId int64
	q := "SELECT COALESCE(MAX(commentId)+1,1) FROM Comment WHERE msgId = $1 and convId = $2"
	err = tx.QueryRow(q, msgId, convId).Scan(&ogId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return 0, err
	}

	cid, condition, err := CommentCheckUser(db, msgId, convId, usrId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return 0, err
	}

	if condition {
		// Update existing comment
		updateQ := `UPDATE Comment AS C
			  SET content = $1
			  WHERE  C.userId = $2 AND C.convId=$3 AND C.msgId = $4 AND C.commentId = $5`
		_, err = tx.Exec(updateQ, commentContent, usrId, convId, msgId, cid)
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			return 0, err
		}

		if err = tx.Commit(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			return 0, err
		}
		return cid, nil
	}

	// Create new comment
	q = "INSERT INTO Comment (commentId,msgId,convId,userId,userName,content) VALUES($1,$2,$3,$4,$5,$6)"
	name, err := db.GetUserName(usrId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return 0, err
	}

	_, err = tx.Exec(q, ogId, msgId, convId, usrId, name, commentContent)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return 0, err
	}

	if err = tx.Commit(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return 0, err
	}

	return ogId, nil
}

func (db *appdbimpl) PhotoMessage(picture model.Picture, messageId int64, msgInput model.Message, conversationId int64, userId int64) (err error) {
	tx, err := db.BeginTx()
	if err != nil {
		return err
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				err = nil
			}
		}
	}()

	if msgInput.Content == "" {
		msgInput.Content = "📷 Photo"
	}

	_, err = db.createMessageWithTx(tx, msgInput.Content, picture.Id, userId, conversationId, msgInput.RepliedId, msgInput.RepliedConvId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return err
	}

	return tx.Commit()
}
