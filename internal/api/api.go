package api

import (
	"context"
	"encoding/json"
	"errors"
	"image/jpeg"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gen2brain/go-fitz"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4"
	"github.com/mekdep/server/config"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"github.com/mekdep/server/internal/utils"
)

func Success(c *gin.Context, data gin.H) {
	c.JSON(http.StatusOK, SuccessResponseObject(data))
}

func handleError(c *gin.Context, err error) {
	if errA, ok := err.(*app.AppError); ok {
		if errA == app.ErrUnauthorized {
			if utils.GetLoggerDesc() == "" {
				utils.Logger.Error(err)
			}
			c.JSON(http.StatusUnauthorized, ErrorResponseObject(errA, nil, nil))
		} else if errA == app.ErrForbidden {
			if utils.GetLoggerDesc() == "" {
				// utils.Logger.Error(err)
			}
			c.JSON(http.StatusForbidden, ErrorResponseObject(errA, nil, nil))
		} else if errA == app.ErrNotfound || errA == pgx.ErrNoRows {
			if utils.GetLoggerDesc() == "" {
				utils.Logger.Error(err)
			}
			c.JSON(http.StatusNotFound, ErrorResponseObject(errA, nil, nil))
		} else if errA == app.ErrNotPaid {
			c.JSON(http.StatusPaymentRequired, ErrorResponseObject(errA, nil, nil))
		} else {
			c.JSON(http.StatusBadRequest, ErrorResponseObject(errA, nil, nil))
		}
	} else if errA, ok := err.(*app_validation.AppError); ok {
		if errA == app_validation.ErrUnauthorized {
			if utils.GetLoggerDesc() == "" {
				utils.Logger.Error(err)
			}
			c.JSON(http.StatusUnauthorized, ErrorResponseObject(nil, errA, nil))
		} else if errA == app_validation.ErrForbidden {
			if utils.GetLoggerDesc() == "" {
				// utils.Logger.Error(err)
			}
			c.JSON(http.StatusForbidden, ErrorResponseObject(nil, errA, nil))
		} else if errA == app_validation.ErrNotfound || errA == pgx.ErrNoRows {
			if utils.GetLoggerDesc() == "" {
				utils.Logger.Error(err)
			}
			c.JSON(http.StatusNotFound, ErrorResponseObject(nil, errA, nil))
		} else if errA == app_validation.ErrNotPaid {
			c.JSON(http.StatusPaymentRequired, ErrorResponseObject(nil, errA, nil))
		} else {
			c.JSON(http.StatusBadRequest, ErrorResponseObject(nil, errA, nil))
		}
	} else if errA, ok := err.(app_validation.AppErrorCollection); ok {
		c.JSON(http.StatusBadRequest, ErrorResponseObject(nil, nil, &errA))
	} else {
		if utils.GetLoggerDesc() == "" {
			utils.Logger.Error(err)
		}
		if config.Conf.AppEnv == config.APP_ENV_DEV {
			c.JSON(http.StatusInternalServerError, ErrorResponseObject(app.NewAppError(err.Error(), "", ""), nil, nil))
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponseObject(app.NewAppError("something went wrong, please contact admin.", "", ""), nil, nil))
		}
	}
}
func userLog(data models.UserLog) error {
	go func() {
		// delete security keys
		secKeys := []string{"password", "otp", "device_token", "token"}
		prStr, _ := json.Marshal(data.SubjectProperties)
		pr := map[string]interface{}{}
		_ = json.Unmarshal(prStr, &pr)
		for _, k := range secKeys {
			if _, ok := pr[k]; ok {
				delete(pr, k)
			}
		}
		data.SubjectProperties = pr

		_, err := store.Store().UserLogsCreate(context.Background(), data)
		if err != nil {
			utils.LoggerDesc("in user log worker").Error(err)
		}
	}()
	return nil
}

