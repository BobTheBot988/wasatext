package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"gitlab.com/mycompany8201046/myProject/service/api/model"
	"gitlab.com/mycompany8201046/myProject/service/api/reqcontext"
)

func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var err error
	var User model.User
	rt.baseLogger.Info("Setting username")
	// The username is in the request
	err = json.NewDecoder(r.Body).Decode(&User)
	if err != nil {
		rt.baseLogger.Error(err)

		rt.internalError(500, err, r, w)
		return
	}
	//	User.UserId, _ = strconv.ParseInt(ps.ByName("userId"), 10, 64)
	_, err = rt.db.SetMyUserName(User.Name, User.UserId)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

	w.WriteHeader(204)
}
func (rt *_router) getUserPicture(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var pathContainer model.Path
	var err error
	rt.baseLogger.Info("Get User Picture")
	userId, err := strconv.ParseInt(ps.ByName("userId"), 10, 64)
	if err != nil {
		rt.internalError(400, model.AddError(model.ErrMalformedUserId, err), r, w)
		return
	}
	pathContainer.Path, err = rt.db.GetUserPhoto(userId)

	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

	err = rt.SendImageBack(w, pathContainer.Path)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
}
func (rt *_router) getUsersNotInConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var User model.User
	var UserList []model.User

	rt.baseLogger.Info("Getting Not In conversation")
	w.Header().Set("Content-Type", "application/json")

	userId, err := strconv.ParseInt(ps.ByName("userId"), 10, 64)
	if err != nil {
		rt.internalError(400, model.AddError(model.ErrMalformedUserId, err), r, w)
		return
	}
	rows, err := rt.db.GetUsersNotInConversation(userId)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

	for rows.Next() {
		err = rows.Scan(
			&User.UserId,
			&User.Name,
			&User.UserPhoto,
		)
		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}
		UserList = append(UserList, User)
	}

	defer rows.Close()
	if rows.Err() != nil {
		rt.internalError(500, err, r, w)
		return
	}

	w.WriteHeader(200)
	err = json.NewEncoder(w).Encode(UserList)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
}
func (rt *_router) getUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var User model.User
	var UserList []model.User

	rt.baseLogger.Info("Getting Users")
	w.Header().Set("Content-Type", "application/json")

	rows, err := rt.db.GetUsers()
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

	for rows.Next() {
		err = rows.Scan(
			&User.UserId,
			&User.Name,
			&User.UserPhoto,
		)
		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}
		UserList = append(UserList, User)
	}

	defer rows.Close()
	if rows.Err() != nil {
		rt.internalError(500, err, r, w)
		return
	}

	w.WriteHeader(200)
	err = json.NewEncoder(w).Encode(UserList)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
}
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var err error
	var User model.User
	w.Header().Set("Content-Type", "application/json")
	rt.baseLogger.Info("Logging In")

	err = json.NewDecoder(r.Body).Decode(&User)
	rt.baseLogger.Infof("Username: %s\n", User.Name)
	if User.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

	rt.baseLogger.Println("Checking if username already exists...")
	User.UserId, err = rt.db.CheckUsername(User.Name)
	if errors.Is(err, sql.ErrNoRows) {
		rt.baseLogger.Infof("The user was not found creating it: %s\n", User.Name)
		var result sql.Result
		result, err = rt.db.InsertUser(User.Name)
		if err != nil {
			rt.baseLogger.Errorf("The user could not be created: %v\n", err)
			rt.internalError(500, model.AddErrorString("The user could not be created", err.Error()), r, w)
			return
		}
		User.UserId, err = result.LastInsertId()

		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}
		User.UserPhoto, err = rt.db.GetUserPhoto(User.UserId)
		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}
	} else {
		rt.baseLogger.Infof("The user was found userId: %d.\n", User.UserId)
		w.WriteHeader(http.StatusAccepted)
		User.UserPhoto, err = rt.db.GetUserPhoto(User.UserId)
		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}
		rt.baseLogger.Infof("The user photo is   : %s", User.UserPhoto)
		err = json.NewEncoder(w).Encode(User)
		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}
		return
	}

	rt.baseLogger.Infof("UserId %d\n", User.UserId)
	User.UserPhoto, err = rt.db.GetUserPhoto(User.UserId)
	rt.baseLogger.Infof("The user photo is   : %s", User.UserPhoto)
	if err == nil {
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(User)
		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}
		rt.baseLogger.Infof("User successfully logged!!\n")
	} else {
		rt.baseLogger.Errorf("Error %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	rt.uploadPhotoHandler(w, r, ps, ctx, 0)
}
