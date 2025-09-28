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

func (rt *_router) internalError(errcode int32, e error, r *http.Request, w http.ResponseWriter) {
	var er model.CustomError
	rt.baseLogger.Error(e)
	rt.baseLogger.Print("Body of request", r.Body)
	er.Code = errcode
	er.Message = e.Error()
	w.WriteHeader(int(er.Code))
	err := json.NewEncoder(w).Encode(er)
	if err != nil {
		rt.baseLogger.Error("Error in writing to client:", err)
	}

}

// createConversation handles the creation of a new conversation between specified users.
// It first checks authentication, then decodes the user list from the request body.
// If a conversation already exists between the users, it returns a 401 status.
// On successful creation, it returns the new conversation ID with a 201 status.
func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if !rt.isAuthed(w, r, ps, ctx) {
		return
	}

	var userList model.UserIdList
	var convId model.ConversationId
	err := json.NewDecoder(r.Body).Decode(&userList)
	if err != nil {
		rt.internalError(400, err, r, w)
		return
	}

	rt.baseLogger.Info("Creating Conversation: \nUserIds:", userList.UserId)
	convId.Value, err = rt.db.CreateConversation(userList.UserId)
	if convId.Value == 0 {
		rt.internalError(500, errors.New("the conversation was not created there was a problem"), r, w)
		return
	}
	if errors.Is(err, errors.New("the chat already exists")) {
		rt.internalError(400, err, r, w)
		return
	} else if err != nil && !errors.Is(err, sql.ErrTxDone) {
		rt.internalError(500, err, r, w)
		return
	}

	w.WriteHeader(201)
	err = json.NewEncoder(w).Encode(convId)
	if err != nil {
		rt.baseLogger.Error(err)
		return
	}
}
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	rt.baseLogger.Info("Getting conversations:")
	if !rt.isAuthed(w, r, ps, ctx) {
		return
	}
	var sql_rows *sql.Rows
	var err error
	var usrId int64
	var convs []model.ConversationPw
	var tmp_conv model.ConversationPw

	usrId, err = strconv.ParseInt(ps.ByName("userId"), 10, 64)
	if err != nil {
		rt.internalError(400, err, r, w)
		rt.baseLogger.Error(err)
		return
	}

	rt.baseLogger.Info("Getting Conversations:")
	sql_rows, err = rt.db.GetConversations(usrId)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	for sql_rows.Next() {
		err = sql_rows.Scan(
			&tmp_conv.ConversationId,
			&tmp_conv.LastMsgContent,
			&tmp_conv.LastMsgTimeStamp,
			&tmp_conv.GroupId,
			&tmp_conv.UserId,
		)
		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}

		rt.baseLogger.Info("Getting Name Of conversation")
		tmp_conv.Name, err = rt.db.GetConvName(tmp_conv.ConversationId, usrId)
		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}
		rt.baseLogger.Info("Getting Picture of Conversation")
		tmp_conv.Photo, err = rt.db.GetConversationPhoto(tmp_conv.ConversationId, usrId)
		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}
		convs = append(convs, tmp_conv)
	}
	defer sql_rows.Close()

	if sql_rows.Err() != nil {
		rt.internalError(500, err, r, w)
		return
	}

	/*
	 *	 Get every group and every single conversation and return them in a json format with pictures included
	 */

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(convs)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

}

func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	rt.baseLogger.Info("Getting SINGLE conversation:")
	if !rt.isAuthed(w, r, ps, ctx) {
		return
	}
	var conv model.Conversation
	var err error
	conv.Id, err = strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if err != nil {
		rt.internalError(400, err, r, w)
		rt.baseLogger.Error(err)
		return
	}

	rows, err := rt.db.GetConversation(conv.Id)

	if err != nil {
		rt.internalError(500, err, r, w)
		rt.baseLogger.Error(err)
		return
	}

	for rows.Next() {
		var message model.Message
		err = rows.Scan(&message.Id,
			&message.Content,
			&message.Timestamp,
			&message.Sender.UserId,
			&message.ConvId,
			&message.PictureId,
			&message.RepliedId,
			&message.RepliedConvId)

		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}

		rt.PrintNumberOfOpenConnections()
		message.Sender.Name, err = rt.db.GetUserName(message.Sender.UserId)
		rt.PrintNumberOfOpenConnections()
		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}

		rt.PrintNumberOfOpenConnections()
		message.CommentList, err = rt.db.GetFinalComment(message.Id, message.ConvId)
		rt.PrintNumberOfOpenConnections()

		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}
		conv.Messages = append(conv.Messages, message)
	}
	defer rows.Close()
	if rows.Err() != nil {
		rt.internalError(500, err, r, w)
		return
	}
	rt.PrintNumberOfOpenConnections()
	user_rows, err := rt.db.GetUsersByConv(conv.Id)
	rt.PrintNumberOfOpenConnections()
	if err != nil {
		rt.internalError(500, err, r, w)
	}

	for user_rows.Next() {
		var user model.User
		err = user_rows.Scan(
			&user.UserId,
			&user.Name,
			&user.UserPhoto,
		)
		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}
		conv.Users = append(conv.Users, user)

	}
	defer user_rows.Close()
	if user_rows.Err() != nil {
		rt.internalError(500, err, r, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(conv)

	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	// row, err := rt.db.GetConversation(id)
}
