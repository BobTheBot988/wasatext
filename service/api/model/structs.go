package model

import "errors"

var ErrMalformedUserId = errors.New("userId is not correct")
var ErrMalformedConvId = errors.New("conversationId is not correct")
var ErrMalformedMessageId = errors.New("messageId is not correct")
var ErrMalformedGroupId = errors.New("groupId is not correct")
var ErrMalformedPhotoId = errors.New("photoId is not correct")

type ConversationPw struct {
	Name string `json:"name"`
	// profilePic     string
	ConversationId   int64  `json:"id"`
	LastMsgContent   string `json:"lastMsgContent"`
	LastMsgTimeStamp int64  `json:"lastMsgTimeStmp"`
	GroupId          int64  `json:"groupId"`
	UserId           int64  `json:"userId"`
	Photo            string `json:"photo"`
}
type CustomError struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type Picture struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Size uint32 `json:"size"`
}

type User struct {
	UserId    int64  `json:"id"`
	Name      string `json:"name"`
	UserPhoto string `json:"photo"`
}
type UserIdList struct {
	UserId []int64 `json:"userIdList"`
}
type Comment struct {
	Id             int64  `json:"id"`
	Content        string `json:"content"`
	MessageId      int64  `json:"messageId"`
	ConversationId int64  `json:"conversationId"`
	UserName       string `json:"userName"`
	UserId         int64  `json:"userId"`
}
type Message struct {
	Id            int64     `json:"id"`
	Sender        User      `json:"sender"`
	ConvId        int64     `json:"conversationId"`
	CommentList   []Comment `json:"commentList"`
	Content       string    `json:"content"`
	Timestamp     int64     `json:"timestamp"`
	PictureId     int64     `json:"pictureId"`
	RepliedId     int64     `json:"repliedId"`
	RepliedConvId int64     `json:"repliedConvId"`
}

type MessageReadStatus struct {
	HasBeenRead   bool    `json:"hasBeenRead"`
	ReadByUsers   []int64 `json:"readByUsers"`
	UnreadByUsers []int64 `json:"unreadByUsers"`
	MessageId     int64   `json:"messageId"`
}
type MessageInput struct {
	Sender        User   `json:"sender"`
	ConvId        int64  `json:"conversationId"`
	Content       string `json:"content"`
	RepliedId     int64  `json:"repliedId"`
	RepliedConvId int64  `json:"repliedConvId"`
}
type MsgForward struct {
	ConvId int64 `json:"forwardTo"`
}
type MessageId struct {
	Value int64 `json:"Id"`
}

type Conversation struct {
	Name     string    `json:"name"`
	Id       int64     `json:"id"`
	Users    []User    `json:"participants"`
	Messages []Message `json:"messages"`
}
type ConversationId struct {
	Value int64 `json:"convId"`
}

/*
	type Param struct{
		Key
		Value
	}
*/
type Group struct {
	GroupId int64   `json:"groupId"`
	UserId  []int64 `json:"userIdList"`
	Name    string  `json:"name"`
}
type GroupPw struct {
	Id   int64  `json:"id"`
	Pic  string `json:"picture"`
	Desc string `json:"desc"`
	Name string `json:"name"`
}
type Path struct {
	Path string `json:"path"`
}
