package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"gitlab.com/mycompany8201046/myProject/service/api/model"
	"gitlab.com/mycompany8201046/myProject/service/api/reqcontext"
)

func (rt *_router) getComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	rt.baseLogger.Info("Getting Comments")
	if !rt.isAuthed(w, r, ps, ctx) {
		return
	}

	var CommentList []model.Comment
	var TmpComment model.Comment

	convId, err := strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if err != nil {
		rt.internalError(400, err, r, w)
		rt.baseLogger.Error(err)
		return
	}

	messageId, err := strconv.ParseInt(ps.ByName("messageId"), 10, 64)
	if err != nil {
		rt.internalError(400, err, r, w)
		rt.baseLogger.Error(err)
		return
	}

	rt.PrintNumberOfOpenConnections()
	rows, err := rt.db.GetComments(convId, messageId)

	rt.PrintNumberOfOpenConnections()
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

	for rows.Next() {
		err = rows.Scan(
			&TmpComment.Id,
			&TmpComment.Content,
			&TmpComment.MessageId,
			&TmpComment.ConversationId,
			&TmpComment.UserId,
			&TmpComment.UserName,
		)

		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}
		CommentList = append(CommentList, TmpComment)
	}
	defer rows.Close()
	err = rows.Err()

	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

	w.WriteHeader(200)
	err = json.NewEncoder(w).Encode(CommentList)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

}
