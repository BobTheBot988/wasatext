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

func (rt *_router) getMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if !rt.isAuthed(w, r, ps, ctx) {
		return
	}

	var err error
	var convId int64
	var message model.Message

	rt.baseLogger.Info("Getting Message")
	convId, err = strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if err != nil {
		rt.internalError(400, model.AddError(model.ErrMalformedConvId, err), r, w)
		return
	}

	message.Id, err = strconv.ParseInt(ps.ByName("messageId"), 10, 64)
	if err != nil {
		rt.internalError(400, model.AddError(model.ErrMalformedMessageId, err), r, w)
		return
	}

	row := rt.db.GetMessage(message.Id, convId)
	err = row.Scan(&message.Id,
		&message.Content,
		&message.Timestamp,
		&message.Sender.UserId,
		&message.ConvId,
		&message.PictureId,
		&message.RepliedId,
		&message.RepliedConvId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(200)
		_, e := w.Write(nil)
		if e != nil {
			rt.baseLogger.Error(e)
			return
		}
		return
	} else if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	w.WriteHeader(200)
	err = json.NewEncoder(w).Encode(message)
	if err != nil {
		rt.baseLogger.Error("getMessage error:", err)
		return
	}
}

func (rt *_router) getMessageStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var err error
	var msgId, usrId, convId int64
	var messageReadStatus model.MessageReadStatus
	rt.baseLogger.Info("Getting Message Status")
	usrId, err = strconv.ParseInt(ps.ByName("userId"), 10, 64)
	if err != nil {
		rt.internalError(400, model.AddError(model.ErrMalformedUserId, err), r, w)
		return
	}

	convId, err = strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if err != nil {
		rt.internalError(400, model.AddError(model.ErrMalformedConvId, err), r, w)
		return
	}

	msgId, err = strconv.ParseInt(ps.ByName("messageId"), 10, 64)
	if err != nil {
		rt.internalError(400, model.AddError(model.ErrMalformedMessageId, err), r, w)
		return
	}

	messageReadStatus.HasBeenRead, err = rt.db.HasMessageBeenRead(msgId, convId, usrId)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

	/* // TODO Extra implement this to see who has read the message
	rows, err = rt.db.WhoHasReadMessage(msgId, convId, usrId)
	for rows.Next() {
		err = rows.Scan()
		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}
	}*/

	w.WriteHeader(200)
	err = json.NewEncoder(w).Encode(messageReadStatus)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

}

func (rt *_router) readMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var err error
	var msgId, usrId, convId int64

	rt.baseLogger.Info("Reading Message")
	usrId, err = strconv.ParseInt(ps.ByName("userId"), 10, 64)
	if rt.malformedUserIdReq(err, w, r) {
		return
	}

	convId, err = strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if rt.malformedConvIdReq(err, w, r) {
		return
	}
	msgId, err = strconv.ParseInt(ps.ByName("messageId"), 10, 64)
	if rt.malformedMessageIdReq(err, w, r) {
		return
	}
	err = rt.db.ReadMessage(msgId, convId, usrId)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	w.WriteHeader(204)
}

func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var err error
	var usrId, convId int64
	var msgInput model.MessageInput
	var msgId model.MessageId

	rt.baseLogger.Info("Sending Message")
	usrId, err = strconv.ParseInt(ps.ByName("userId"), 10, 64)
	if rt.malformedUserIdReq(err, w, r) {
		return
	}

	convId, err = strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if rt.malformedConvIdReq(err, w, r) {
		return
	}

	err = json.NewDecoder(r.Body).Decode(&msgInput)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	if len(msgInput.Content) > 1000 {
		e := errors.New("the message cannot be bigger than a 1000 chars")
		rt.internalError(400, e, r, w)
		return

	}

	msgId.Value, err = rt.db.CreateMessage(msgInput, 0, usrId, convId)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	err = json.NewEncoder(w).Encode(msgId)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	w.WriteHeader(http.StatusCreated)

}

func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var err error
	var usrId, convId, messageId int64
	var comment model.Comment

	rt.baseLogger.Info("Commenting Message")
	usrId, err = strconv.ParseInt(ps.ByName("userId"), 10, 64)
	if rt.malformedUserIdReq(err, w, r) {
		return
	}

	convId, err = strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if rt.malformedConvIdReq(err, w, r) {
		return
	}
	messageId, err = strconv.ParseInt(ps.ByName("messageId"), 10, 64)
	if rt.malformedMessageIdReq(err, w, r) {
		return
	}
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	if len(comment.Content) > 20 {
		e := errors.New("the message cannot be bigger then  20 chars")
		rt.internalError(400, e, r, w)
		return
	}

	comment.Id, err = rt.db.CommentMessage(messageId, convId, usrId, comment.Content)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	w.WriteHeader(204)
}

func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var err error
	var usrId, convId, messageId int64

	rt.baseLogger.Info("Uncommenting Message")
	usrId, err = strconv.ParseInt(ps.ByName("userId"), 10, 64)
	if rt.malformedUserIdReq(err, w, r) {
		return
	}

	convId, err = strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if rt.malformedConvIdReq(err, w, r) {
		return
	}
	messageId, err = strconv.ParseInt(ps.ByName("messageId"), 10, 64)
	if rt.malformedMessageIdReq(err, w, r) {
		return
	}
	err = rt.db.RemoveComment(usrId, convId, messageId)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

	w.WriteHeader(204)
}

func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var err error
	var convId, msgId int64

	rt.baseLogger.Info("Deleting Message")
	convId, err = strconv.ParseInt(ps.ByName("conversationId"), 10, 64)

	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	msgId, err = strconv.ParseInt(ps.ByName("messageId"), 10, 64)
	if rt.malformedMessageIdReq(err, w, r) {
		return
	}
	_, err = rt.db.DeleteMessage(msgId, convId)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	w.WriteHeader(204)

}

func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	var err error
	var usrId, convId, msgId int64
	var body model.MsgForward

	rt.baseLogger.Info("Forwarding message")
	usrId, err = strconv.ParseInt(ps.ByName("userId"), 10, 64)
	if rt.malformedUserIdReq(err, w, r) {
		return
	}

	convId, err = strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if rt.malformedConvIdReq(err, w, r) {
		return
	}
	msgId, err = strconv.ParseInt(ps.ByName("messageId"), 10, 64)
	if rt.malformedMessageIdReq(err, w, r) {
		return
	}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	_, err = rt.db.ForwardMessage(msgId, convId, usrId, body.ConvId)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
	w.WriteHeader(204)

}

func (rt *_router) sendPhotoMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	rt.baseLogger.Info("Sending photo message")
	rt.uploadPhotoHandler(w, r, ps, ctx, 2)

}
