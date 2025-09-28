package database

import (
	"database/sql"
	"errors"

	"gitlab.com/mycompany8201046/myProject/service/api/model"
)

/* func getConvId(db *appdbimpl, grpId int64) (int64, error) {
	var convId int64
	var err error
	query := "SELECT convId FROM GroupTB WHERE groupId = $1;"
	err = db.c.QueryRow(query, grpId).Scan(&convId)
	return convId, err
} */

func getConvIdWithTx(tx *HookedTx, grpId int64) (int64, error) {
	var convId int64
	var err error
	query := "SELECT convId FROM GroupTB WHERE groupId = $1;"
	err = tx.QueryRow(query, grpId).Scan(&convId)
	if errors.Is(err, sql.ErrTxDone) {
		err = nil
	}
	return convId, err
}

/* func check_user_in_group(db *appdbimpl, groupId int64, user int64) (bool, error) {
	query := "SELECT groupId FROM Group_User WHERE userId = $1 AND groupId = $2"
	var g int64
	err := db.c.QueryRow(query, user, groupId).Scan(&g)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return true, err
	}
	return false, nil
} */

func checkUserInGroupWithTx(tx *HookedTx, groupId int64, user int64) (bool, error) {
	query := "SELECT groupId FROM Group_User WHERE userId = $1 AND groupId = $2"
	var g int64
	err := tx.QueryRow(query, user, groupId).Scan(&g)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return true, err
	}
	return false, nil
}

func (db *appdbimpl) CreateGroup(g_name string, usrId []int64) (res sql.Result, err error) {
	if len(usrId) < 3 {
		return nil, errors.New("you need at least three users to create a new group")
	}

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

	var result sql.Result
	var query string
	var conv_id int64

	conv_id, err = db.createConversationWithTx(tx, usrId)
	if err != nil {
		return nil, err
	}

	query = "INSERT INTO GroupTB (ConvId,Name) VALUES($1,$2)"
	result, err = tx.Exec(query, conv_id, g_name)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	group_id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	query = "INSERT INTO Group_User (groupId,userId) VALUES($1,$2)"
	for _, i := range usrId {
		result, err = tx.Exec(query, group_id, i)
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			return nil, err
		}
	}

	// Commit with hooks
	if err = tx.Commit(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	return result, nil
}

// Helper method to create conversation within existing transaction
func (db *appdbimpl) createConversationWithTx(tx *HookedTx, users []int64) (int64, error) {
	if len(users) < 2 {
		return 0, errors.New("not enough users to create a conversation")
	}

	if len(users) == 2 {
		exists, err := db.checkConvAlreadyExists(tx, users)
		if err != nil {
			return 0, err
		}
		if exists {
			return 0, errors.New("the chat already exists")
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

	return last_inserted_id, nil
}

/*
	func (db *appdbimpl) GetUsersInGroup(groupId int64) (*sql.Rows, error) {
		query := "SELECT User.* FROM User RIGHT JOIN Group_User ON User.userId = Group_User.userId AND Group_User.groupId = $1"
		rows, err := db.c.Query(query, groupId)
		if err != nil {
			return nil, err
		}
		return rows, nil
	}
*/

func (db *appdbimpl) GetGroupInfo(groupId int64) *sql.Row {
	query := "SELECT G.Name,IFNULL(G.Description,'')AS Description,IFNULL(photo,'./images/defaultPP.png') AS Photo FROM GroupTB AS G WHERE groupId = $1"
	return db.c.QueryRow(query, groupId)
}

func (db *appdbimpl) GetUsersByGroup(groupId int64) (res *sql.Rows, err error) {
	query := "SELECT User.userId,User.userName, IFNULL(User.userPhoto,'') AS photo FROM Group_User INNER JOIN USER ON User.userId = Group_User.userId AND Group_User.groupId = $1"
	return db.c.Query(query, groupId)
}

func (db *appdbimpl) AddGroup(newUserIds []int64, grpId int64) (sql.Result, error) {
	tx, err := db.BeginTx()

	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			if errors.Is(err, sql.ErrTxDone) {
				err = nil
			}
		}
	}()

	var result sql.Result
	var convId int64
	var is_in_group bool

	convId, err = getConvIdWithTx(tx, grpId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	query := "INSERT INTO Conv_User (convId,usrId) VALUES($1,$2)"
	for _, userId := range newUserIds {
		is_in_group, err = checkUserInGroupWithTx(tx, grpId, userId)
		if is_in_group {
			return nil, errors.New("user: is in group")
		}
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			return nil, err
		}

		result, err = tx.Exec(query, convId, userId)
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			return nil, err
		}

		q := "INSERT INTO Group_User (groupId,userId) VALUES($1,$2)"
		_, err = tx.Exec(q, grpId, userId)
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			return nil, err
		}
	}

	// Commit with hooks
	if err = tx.Commit(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	return result, nil
}

func (db *appdbimpl) GetGroupPhoto(groupId int64) (string, error) {
	var p string
	query := "SELECT IFNULL(photo,'./images/defaultPP.png') AS photo FROM GroupTB WHERE groupId = $1"
	err := db.c.QueryRow(query, groupId).Scan(&p)
	if err != nil {
		return "", err
	}
	return p, nil
}

func (db *appdbimpl) LeaveGroup(usrId int64, grpId int64) (res sql.Result, err error) {
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

	var convId int64

	convId, err = getConvIdWithTx(tx, grpId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	query := "DELETE FROM Conv_User WHERE usrId = $1 AND convId = $2"
	_, err = tx.Exec(query, usrId, convId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	query = "DELETE FROM Group_User WHERE userId = $1 AND groupId = $2"
	result, err := tx.Exec(query, usrId, grpId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	// Commit with hooks
	if err = tx.Commit(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	return result, nil
}

func (db *appdbimpl) SetGroupDesc(newDesc string, grpId int64) (res sql.Result, err error) {
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

	var result sql.Result

	query := "UPDATE GroupTB SET Description = $1 WHERE groupId = $2"
	result, err = tx.Exec(query, newDesc, grpId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	// Commit with hooks
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

func (db *appdbimpl) SetGroupName(newName string, grpId int64) (res sql.Result, err error) {
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

	var result sql.Result

	query := "UPDATE GroupTB SET Name = $1 WHERE groupId = $2"
	result, err = tx.Exec(query, newName, grpId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	// Commit with hooks
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

func (db *appdbimpl) SetGroupPhoto(pic model.Picture, grpId int64) (res sql.Result, err error) {
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

	q := "UPDATE GroupTB SET photo = $1 WHERE groupId = $2"
	result, err := tx.Exec(q, pic.Path, grpId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}
	// Commit with hooks
	if err = tx.Commit(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return nil, err
	}

	return result, nil
}
