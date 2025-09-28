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

func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	if !rt.isAuthed(w, r, ps, ctx) {
		return
	}
	var group model.Group
	rt.baseLogger.Info("Creating Group")

	err := json.NewDecoder(r.Body).Decode(&group)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	_, err = rt.db.CreateGroup(group.Name, group.UserId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		rt.internalError(500, err, r, w)
		return
	}
	w.WriteHeader(204)
}
func (rt *_router) getGroupInfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if !rt.isAuthed(w, r, ps, ctx) {
		return
	}
	var groupPw model.GroupPw

	groupId, err := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if err != nil {
		rt.internalError(400, err, r, w)
		return
	}
	row := rt.db.GetGroupInfo(groupId)
	err = row.Scan(&groupPw.Name, &groupPw.Desc, &groupPw.Pic)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	w.WriteHeader(200)
	err = json.NewEncoder(w).Encode(groupPw)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

}

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if !rt.isAuthed(w, r, ps, ctx) {
		return
	}
	groupId, err := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if err != nil {
		rt.internalError(400, err, r, w)
		return
	}
	userId, err := strconv.ParseInt(ps.ByName("userId"), 10, 64)
	if err != nil {
		rt.internalError(400, err, r, w)
		return
	}
	_, err = rt.db.LeaveGroup(userId, groupId)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	w.WriteHeader(204)
}

func (rt *_router) getGroupUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if !rt.isAuthed(w, r, ps, ctx) {
		return
	}
	var userList []model.User
	var tmpUser model.User

	groupId, err := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if err != nil {
		rt.internalError(400, err, r, w)
		return
	}
	rows, err := rt.db.GetUsersByGroup(groupId)
	if err != nil {
		rt.internalError(500, err, r, w)
		return

	}

	for rows.Next() {
		err = rows.Scan(
			&tmpUser.UserId,
			&tmpUser.Name,
			&tmpUser.UserPhoto,
		)
		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}
		userList = append(userList, tmpUser)

	}
	defer rows.Close()
	if rows.Err() != nil {
		rt.internalError(500, rows.Err(), r, w)
		return
	}
	w.WriteHeader(200)
	err = json.NewEncoder(w).Encode(userList)

	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

}
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId, err := strconv.ParseInt(ps.ByName("userId"), 10, 64)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		rt.internalError(400, err, r, w)
		return
	}
	groupId, err := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		rt.internalError(400, err, r, w)
		return
	}
	var userList []int64
	userList = append(userList, userId)
	_, err = rt.db.AddGroup(userList, groupId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		rt.internalError(500, err, r, w)
		return
	}
	w.WriteHeader(204)
}

func (rt *_router) getGroupPicture(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if !rt.isAuthed(w, r, ps, ctx) {
		return
	}
	var pathContainer model.Path
	groupId, err := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		rt.internalError(400, err, r, w)
		return
	}
	pathContainer.Path, err = rt.db.GetGroupPhoto(groupId)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		rt.internalError(500, err, r, w)
		return
	}
	err = rt.SendImageBack(w, pathContainer.Path)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		rt.internalError(500, err, r, w)
		return
	}
}
func (rt *_router) setGroupDesc(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if !rt.isAuthed(w, r, ps, ctx) {
		return
	}
	var grp model.GroupPw
	groupId, err := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		rt.internalError(400, err, r, w)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&grp)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		rt.internalError(500, err, r, w)
		return
	}
	rt.baseLogger.Infof("Id:%d", grp.Id)
	rt.baseLogger.Infof("Description:%s\n", grp.Desc)
	_, err = rt.db.SetGroupDesc(grp.Desc, groupId)

	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		rt.internalError(500, err, r, w)
		return
	}
	w.WriteHeader(204)
}
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if !rt.isAuthed(w, r, ps, ctx) {
		return
	}
	var grp model.GroupPw
	groupId, err := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		rt.internalError(400, err, r, w)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&grp)
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		rt.internalError(500, err, r, w)
		return
	}
	_, err = rt.db.SetGroupName(grp.Name, groupId)

	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		rt.internalError(500, err, r, w)
		return
	}
	w.WriteHeader(204)
}
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	rt.uploadPhotoHandler(w, r, ps, ctx, 1)
}