func ParseDate(s string) (time.Time, error) {
	tt, err := time.ParseInLocation(time.DateOnly, s, config.RequestLocation)
	if err != nil {
		return time.Time{}, err
	}
	tt = tt.UTC().Local()
	return tt, nil
}

func ParseDateUTC(s string) (time.Time, error) {
	tt, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return time.Time{}, err
	}
	tt = tt.UTC().Local()
	return tt, nil
}

func BindAndValidate(c *gin.Context, r interface{}) (errMessage string, errKey string) {
	if err := c.Bind(r); err != nil {
		errMessage = err.Error()
		return
	}

	v := validator.New()
	if err := v.Struct(r); err != nil {
		err := err.(validator.ValidationErrors)[0]
		errMessage = err.Tag()
		errKey = (err.Field())
		return
	}
	return
}

func BindAny(c *gin.Context, r interface{}) error {
	if err := c.ShouldBind(r); err != nil {
		return app_validation.ErrInvalid.SetKey("json").SetComment(err.Error())
	}
	return nil
}

func SuccessResponseObject(data gin.H) gin.H {
	data["success"] = true
	return data
}

// TODO: remove app.AppError
func ErrorResponseObject(err *app.AppError, errr *app_validation.AppError, errs *app_validation.AppErrorCollection) gin.H {
	res := gin.H{
		"success": false,
		"data":    nil,
		"error":   nil,
		"errors":  nil,
	}
	if errs != nil && errs.HasError() {
		// TODO: remove app.AppError
		err = app.NewAppError(
			errs.Errors[0].Code(),
			errs.Errors[0].Key(),
			errs.Errors[0].Comment(),
		)
	}
	if errr != nil {
		err = app.NewAppError(
			errr.Code(),
			errr.Key(),
			errr.Comment(),
		)
	}
	if errs == nil || !errs.HasError() {
		errs = &app_validation.AppErrorCollection{
			Errors: []app_validation.AppError{
				*app_validation.NewAppError(
					err.Code(),
					err.Key(),
					err.Comment(),
				),
			},
		}
	}
	res["error"] = gin.H{
		"code":    err.Code(),
		"key":     err.Key(),
		"comment": err.Comment(),
	}
	res["errors"] = []gin.H{}
	for _, v := range errs.Errors {
		res["errors"] = append(res["errors"].([]gin.H), gin.H{
			"code":    v.Code(),
			"key":     v.Key(),
			"comment": v.Comment(),
		})
	}
	return res
}

func GetUser(c *gin.Context) *models.User {
	m, _ := c.Get("user")
	if m == nil {
		return nil
	}
	return m.(*models.User)
}

func IsUserTeacher(c *gin.Context) bool {
	return HasRole(c, models.RoleTeacher)
}

func GetRole(c *gin.Context) *models.Role {
	u := GetUser(c)
	if u != nil {
		for _, v := range u.Schools {
			return &v.RoleCode
		}
	}
	return nil
}

func HasRole(c *gin.Context, role models.Role) bool {
	u := GetUser(c)
	if u != nil {
		for _, v := range u.Schools {
			if v.RoleCode == role {
				return true
			}
		}
	}
	return false
}

func handleFilesUpload(c *gin.Context, key string, folder string) ([]string, error) {
	c.Request.ParseMultipartForm(100)
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}
	files := form.File[key]

	paths := []string{}
	for _, f := range files {
		path, _, err := handleFile(c, f, folder)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func deleteFile(c *gin.Context, path string, folder string) error {

	return nil
}

func handleFile(c *gin.Context, handler *multipart.FileHeader, subFolder string) (string, int64, error) {
	var err error
	uploadPath := "web/uploads/" + subFolder + "/"
	fParts := strings.Split(handler.Filename, ".")
	ext := "." + strings.ToLower(fParts[len(fParts)-1])
	validExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".zip": true, ".rar": true,
		".mp3": true, ".mp4": true, ".pdf": true, ".docx": true, ".doc": true, ".xlsx": true,
	}
	if !validExts[ext] {
		return "", 0, errors.New("invalid extension: " + ext)
	}
	folderHash := RandStringBytes(16)
	uploadPath = uploadPath + folderHash + "/"
	fileName := handler.Filename
	fileSize := handler.Size / 1024
	err = c.SaveUploadedFile(handler, uploadPath+fileName)
	if err != nil {
		utils.LoggerDesc("HTTP error").Error(err)
		return "", 0, err
	}

	err = os.Chmod(uploadPath, 0777)
	if err != nil {
		return "", 0, err
	}
	err = os.Chmod(uploadPath+fileName, 0777)
	if err != nil {
		return "", 0, err
	}

	return subFolder + "/" + folderHash + "/" + fileName, fileSize, nil
}

