package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func BookRoutes(api *gin.RouterGroup) {
	bookRoutes := api.Group("/books")
	{
		bookRoutes.GET("", BookList)
		bookRoutes.PUT(":id", BookUpdate)
		bookRoutes.POST("", BookCreate)
		bookRoutes.DELETE("", BookDelete)
		bookRoutes.GET(":id", BookDetail)
		bookRoutes.GET("/authors", BookGetAuthors)
	}
}

func booksListQuery(ses *utils.Session, data models.BookFilterRequest) ([]*models.BookResponse, int, error) {
	res, total, err := app.BookList(ses, data)
	return res, total, err
}

func booksAvailableCheck(ses *utils.Session, data models.BookFilterRequest) (bool, error) {
	_, t, err := booksListQuery(ses, data)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

func BookList(c *gin.Context) {
	ses := utils.InitSession(c)
	r := models.BookFilterRequest{}
	if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
		app.NewAppError(errMsg, errKey, "")
		return
	}

	books, total, err := booksListQuery(&ses, r)
	if err != nil {
		handleError(c, err)
		return
	}

	Success(c, gin.H{
		"books": books,
		"total": total,
	})
}

func BookDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminBooks, func(u *models.User) error {
		id := c.Param("id")
		if ok, err := booksAvailableCheck(&ses, models.BookFilterRequest{ID: id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		m, err := app.BookDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"book": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func BookGetAuthors(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(u *models.User) error {
		authors, err := app.BookGetAuthors(&ses)
		if err != nil {
			return err
		}

		Success(c, gin.H{
			"authors": authors,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func BookUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminBooks, func(user *models.User) (err error) {
		r := models.BookRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		id := c.Param("id")
		r.ID = id

		if id == "" {
			return app.ErrRequired.SetKey("id")
		}
		err = app_validation.ValidateBooksCreate(&ses, r)
		if err != nil {
			return err
		}
		if ok, err := booksAvailableCheck(&ses, models.BookFilterRequest{ID: r.ID}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}
		book, bookModel, err := app.BookUpdate(&ses, r)
		if err != nil {
			return err
		}
		// upload file
		file, fileSize, err := handleFileUploadForBook(c, "file", "books", true, bookModel)
		if err != nil {
			return app.NewAppError(err.Error(), "file", "")
		}

		// if user upload file_preview
		previewImageByHand, _, err := handleFileUploadForBook(c, "file_preview", "books", true, bookModel)
		if err != nil {
			return app.NewAppError(err.Error(), "file_preview", "")
		}
		if previewImageByHand != "" {
			bookModel.FilePreview = &previewImageByHand
		}

		if file != "" {
			bookModel.File = &file
			fileSizeInt := int(fileSize)
			bookModel.FileSize = &fileSizeInt
		}

		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &r.ID,
			Subject:           models.LogSubjectBook,
			SubjectAction:     models.LogActionUpdate,
			SubjectProperties: r,
		})

		Success(c, gin.H{
			"book": book,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func BookCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminBooks, func(user *models.User) (err error) {
		r := models.BookRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}
		err = app_validation.ValidateBooksCreate(&ses, r)
		if err != nil {
			return err
		}
		book, bookModel, err := app.BookCreate(&ses, r)
		if err != nil {
			return err
		}
		// upload file
		file, fileSize, err := handleFileUploadForBook(c, "file", "books", true, bookModel)
		if err != nil {
			return app.NewAppError(err.Error(), "file", "")
		}
		// if user upload file_preview
		previewImageByHand, _, err := handleFileUploadForBook(c, "file_preview", "books", true, bookModel)
		if err != nil {
			return app.NewAppError(err.Error(), "file_preview", "")
		}
		if previewImageByHand != "" {
			bookModel.FilePreview = &previewImageByHand
		}
		if file != "" {
			bookModel.File = &file
			fileSizeInt := int(fileSize)
			bookModel.FileSize = &fileSizeInt
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &book.ID,
			Subject:           models.LogSubjectBook,
			SubjectAction:     models.LogActionCreate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"book": book,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func BookDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminBooks, func(user *models.User) error {
		var ids []string = c.QueryArray("ids")
		if len(ids) == 0 {
			return app.ErrRequired.SetKey("ids")
		}

		if ok, err := booksAvailableCheck(&ses, models.BookFilterRequest{IDs: &ids}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		books, err := app.BookDelete(&ses, ids)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			Subject:           models.LogSubjectBook,
			SubjectAction:     models.LogActionDelete,
			SubjectProperties: ids,
		})
		Success(c, gin.H{
			"books": books,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
