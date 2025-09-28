/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/mycompany8201046/myProject/service/api/model"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	GetConversation(convId int64) (*sql.Rows, error)
	GetConvName(convId int64, usrId int64) (string, error)
	GetConversations(usrId int64) (*sql.Rows, error)

	GetMessage(msgId int64, convId int64) *sql.Row
	HasMessageBeenRead(msgId int64, convId int64, userId int64) (bool, error)
	WhoHasReadMessage(msgId int64, convId int64, userId int64) (*sql.Rows, error)
	WhoHasNotReadMessage(msgId int64, convId int64, userId int64) (*sql.Rows, error)
	ReadMessage(msgId int64, convId int64, userId int64) error
	DeleteMessage(msgId int64, convId int64) (sql.Result, error)

	CreateMessage(message model.MessageInput, photoId int64, usrId int64, convId int64) (int64, error)
	ForwardMessage(OgMessageId int64, OgConvId int64, usrId int64, convId int64) (sql.Result, error)
	GetComments(convId int64, messageId int64) (*sql.Rows, error)
	CommentMessage(msgId int64, convId int64, usrId int64, commentContent string) (int64, error)
	PhotoMessage(picture model.Picture, messageId int64, msgInput model.Message, conversationId int64, userId int64) error

	GetFinalComment(msgId int64, convId int64) ([]model.Comment, error)
	RemoveComment(userId int64, convId int64, messageId int64) error
	CreateConversation(users []int64) (int64, error)

	CheckUsername(usrName string) (int64, error)
	InsertUser(newUsrName string) (sql.Result, error)
	SetMyUserName(newUsrName string, usrId int64) (sql.Result, error)

	GetConversationPhoto(convId int64, userId int64) (string, error)
	SetUserPhoto(pic model.Picture, usrId int64) (sql.Result, error)
	InsertPhoto(pic model.Picture) (int64, error)
	GetUserPhoto(userId int64) (string, error)
	GetGroupPhoto(groupId int64) (string, error)
	GetPhoto(photoId int64) (string, error)

	GetUsers() (*sql.Rows, error)
	GetUsersNotInConversation(userId int64) (*sql.Rows, error)
	GetUserName(userName int64) (string, error)
	GetUsersByConv(convId int64) (*sql.Rows, error)
	GetGroupInfo(groupId int64) *sql.Row
	GetUsersByGroup(groupId int64) (*sql.Rows, error)
	CreateGroup(g_name string, usrIds []int64) (sql.Result, error)
	AddGroup(newUsrIds []int64, grpId int64) (sql.Result, error)
	LeaveGroup(usrId int64, grpId int64) (sql.Result, error)
	SetGroupName(newName string, grpId int64) (sql.Result, error)
	SetGroupPhoto(pic model.Picture, grpId int64) (sql.Result, error)
	SetGroupDesc(newDesc string, grpId int64) (sql.Result, error)
	Ping() error
	SanitizeString(s string) (string, error)
	AddPreCommitHook(hook PreCommitHook)
	BeginTx() (*HookedTx, error)
	CreateSanitizeHook(validations map[string][]string) PreCommitHook
	GetNumOpenConn() int
	GetNumCurrentlyUsedConn() int
}
type PreCommitHook func(tx *sql.Tx) error

type appdbimpl struct {
	c              *sql.DB
	preCommitHooks []PreCommitHook
}

func (db *appdbimpl) GetNumCurrentlyUsedConn() int {
	return db.c.Stats().InUse
}

func (db *appdbimpl) GetNumOpenConn() int {

	return db.c.Stats().OpenConnections
}

func (db *appdbimpl) AddPreCommitHook(hook PreCommitHook) {
	db.preCommitHooks = append(db.preCommitHooks, hook)
}

func (db *appdbimpl) BeginTx() (*HookedTx, error) {
	tx, err := db.c.Begin()

	if err != nil {
		return nil, err
	}
	return &HookedTx{Tx: tx, hooks: db.preCommitHooks}, nil

}

