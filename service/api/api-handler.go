package api

import (
	"net/http"
)

func (rt *_router) Handler() http.Handler {
	// Set up NotFound handler
	rt.router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("Custom 404 page: The requested resource was not found"))
		if err != nil {
			rt.baseLogger.Error(err)
		}
	})

	// GET methods
	// 	rt.router.GET("/", rt.wrap(rt.home_test))
	rt.router.GET("/users", rt.wrap(rt.getUsers))
	rt.router.GET("/users/:userId", rt.wrap(rt.getUsersNotInConversation))
	rt.router.GET("/users/:userId/conversations", rt.wrap(rt.getMyConversations))
	rt.router.GET("/users/:userId/conversations/:conversationId", rt.wrap(rt.getConversation))
	rt.router.GET("/users/:userId/photo", rt.wrap(rt.getUserPicture))
	rt.router.GET("/groups/:groupId", rt.wrap(rt.getGroupInfo))
	rt.router.GET("/groups/:groupId/users", rt.wrap(rt.getGroupUsers))
	rt.router.GET("/groups/:groupId/photo", rt.wrap(rt.getGroupPicture))
	rt.router.GET("/photos/:photoId", rt.wrap(rt.getPhoto))
	rt.router.GET("/users/:userId/conversations/:conversationId/messages/:messageId/status", rt.wrap(rt.getMessageStatus))
	rt.router.GET("/users/:userId/conversations/:conversationId/messages/:messageId/comments", rt.wrap(rt.getComments))
	rt.router.GET("/conversations/:conversationId/messages/:messageId", rt.wrap(rt.getMessage))

	// POST methods
	rt.router.POST("/session", rt.wrap(rt.doLogin))
	rt.router.POST("/groups", rt.wrap(rt.createGroup))
	rt.router.POST("/groups/:groupId/users/:userId", rt.wrap(rt.addToGroup))
	rt.router.POST("/groups/:groupId/name", rt.wrap(rt.setGroupName))
	rt.router.POST("/groups/:groupId/desc", rt.wrap(rt.setGroupDesc))
	rt.router.POST("/conversations/create", rt.wrap(rt.createConversation))
	rt.router.POST("/users/:userId/conversations/:conversationId/messages", rt.wrap(rt.sendMessage))
	rt.router.POST("/users/:userId/conversations/:conversationId/messages/photo", rt.wrap(rt.sendPhotoMessage))
	rt.router.POST("/users/:userId/conversations/:conversationId/messages/read/:messageId", rt.wrap(rt.readMessage))
	rt.router.POST("/users/:userId/conversations/:conversationId/messages/forward/:messageId", rt.wrap(rt.forwardMessage))
	rt.router.POST("/users/:userId/conversations/:conversationId/messages/comments/:messageId", rt.wrap(rt.commentMessage))
	rt.router.POST("/users/:userId/username", rt.wrap(rt.setMyUserName))
	rt.router.POST("/photos/:photoId/users/:userId/photo", rt.wrap(rt.setMyPhoto))
	rt.router.POST("/photos/:photoId/groups/:groupId/photo", rt.wrap(rt.setGroupPhoto))

	// DELETE methods
	rt.router.DELETE("/users/:userId/conversations/:conversationId/messages/:messageId/comments/:commentId", rt.wrap(rt.uncommentMessage))
	rt.router.DELETE("/users/:userId/conversations/:conversationId/messages/:messageId", rt.wrap(rt.deleteMessage))
	rt.router.DELETE("/groups/:groupId/users/:userId", rt.wrap(rt.leaveGroup))

	return rt.router
}
