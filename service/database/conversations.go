package database

import (
	"database/sql"
	"errors"
	"fmt"
)

func (db *appdbimpl) GetConversation(convId int64) (*sql.Rows, error) {
	query := "SELECT messageId,content,mtime,usrSenderId,convId,IFNULL(photoId,-1),IFNULL(repliedId,0),IFNULL(repliedConvId,0) AS photoId FROM Message AS M WHERE M.convId = $1"
	var conversation *sql.Rows

	conversation, err := db.c.Query(query, convId)
	if err != nil {
		defer conversation.Close()
		return nil, err
	}

	return conversation, err
}

func (db *appdbimpl) GetConvName(convId int64, usrId int64) (string, error) {
	query := `SELECT Name FROM GroupTB WHERE GroupTB.convId = $1`
	var name string
	err := db.c.QueryRow(query, convId).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		query = `SELECT userName FROM User INNER JOIN Conv_User ON Conv_User.convId = $1 AND User.userId = Conv_User.usrId AND userId != $2`
		err = db.c.QueryRow(query, convId, usrId).Scan(&name)
		if err != nil {

			return "", fmt.Errorf("error: \n convId:%d \n %s", convId, err.Error())
		}
	} else if err != nil {
		return "", err
	}

	return name, nil
}

func (db *appdbimpl) GetConversationPhoto(convId int64, userId int64) (string, error) {
	var groupId int64
	q := "SELECT groupId FROM GroupTB WHERE convId = $1"
	err := db.c.QueryRow(q, convId).Scan(&groupId)
	var photo string
	if errors.Is(err, sql.ErrNoRows) || groupId == 0 {
		q = "SELECT IFNULL(userPhoto,'images/defaultPP.png') AS userPhoto FROM User WHERE userId == (SELECT usrId FROM Conv_User WHERE convId = 1 AND usrId !=1) "
		err = db.c.QueryRow(q, convId, userId).Scan(&photo)
	} else if err == nil {
		q = "SELECT IFNULL(photo,'images/defaultPP.png') AS photo FROM GroupTB WHERE groupId = $1"
		err = db.c.QueryRow(q, groupId).Scan(&photo)

	}
	return photo, err
}

func (db *appdbimpl) GetConversations(usrId int64) (*sql.Rows, error) {
	/* *********************************************************
		query := `SELECT C.*,IFNULL(G.groupId,0)  AS groupId FROM (SELECT conversationId,
		 IFNULL(Message.content, '') AS content,
	  IFNULL(Message.mtime, unixepoch()) AS mtime FROM Conversation INNER JOIN Conv_User ON  Conversation.conversationId = Conv_User.convId AND Conv_User.usrId=$1
				  LEFT JOIN Message ON Conversation.lastMsgId = Message.messageId AND Message.convId = Conversation.conversationId) AS C LEFT JOIN GroupTB AS G ON G.convId = C.conversationId`
	*/
	query := `SELECT C.*,IFNULL(G.groupId,0) AS groupId,IFNULL(CU.usrId,0) AS userId FROM 
			  ((SELECT conversationId, 
			  IFNULL(Message.content, '') AS content, 
			  IFNULL(Message.mtime, unixepoch()) AS mtime FROM Conversation INNER JOIN Conv_User ON  Conversation.conversationId = Conv_User.convId AND Conv_User.usrId=$1 
			  LEFT JOIN Message ON Conversation.lastMsgId = Message.messageId AND Message.convId = Conversation.conversationId) AS C 
			  LEFT JOIN GroupTB AS G ON G.convId = C.conversationId) AS GC
			  LEFT JOIN Conv_User AS CU ON CU.convId = GC.conversationId  AND CU.usrId != $1  AND GC.groupId IS NULL`

	var conversations *sql.Rows
	var err error

	conversations, err = db.c.Query(query, usrId)
	if err != nil {
		defer conversations.Close()
		return nil, err
	}

	return conversations, err
}

// * The query can be done prettier by using Triggers
func (db *appdbimpl) addUsersToConv(tx *HookedTx, users []int64, last_inserted_conv int64) error {
	var err error

	query := `INSERT INTO Conv_User (convId,usrId) VALUES($1,$2)`

	for _, v := range users {
		_, err = tx.Exec(query, last_inserted_conv, v)
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			return err
		}
	}

	if len(users) == 2 {
		query = `INSERT INTO User_Chat (usrId1,usrId2,convId) VALUES($1,$2,$3)`
		_, err = tx.Exec(query, users[0], users[1], last_inserted_conv)
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			return err
		}
	}
	return nil
}

func (db *appdbimpl) checkConvAlreadyExists(tx *HookedTx, users []int64) (bool, error) {
	query := `SELECT CU.convId FROM Conv_User CU 
			  WHERE CU.convId NOT IN (SELECT convId FROM GroupTB)
			  AND CU.convId IN (
				  SELECT convId FROM Conv_User 
				  WHERE usrId = $1
			  )
			  AND CU.convId IN (
				  SELECT convId FROM Conv_User 
				  WHERE usrId = $2
			  )
			  GROUP BY CU.convId 
			  HAVING COUNT(CU.usrId) = 2`
	var c_id int64

	err := tx.QueryRow(query, users[0], users[1]).Scan(&c_id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return true, err
	}

	if c_id > 0 {
		return true, nil
	}

	return false, nil
}

func (db *appdbimpl) CreateConversation(users []int64) (val int64, err error) {
	if len(users) < 2 {
		return 0, errors.New("not enough users to create a conversation")
	}

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

	if len(users) == 2 {
		var exists bool
		exists, err = db.checkConvAlreadyExists(tx, users)
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			return 0, err
		}
		if exists {
			return 0, errors.New("the chat aready exists")
		}
	}

	query := `INSERT INTO Conversation (lastMsgId) VALUES(0)`

	result, err := tx.Exec(query)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return 0, err
	}

	last_inserted_id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Add users to conversation within the same transaction
	err = db.addUsersToConv(tx, users, last_inserted_id)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return 0, err
	}

	// Commit with hooks
	if err = tx.Commit(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return 0, err
	}

	return last_inserted_id, nil

}
