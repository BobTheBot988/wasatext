package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"gitlab.com/mycompany8201046/myProject/service/api/model"
	"gitlab.com/mycompany8201046/myProject/service/api/reqcontext"
)

func (rt *_router) isAuthed(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) bool {
	rt.baseLogger.Info("Checking for auth")
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		rt.internalError(401, errors.New("no authorization header"), r, w)
		return false
	}

	// Check if the header starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		rt.internalError(401, errors.New("invalid authorization header"), r, w)
		return false
	}

	// Extract the token part (remove "Bearer " prefix)
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		rt.internalError(401, errors.New("no token provided"), r, w)
		return false
	}

	// Parse the token as userId (assuming the token is the userId)
	userId, err := strconv.ParseInt(token, 10, 64)
	if err != nil {
		rt.baseLogger.Error("Invalid token", err)
		rt.internalError(400, model.AddErrorString("Invalid token", err.Error()), r, w)
		return false
	}

	// Verify the user exists in the database
	_, err = rt.db.GetUserName(userId)
	if errors.Is(err, sql.ErrNoRows) {
		rt.internalError(401, model.AddErrorString("User not found:", err.Error()), r, w)
		return false
	} else if err != nil {
		rt.internalError(500, err, r, w)
		return false
	}

	return true
}