func storeBook(book *models.Book) error {
	_, err := store.Store().BookUpdate(context.Background(), book)
	return err
}

func handleFileForBook(c *gin.Context, handler *multipart.FileHeader, subFolder string, m *models.Book, saver func(model *models.Book) error) (string, int64, error) {
	var err error
	uploadDir := "web/uploads/" + subFolder + "/"
	fParts := strings.Split(handler.Filename, ".")
	ext := "." + strings.ToLower(fParts[len(fParts)-1])
	validExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".pdf": true,
	}
	if !validExts[ext] {
		return "", 0, errors.New("invalid extension: " + ext)
	}
	folderHash := RandStringBytes(16)
	uploadDir = uploadDir + folderHash + "/"
	fileName := handler.Filename
	fileSize := handler.Size / 1024
	err = c.SaveUploadedFile(handler, uploadDir+fileName)
	os.Chmod(uploadDir, 0755)
	os.Chmod(uploadDir+fileName, 0755)
	if err != nil {
		utils.LoggerDesc("HTTP error").Error(err)
		return "", 0, err
	}
	var previewImage string
	var pageCount int
	filePath := strings.Replace(uploadDir, "web/uploads/", "", 1) + fileName
	if ext == ".pdf" {
		func() {
			previewImage, pageCount, err = convertPdfToCoverImage(uploadDir, fileName)
			if err != nil {
				utils.LoggerDesc("PDF convert error").Error(err)
				return
			}
			m.Pages = &pageCount
			m.FilePreview = &previewImage
			m.File = &filePath
			err := saver(m)
			if err != nil {
				utils.LoggerDesc("PDF convert error").Error(err)
				return
			}
		}()
	}
	return filePath, fileSize, nil
}

func convertPdfToCoverImage(uploadPath, fileName string) (string, int, error) {
	doc, err := fitz.New(uploadPath + fileName)
	if err != nil {
		return "", 0, err
	}
	defer doc.Close()
	pageCount := doc.NumPage()
	img, err := doc.Image(0)
	previewImage := uploadPath + "hover.jpg"
	f, err := os.Create(previewImage)
	os.Chmod(previewImage, 0755)
	if err != nil {
		return "", 0, err
	}
	err = jpeg.Encode(f, img, &jpeg.Options{jpeg.DefaultQuality})
	if err != nil {
		return "", 0, err
	}
	f.Close()
	return strings.Replace(previewImage, "web/uploads/", "", 1), pageCount, nil
}

func handleFileUpload(c *gin.Context, key string, folder string, public bool) (string, int64, error) {
	c.Request.ParseMultipartForm(10 << 20)
	_, handler, err := c.Request.FormFile(key)
	if err != nil {
		return "", 0, nil
	}
	return handleFile(c, handler, folder)
}

func handleFileUploadForBook(c *gin.Context, key string, folder string, public bool, m *models.Book) (string, int64, error) {
	c.Request.ParseMultipartForm(10 << 20)
	_, handler, err := c.Request.FormFile(key)
	if err != nil {
		return "", 0, nil
	}
	return handleFileForBook(c, handler, folder, m, storeBook)
}

// make utils
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
