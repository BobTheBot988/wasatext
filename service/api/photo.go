package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // register JPEG decoder
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"gitlab.com/mycompany8201046/myProject/service/api/model"
	"gitlab.com/mycompany8201046/myProject/service/api/reqcontext"
)

func (rt *_router) SendImageBack(w http.ResponseWriter, path string) error {

	rt.baseLogger.Info("Sending Image Back")
	imageData, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Check file size (20MB limit)
	if len(imageData) > 20*1024*1024 {
		return errors.New("image is too big")
	}

	// Set proper headers
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(imageData)))

	// Write image data to response
	_, err = w.Write(imageData)

	if err != nil {
		return err
	}

	return nil
}

// saveImageOnDisk saves the provided image to disk as PNG
// Returns the generated filename or empty string on error
func (rt *_router) SaveImageOnDisk(newPhoto image.Image) (string, string, error) {
	rt.baseLogger.Info("Saving Image on Disk")
	// Generate UUID for unique filename
	uuid, err := uuid.NewV4()
	rt.baseLogger.Infof("Creating uuid:%s", uuid.String())
	if err != nil {
		return "", "", err
	}

	// Get executable directory to create consistent paths
	const UPLOAD_DIR = "./images/"

	rt.baseLogger.Infof("Upload Dir :%s", UPLOAD_DIR)
	if err = os.MkdirAll(UPLOAD_DIR, os.ModePerm); err != nil {
		return "", "", err
	}

	// Create filename and full path
	id := uuid.String()
	ext := ".png"
	imgName := id + ext
	imgPath := filepath.Join(UPLOAD_DIR, imgName)

	// Create file

	file, err := os.Create(imgPath)
	if err != nil {
		return "", "", err
	}
	rt.baseLogger.Infof("File Created:%s", file.Name())
	defer func() {
		if err = file.Close(); err != nil {
			rt.baseLogger.Error(err)
		}
	}()

	// Encode and save the image as PNG
	if err = png.Encode(file, newPhoto); err != nil {
		return "", "", err
	}

	// Return consistent paths
	dataPath := filepath.Join(UPLOAD_DIR, imgName)
	relativePath := filepath.Join("images", imgName)

	return dataPath, relativePath, nil
}

func (rt *_router) uploadPhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext, choice uint8) {
	if !rt.isAuthed(w, r, ps, ctx) {
		return
	}
	var err error
	var userId, grpId, conversationId int64
	// Parse multipart form with 20MB max memory
	rt.baseLogger.Println("Photo handling service")
	if err = r.ParseMultipartForm(30 << 20); err != nil {
		rt.internalError(500, err, r, w)
		return
	}

	rt.baseLogger.Infof("Making a choice:%d\n", choice)
	switch choice {
	case 0:
		userId, err = strconv.ParseInt(ps.ByName("userId"), 10, 64)
		if err != nil {
			rt.internalError(400, model.AddError(model.ErrMalformedUserId, err), r, w)
			return
		}

	case 1:
		grpId, err = strconv.ParseInt(ps.ByName("groupId"), 10, 64)
		if err != nil {
			rt.internalError(400, model.AddError(model.ErrMalformedGroupId, err), r, w)
			return
		}

	case 2:
		rt.baseLogger.Println("Message Photo chosen.")
		userId, err = strconv.ParseInt(ps.ByName("userId"), 10, 64)
		if err != nil {
			rt.internalError(500, err, r, w)
			return
		}

		conversationId, err = strconv.ParseInt(ps.ByName("conversationId"), 10, 64)

		if err != nil {
			rt.internalError(500, err, r, w)
			rt.internalError(400, model.AddError(model.ErrMalformedConvId, err), r, w)
			return
		}

	default:
		rt.internalError(500, errors.New("something went wrong PHOTO"), r, w)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		rt.internalError(500, model.AddErrorString("Error on image decoding", err.Error()), r, w)
		return
	}
	filePath, relativeFilePath, err := rt.SaveImageOnDisk(img)
	if err != nil {
		rt.internalError(500, model.AddErrorString("Error on image creation file", err.Error()), r, w)
		return
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		rt.internalError(500, model.AddErrorString("Failed to get file info", err.Error()), r, w)

		return
	}

	// Create Picture record
	picture := model.Picture{
		Name: header.Filename,
		Path: relativeFilePath, // Store relative path for serving
		Size: uint32(fileInfo.Size()),
	}

	if userId != 0 && choice != 2 {
		_, err = rt.db.SetUserPhoto(picture, userId)
	} else if grpId != 0 && choice != 2 {
		_, err = rt.db.SetGroupPhoto(picture, grpId)
	}
	if choice == 2 {
		picture.Id, err = rt.db.InsertPhoto(picture)
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			rt.internalError(500, err, r, w)
			return
		}
		var msgInput model.Message
		// Read individual form fields
		content := r.FormValue("content")
		if len(content) > 1000 {
			e := errors.New("the message cannot be bigger than a 1000 chars")
			rt.internalError(400, e, r, w)
			return
		}
		msgInput.Content = content

		// Read sender data if it was sent as JSON string
		senderData := r.FormValue("sender")
		if senderData != "" {
			err = json.Unmarshal([]byte(senderData), &msgInput.Sender)
			if err != nil {
				rt.baseLogger.Error("Error parsing sender data:", err)
				rt.internalError(400, err, r, w)
				return
			}
			// Use sender data as needed
		}

		// Read other fields
		repliedIdStr := r.FormValue("repliedId")
		rt.baseLogger.Infof("Form:", r.Form)
		var repliedId int64

		if repliedIdStr != "" && repliedIdStr != "null" {
			repliedId, err = strconv.ParseInt(repliedIdStr, 10, 64)
			if err != nil {
				rt.internalError(400, err, r, w)
				return
			}
		}

		msgInput.RepliedId = repliedId
		repliedConvIdStr := r.FormValue("repliedConvId")

		var repliedConvId int64
		if repliedConvIdStr != "" && repliedConvIdStr != "null" {
			repliedConvId, err = strconv.ParseInt(repliedConvIdStr, 10, 64)
			if err != nil {
				rt.internalError(400, err, r, w)
				return
			}
		}
		msgInput.RepliedConvId = repliedConvId
		rt.baseLogger.Println("Creating a photo message with photoId:", picture.Id)
		rt.baseLogger.Info("MessageInput:", msgInput)
		err = rt.db.PhotoMessage(picture, 0, msgInput, conversationId, userId)
	}

	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = rt.SendImageBack(w, picture.Path)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
}
func (rt *_router) getPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if !rt.isAuthed(w, r, ps, ctx) {
		return
	}

	photoId, err := strconv.ParseInt(ps.ByName("photoId"), 10, 64)
	if err != nil {
		rt.internalError(400, model.AddError(model.ErrMalformedPhotoId, err), r, w)
		return
	}

	path, err := rt.db.GetPhoto(photoId)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}

	err = rt.SendImageBack(w, path)
	if err != nil {
		rt.internalError(500, err, r, w)
		return
	}
}