// HookedTx wraps sql.Tx to add pre-commit hooks
type HookedTx struct {
	*sql.Tx
	hooks []PreCommitHook
}

// Commit runs all pre-commit hooks before committing
func (tx *HookedTx) Commit() error {
	// Run all pre-commit hooks
	for _, hook := range tx.hooks {
		if err := hook(tx.Tx); err != nil {
			// Rollback on hook failure
			err = tx.Rollback()
			if err != nil && !errors.Is(err, sql.ErrTxDone) {
				return err
			}
			return fmt.Errorf("pre-commit hook failed: %w", err)
		}
	}

	// Proceed with actual commit
	return tx.Tx.Commit()
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	_, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, err
	}
	// Check if table exists. If not, the database is empty, and we need to create the structure
	var tableName string
	var tx *sql.Tx
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='User';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		tx, err = db.Begin()
		if err != nil {
			return nil, fmt.Errorf("error starting transaction: %w", err)
		}
		// If the user table does not exist we must create every single table
		sqlStmt := [...]string{
			`CREATE TABLE "User" (
	    "userId"	INTEGER CHECK(userId > 0) UNIQUE,
	    "userName"	TEXT UNIQUE,
		"userPhoto" TEXT,
 	    PRIMARY KEY("userId" AUTOINCREMENT)
      );`,
			`CREATE TABLE "GroupTB" (
	    "groupId"	INTEGER NOT NULL CHECK(groupId > 0),
	    "convId"	INTEGER NOT NULL,
		"Name" TEXT,
	    "Description" TEXT,
		"photo"	TEXT,
	    PRIMARY KEY("groupId" AUTOINCREMENT),
	    CONSTRAINT "convID" FOREIGN KEY("convId") REFERENCES "Conversation"("conversationId")
      );`,
			`CREATE TABLE "Group_User" (
	    "groupId"	INTEGER NOT NULL,
	    "userId"	INTEGER NOT NULL,
		FOREIGN KEY("groupId") REFERENCES "GroupTB"("groupId"),
		FOREIGN KEY("userId") REFERENCES "User"("userId")
      );`,
			`CREATE TABLE "Conversation" (
	    "conversationId"	INTEGER NOT NULL CHECK(conversationId > 0) UNIQUE,
	    "lastMsgId"	INTEGER,
	    PRIMARY KEY("conversationId" AUTOINCREMENT) 
      );`,
			` CREATE TABLE "User_Chat" (
		"usrId1"	INTEGER NOT NULL,
		"usrId2"	INTEGER NOT NULL,
		"convId"	INTEGER NOT NULL UNIQUE,
		PRIMARY KEY("usrId2","usrId1"),
		FOREIGN KEY("convId") REFERENCES "Conversation"("conversationId") ON DELETE CASCADE
	);`,
			`CREATE TABLE "Photo" (
	    "id"	INTEGER UNIQUE,
	    "path"	TEXT NOT NULL UNIQUE,
		"size" INTEGER CHECK(size<20000000),
	    PRIMARY KEY("id" AUTOINCREMENT)
      );`,

			`CREATE TABLE "Message" (
	    "messageId"	INTEGER NOT NULL CHECK(messageId > 0),
	    "content"	TEXT NOT NULL,
	    "mtime"	INTEGER NOT NULL,
	    "usrSenderId"	INTEGER NOT NULL,
	    "convId"	INTEGER NOT NULL,
		"photoId" INTEGER,
		"repliedId" INTEGER,
		"repliedConvId" INTEGER,
		"forwarded" BOOL ,
	    PRIMARY KEY("messageId","convId"),
	    CONSTRAINT "convId" FOREIGN KEY("convId") REFERENCES "Conversation"("conversationId"),
	    CONSTRAINT "usrId" FOREIGN KEY("usrSenderId") REFERENCES "User"("userId"),
	    CONSTRAINT "photoId" FOREIGN KEY("photoId") REFERENCES "Photo"("id"),
		CONSTRAINT "repliedMessage" FOREIGN KEY("repliedId","repliedConvId") REFERENCES "Message"("messageId","convId") ON DELETE CASCADE
      )`,
			`CREATE TABLE "Comment" (
	"commentId"	INTEGER NOT NULL CHECK("commentId" > 0),
	"msgId"	INTEGER NOT NULL,
	"convId"	INTEGER NOT NULL,
	"userId"	INTEGER NOT NULL,
	"userName" TEXT ,
	"content"	TEXT NOT NULL,
	PRIMARY KEY("commentId","msgId","convId"),
	CONSTRAINT "msgId" FOREIGN KEY("msgId","convId") REFERENCES "Message"("messageId","convId") ON DELETE CASCADE,
	CONSTRAINT "userId" FOREIGN KEY("userId") REFERENCES "User"("userId")
	 )`,
			`CREATE TABLE "Conv_User" (
	    "convId"	INTEGER NOT NULL,
	    "usrId"	INTEGER NOT NULL,
	    CONSTRAINT "convId" FOREIGN KEY("convId") REFERENCES "Conversation"("conversationId") ON DELETE CASCADE ,
	    CONSTRAINT "usrId" FOREIGN KEY("usrId") REFERENCES "User"("userId") ON DELETE CASCADE
      )`, `CREATE TABLE "MessageReadStatus" (
	    "messageId" INTEGER NOT NULL,
	    "convId" INTEGER NOT NULL,
		"userId" INTEGER NOT NULL,
		"readTime" INTEGER NOT NULL,
		PRIMARY KEY("messageId", "convId", "userId"),
		FOREIGN KEY("messageId", "convId") REFERENCES "Message"("messageId", "convId"),
		FOREIGN KEY("userId") REFERENCES "User"("userId")
	  );`,
			`CREATE TRIGGER delete_comments_before_message 
  		 BEFORE DELETE ON Message
  		 FOR EACH ROW
		BEGIN
   			DELETE FROM Comment 
   		WHERE OLD.messageId = Comment.msgId AND OLD.convId = Comment.convId;
			END`,
			`CREATE TRIGGER update_conversation_after_message
  		 AFTER INSERT ON Message
		BEGIN
   			UPDATE Conversation 
		SET lastMsgId = NEW.messageId  	
		WHERE NEW.convId = conversationId;
			END`,
			`CREATE UNIQUE INDEX MessageReadStatus_messageId_convId_userId ON MessageReadStatus (messageId,convId,userId)`,
			`CREATE TRIGGER prevent_sender_read_status
				BEFORE INSERT ON "MessageReadStatus"
			FOR EACH ROW
			BEGIN
				SELECT CASE 
					WHEN (
						SELECT "usrSenderId" FROM "Message" 
						WHERE "messageId" = NEW."messageId" 
						AND "convId" = NEW."convId"
					) = NEW."userId" 
					THEN RAISE(IGNORE)
				END;
			END`,
			`CREATE TRIGGER update_conversation_after_message_delete
				AFTER DELETE ON Message
				BEGIN
				UPDATE Conversation 
				SET lastMsgId = (
					SELECT COALESCE(
						(SELECT messageId FROM Message 
							WHERE convId = OLD.convId 
							ORDER BY mtime DESC, messageId DESC 
							LIMIT 1), 
						0
					)
				)
				WHERE conversationId = OLD.convId;
				END`,
			`CREATE TRIGGER delete_group_empty AFTER DELETE ON Group_User 
			FOR EACH ROW  WHEN 
			(SELECT (SELECT COUNT(userId) FROM Group_User WHERE groupId = OLD.groupId) = 0)
			BEGIN 
			DELETE FROM GroupTB WHERE GroupTB.groupId = OLD.groupId;
			END;`,
			`CREATE TRIGGER comprehensive_conversation_cleanup
			AFTER DELETE ON Message
			FOR EACH ROW
			WHEN (SELECT COUNT(messageId) FROM Message WHERE convId = OLD.convId) = 0
			BEGIN
			    -- When the last message is deleted, clean up everything
			    DELETE FROM MessageReadStatus WHERE convId = OLD.convId;
			    DELETE FROM Comment WHERE convId = OLD.convId;
			    -- Messages already deleted (what triggered this)
			    DELETE FROM Group_User WHERE groupId IN (
			        SELECT groupId FROM GroupTB WHERE convId = OLD.convId
			    );
			    DELETE FROM GroupTB WHERE convId = OLD.convId;
			    DELETE FROM User_Chat WHERE convId = OLD.convId;
			    DELETE FROM Conv_User WHERE convId = OLD.convId;
			    DELETE FROM Conversation WHERE conversationId = OLD.convId;
			END;`,
			`INSERT INTO User (userName) VALUES("Alice");`,
			`INSERT INTO User (userName) VALUES("Bob");`,
			`INSERT INTO User (userName) VALUES("Charlie");`,
			`INSERT INTO User (userName) VALUES("Alex");`,
			`INSERT INTO User (userName) VALUES("Bobby");`,
			`INSERT INTO User (userName) VALUES("Charles");`,
			`INSERT INTO User (userName) VALUES("Linda");`,
			`INSERT INTO User (userName) VALUES("Versace");`,
			`INSERT INTO User (userName) VALUES("Gucci");`,
			`INSERT INTO Conversation (lastMsgId) VALUES(0);`,
			`INSERT INTO Conversation (lastMsgId) VALUES(0);`,
			`INSERT INTO Conv_User (convId,usrId) VALUES(1,1);`,
			`INSERT INTO Conv_User (convId,usrId) VALUES(1,2);`,
			`INSERT INTO Conv_User (convId,usrId) VALUES(2,1);`,
			`INSERT INTO Conv_User (convId,usrId) VALUES(2,2);`,
			`INSERT INTO Conv_User (convId,usrId) VALUES(2,3);`,
			`INSERT INTO GroupTB (convId,Description,Name) VALUES(2,"Ciao,gruppo cybersec","Gruppo Cybersec");`,
			`INSERT INTO Group_User (groupId,userId) VALUES(1,1);`,
			`INSERT INTO Group_User (groupId,userId) VALUES(1,2);`,
			`INSERT INTO Group_User (groupId,userId) VALUES(1,3);`,
			`INSERT INTO Message (messageId,content,mtime,usrSenderId,convId) VALUES(1,"Ciao Bob!!!",1737478490,1,1)`,
			`INSERT INTO Message (messageId,content,mtime,usrSenderId,convId) VALUES(2,"Hey Ali 💓.",1737478500,2,1)`,
			`INSERT INTO Message (messageId,content,mtime,usrSenderId,convId) VALUES(3,"Domani dove vuoi andare?",1737478510,2,1)`,
			`INSERT INTO Message (messageId,content,mtime,usrSenderId,convId) VALUES(4,"Non so...",unixepoch(),1,1)`,
			`INSERT INTO Message (messageId,content,mtime,usrSenderId,convId) VALUES(1,"Benvenuti🥳🎉🎊 nel gruppo di cybersec💻!!!",1737478600,3,2)`,
			`INSERT INTO Comment (commentId,msgId,convId,userId,userName,content) VALUES(1,2,1,1,"Alice","😍")`,
			`INSERT INTO Photo (path,size) VALUES("./images/defaultPP.png",2183)`,
		}
		for _, query := range sqlStmt {
			_, err = tx.Exec(query)
			if err != nil {
				err = tx.Rollback()
				if err != nil {
					return nil, errors.New("error rolling  back the changes:" + err.Error())
				}
				return nil, errors.New("error creating database structure:" + err.Error())
			}
		}
		err = tx.Commit()
		if err != nil {
			return nil, errors.New("error committing transaction:" + err.Error())
		}
	}
	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
