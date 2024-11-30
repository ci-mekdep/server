package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func SchoolTransferRoutes(api *gin.RouterGroup) {
	r := api.Group("/school-transfers")
	{
		r.GET("", SchoolTransferList)
		r.GET(":id", SchoolTransferDetail)
		r.POST("", SchoolTransferCreate)
		r.DELETE("", SchoolTransferDelete)
		r.GET("/inbox", SchoolTransferInbox)
		r.PUT("/inbox/:id/status", SchoolTransferInboxUpdate)
		r.GET("/inbox/:id", SchoolTransferInboxDetail)
	}
}

func schoolTransfersListQuery(ses *utils.Session, data models.SchoolTransferQueryDto) ([]models.SchoolTransferResponse, int, error) {
	if data.Limit == 0 {
		data.Limit = 12
	}
	if data.Offset == 0 {
		data.Offset = 0
	}
	data.SourceSchoolId = ses.GetSchoolId()
	res, err := app.SchoolTransfersList(ses, &data)
	return res.SchoolTransfersResponse, res.Total, err
}

func schoolTransfersInboxListQuery(ses *utils.Session, data models.SchoolTransferQueryDto) ([]models.SchoolTransferResponse, int, error) {
	if data.Limit == 0 {
		data.Limit = 12
	}
	if data.Offset == 0 {
		data.Offset = 0
	}
	data.TargetSchoolId = ses.GetSchoolId()
	res, err := app.SchoolTransfersList(ses, &data)
	return res.SchoolTransfersResponse, res.Total, err
}

func schoolTransfersAvailableCheck(ses *utils.Session, data models.SchoolTransferQueryDto) (bool, error) {
	_, t, err := schoolTransfersListQuery(ses, data)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

func schoolTransfersInboxAvailableCheck(ses *utils.Session, data models.SchoolTransferQueryDto) (bool, error) {
	_, t, err := schoolTransfersInboxListQuery(ses, data)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

func SchoolTransferList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSchoolTransfers, func(user *models.User) (err error) {
		r := models.SchoolTransferQueryDto{}
		if err = BindAny(c, &r); err != nil {
			return err
		}
		l, t, err := schoolTransfersListQuery(&ses, r)
		if err != nil {
			return err
		}
		if len(l) < 1 {
			l = []models.SchoolTransferResponse{}
		}
		Success(c, gin.H{
			"total":            t,
			"school_transfers": l,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SchoolTransferDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSchoolTransfers, func(u *models.User) (err error) {
		req := models.SchoolTransferQueryDto{}
		id := c.Param("id")
		if id == "" {
			return app.ErrNotfound
		}
		if ok, err := schoolTransfersAvailableCheck(&ses, models.SchoolTransferQueryDto{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		req.ID = &id
		response, err := app.SchoolTransferDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"school_transfer": response,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SchoolTransferCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminSchoolTransfers, func(user *models.User) (err error) {
		req := models.SchoolTransferCreateDto{}
		if err := BindAny(c, &req); err != nil {
			return err
		}
		req = models.SchoolTransferCreateDto{
			StudentId:         req.StudentId,
			TargetSchoolId:    req.TargetSchoolId,
			SourceClassroomId: req.SourceClassroomId,
			SourceSchoolId:    ses.GetSchoolId(),
			SentBy:            &ses.GetUser().ID,
		}
		req.Status = new(string)
		*req.Status = string(models.StatusWaiting)
		err = app_validation.ValidateSchoolTransfersCreate(&ses, req)
		if err != nil {
			return err
		}
		m, err := app.SchoolTransfersCreate(&ses, &req)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &m.ID,
			Subject:           models.LogSubjectSchoolTransfers,
			SubjectAction:     models.LogActionCreate,
			SubjectProperties: req,
		})
		Success(c, gin.H{
			"school_transfer": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SchoolTransferDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminSchoolTransfers, func(user *models.User) error {
		var ids []string = c.QueryArray("ids")
		if len(ids) == 0 {
			return app.ErrRequired.SetKey("ids")
		}

		if ok, err := schoolTransfersAvailableCheck(&ses, models.SchoolTransferQueryDto{IDs: ids}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		l, err := app.SchoolTransfersDelete(&ses, ids)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			Subject:           models.LogSubjectSchoolTransfers,
			SubjectAction:     models.LogActionDelete,
			SubjectProperties: ids,
		})
		Success(c, gin.H{
			"school_transfers": l,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SchoolTransferInbox(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSchoolTransfers, func(u *models.User) error {
		r := models.SchoolTransferQueryDto{}
		if err := BindAny(c, &r); err != nil {
			return err
		}
		res, total, err := schoolTransfersInboxListQuery(&ses, r)
		if err != nil {
			return err
		}
		if len(res) < 1 {
			res = []models.SchoolTransferResponse{}
		}
		Success(c, gin.H{
			"school_transfers": res,
			"total":            total,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SchoolTransferInboxUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminSchoolTransfers, func(user *models.User) (err error) {
		r := models.SchoolTransferCreateDto{}
		if err := BindAny(c, &r); err != nil {
			return err
		}
		r = models.SchoolTransferCreateDto{
			Status:            r.Status,
			TargetClassroomId: r.TargetClassroomId,
			ReceivedBy:        &ses.GetUser().ID,
		}
		id := c.Param("id")
		r.ID = &id
		if id == "" {
			return app.ErrRequired.SetKey("id")
		}
		if ok, err := schoolTransfersInboxAvailableCheck(&ses, models.SchoolTransferQueryDto{ID: r.ID}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		err = app_validation.ValidateSchoolTransfersInboxCreate(&ses, r)
		if err != nil {
			return err
		}

		m, err := app.SchoolTransfersUpdate(&ses, &r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &m.ID,
			Subject:           models.LogSubjectSchoolTransfers,
			SubjectAction:     models.LogActionUpdate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"school_transfer": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SchoolTransferInboxDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSchoolTransfers, func(u *models.User) (err error) {
		req := models.SchoolTransferQueryDto{}
		id := c.Param("id")
		if id == "" {
			return app.ErrNotfound
		}
		if ok, err := schoolTransfersInboxAvailableCheck(&ses, models.SchoolTransferQueryDto{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		req.ID = &id
		response, err := app.SchoolTransferDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"school_transfer": response,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
