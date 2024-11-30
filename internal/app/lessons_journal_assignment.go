package app

import (
	"errors"
	"math/rand"
	"mime/multipart"
	"os"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"github.com/mekdep/server/internal/utils"
)

func LessonUpdateAssignment(c *gin.Context, data *models.AssignmentRequest) (*models.Lesson, error) {
	// get existing if it is
	ml, err := store.Store().LessonsFindById(c, *data.LessonID)
	if err != nil {
		return nil, err
	}
	if ml.AssignmentFiles == nil {
		ml.AssignmentFiles = &[]string{}
	}

	// update existing, first delete
	paths := *ml.AssignmentFiles
	if data.FilesDelete != nil {
		if len(*data.FilesDelete) < 1 || (*data.FilesDelete)[0] == "" {
			*data.FilesDelete = paths
		}
		for _, v := range *data.FilesDelete {
			deleteFile(c, v, "assignments")
			k := slices.Index(paths, v)
			if k >= 0 {
				paths = slices.Delete(paths, k, k+1)
			}
		}
	}
	// then upload
	tmp, err := handleFilesUpload(c, "files", "assignments")
	if err != nil {
		return nil, err
	}
	paths = append(paths, tmp...)
	data.Files = &paths

	// update model
	ml.AssignmentTitle = data.Title
	ml.AssignmentContent = data.Content
	ml.AssignmentFiles = data.Files
	ml, err = store.Store().LessonsUpdate(c, ml)

	// deprecated
	ma := models.Assignment{}
	ma.FromRequest(data)
	ml.Assignment = &ma

	return &ml, nil
}

// TODO refactor: same function also has in api.go
func handleFilesUpload(c *gin.Context, key string, folder string) ([]string, error) {
	c.Request.ParseMultipartForm(100)
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}
	files := form.File[key]

	paths := []string{}
	for _, f := range files {
		path, err := handleFile(c, f, folder)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

// TODO refactor: same function also has in api.go
func deleteFile(c *gin.Context, path string, folder string) error {

	return nil
}

func extractFullUrl(url2 string) string {
	prefix := "https://mekdep.edu.tm/uploads/"
	if strings.HasPrefix(url2, prefix) {
		filePath := strings.TrimPrefix(url2, prefix)
		return filePath
	}
	return url2
}

func handleFile(c *gin.Context, handler *multipart.FileHeader, subFolder string) (string, error) {
	var err error
	uploadPath := "web/uploads/" + subFolder + "/"
	fParts := strings.Split(handler.Filename, ".")
	ext := "." + strings.ToLower(fParts[len(fParts)-1])
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".zip" && ext != ".rar" && ext != ".mp3" && ext != ".mp4" && ext != ".pdf" && ext != ".docx" && ext != ".doc" && ext != ".xlsx" {
		return "", errors.New("invalid extension: " + ext)
	}
	folderHash := RandStringBytes(16)
	uploadPath = uploadPath + folderHash + "/"
	fileName := handler.Filename

	err = c.SaveUploadedFile(handler, uploadPath+fileName)
	if err != nil {
		utils.LoggerDesc("HTTP error").Error(err)
		return "", err
	}
	err = os.Chmod(uploadPath, 0777)
	if err != nil {
		return "", err
	}
	err = os.Chmod(uploadPath+fileName, 0777)
	if err != nil {
		return "", err
	}

	return subFolder + "/" + folderHash + "/" + fileName, nil
}

// make utils
// TODO refactor: same function also has in api.go
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
