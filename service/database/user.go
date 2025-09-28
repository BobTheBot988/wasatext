package database

import (
	"database/sql"
	"errors"

	"gitlab.com/mycompany8201046/myProject/service/api/model"
)

func checkUserId(db *appdbimpl, usrId int64) error {
	query := "SELECT userId FROM User WHERE userId= $1"
	var tmp *string
	err := db.c.QueryRow(query, usrId).Scan(&tmp)
	return err
}

func (db *appdbimpl) GetUsers() (*sql.Rows, error) {
	query := "SELECT userId,userName,IFNULL(userPhoto,'./images/defaultPP.png') AS userPhoto FROM User"
	rows, err := db.c.Query(query)
	return rows, err
}

func (db *appdbimpl) GetUsersNotInConversation(userId int64) (*sql.Rows, error) {
	query := "SELECT userId,userName,IFNULL(userPhoto,'./images/defaultPP.png') AS userPhoto FROM User WHERE userId NOT IN (SELECT usrId FROM Conv_User WHERE convId = $1)"
	rows, err := db.c.Query(query, userId)
	return rows, err
}

func (db *appdbimpl) GetUsersByConv(convId int64) (*sql.Rows, error) {
	query := "SELECT userId,userName,IFNULL(userPhoto,'./images/defaultPP.png') AS userPhoto  FROM User INNER JOIN Conv_User ON Conv_User.convId = $1 AND User.userId = Conv_User.usrId"
	rows, err := db.c.Query(query, convId)
	return rows, err
}

func (db *appdbimpl) CheckUsername(usrName string) (int64, error) {
	query := "SELECT userId FROM User WHERE userName= $1"
	var usrId int64
	err := db.c.QueryRow(query, usrName).Scan(&usrId)
	return usrId, err
}

func (db *appdbimpl) GetUserPhoto(userId int64) (string, error) {
	var p string
	query := "SELECT IFNULL(userPhoto,'./images/defaultPP.png') AS userPhoto FROM User where userId= $1"
	err := db.c.QueryRow(query, userId).Scan(&p)
	if err != nil {
		return "", model.AddError(errors.New("user Photo could not be found"), err)
	}
	return p, nil

}

func (db *appdbimpl) GetUserName(usrId int64) (string, error) {
	query := "SELECT userName FROM User WHERE userId=$1"
	var name string
	err := db.c.QueryRow(query, usrId).Scan(&name)
	if err != nil {
		return "", err
	}
	return name, err

}

func (db *appdbimpl) InsertUser(newUsrName string) (sql.Result, error) {
	query := "INSERT INTO User (userName) VALUES($1);"
	res, err := db.c.Exec(query, newUsrName)

	return res, err
}

func (db *appdbimpl) SetUserPhoto(pic model.Picture, usrId int64) (sql.Result, error) {
	q := "UPDATE User SET userPhoto =$1 WHERE userId = $2"
	return db.c.Exec(q, pic.Path, usrId)

}

func (db *appdbimpl) SetMyUserName(newUsrName string, usrId int64) (sql.Result, error) {
	var err, err2 error
	var result sql.Result

	err = checkUserId(db, usrId)
	_, err2 = db.CheckUsername(newUsrName)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		// the user does not exist
		result, err = db.InsertUser(newUsrName)
	case errors.Is(err2, sql.ErrNoRows):
		// A user with the same name does not exist
		query := "UPDATE User SET userName = $1 WHERE userId= $2"
		result, err = db.c.Exec(query, newUsrName, usrId)
	default:
		// There exist a user with the same name
		result = nil
		err = errors.New("you can't have two users with the same username")
	}

	return result, err
}
