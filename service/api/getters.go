package api

import (
	"net/http"

	"gitlab.com/mycompany8201046/myProject/service/api/model"
)

func (rt *_router) malformedUserIdReq(err error, w http.ResponseWriter, r *http.Request) bool {
	if err != nil {
		rt.internalError(400, model.AddError(model.ErrMalformedUserId, err), r, w)
		return true
	}
	return false
}

/*func (rt *_router) malformedGroupIdReq(err error, w http.ResponseWriter, r *http.Request) bool {
	if err != nil {
		rt.internalError(400, model.AddError(model.ErrMalformedGroupId, err), r, w)
		return true
	}
	return false
}*/

func (rt *_router) malformedConvIdReq(err error, w http.ResponseWriter, r *http.Request) bool {
	if err != nil {
		rt.internalError(400, model.AddError(model.ErrMalformedConvId, err), r, w)
		return true
	}
	return false

}
func (rt *_router) malformedMessageIdReq(err error, w http.ResponseWriter, r *http.Request) bool {
	if err != nil {
		rt.internalError(400, model.AddError(model.ErrMalformedMessageId, err), r, w)
		return true
	}
	return false
}
