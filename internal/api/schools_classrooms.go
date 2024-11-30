package api

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func ClassroomRoutes(api *gin.RouterGroup) {
	cRoutes := api.Group("/classrooms")
	{
		cRoutes.GET("", ClassroomList)
		cRoutes.GET("/values", ClassroomListValues)
		cRoutes.GET(":id", ClassroomDetail)
		cRoutes.PUT(":id", ClassroomUpdate)
		cRoutes.PATCH(":id/relations", ClassroomPatchRelations)
		cRoutes.POST("", ClassroomCreate)
		cRoutes.DELETE("", ClassroomDelete)
	}
}

func classroomListQuery(ses *utils.Session, data models.ClassroomFilterRequest, IsValue bool) ([]*models.ClassroomResponse, []models.ClassroomValueResponse, int, error) {
	if data.SchoolIds == nil {
		data.SchoolIds = &[]string{}
	}
	if len(*data.SchoolIds) > 0 {
		sids := *data.SchoolIds
		data.SchoolIds = &[]string{}
		for _, sid := range sids {
			if slices.Contains(ses.GetSchoolIds(), sid) {
				*data.SchoolIds = append(*data.SchoolIds, sid)
			}
		}
	}
	if len(*data.SchoolIds) < 1 {
		*data.SchoolIds = ses.GetSchoolIds()
	}
	if *ses.GetRole() == models.RoleTeacher && ses.GetUser().TeacherClassroom != nil {
		data.ID = &ses.GetUser().TeacherClassroom.ID
	}
	if IsValue {
		res, total, err := app.ClassroomListValues(ses, data)
		return nil, res, total, err
	} else {
		res, total, err := app.ClassroomsList(ses, data)
		return res, nil, total, err
	}
}

func classroomAvailableCheck(ses *utils.Session, data models.ClassroomFilterRequest) (bool, error) {
	_, _, t, err := classroomListQuery(ses, data, false)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

func ClassroomListValues(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminClassrooms, func(u *models.User) (err error) {
		r := models.ClassroomFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		_, classrooms, total, err := classroomListQuery(&ses, r, true)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"total":      total,
			"classrooms": classrooms,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ClassroomList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminClassrooms, func(user *models.User) error {
		r := models.ClassroomFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}

		l, _, total, err := classroomListQuery(&ses, r, false)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"total":      total,
			"classrooms": l,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ClassroomDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminClassrooms, func(user *models.User) error {
		id := c.Param("id")
		if ok, err := classroomAvailableCheck(&ses, models.ClassroomFilterRequest{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		m, err := app.ClassroomsDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"classroom": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ClassroomPatchRelations(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminClassrooms, func(u *models.User) error {
		r := models.ClassroomRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		id := c.Param("id")
		if ok, err := classroomAvailableCheck(&ses, models.ClassroomFilterRequest{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		r.ID = &id
		err := app.ClassroomsUpdateRelations(&ses, id, r)
		if err != nil {
			Success(c, gin.H{})
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            ses.GetUser().ID,
			SubjectId:         r.ID,
			Subject:           models.LogSubjectClassrooms,
			SubjectAction:     models.LogActionUpdate,
			SubjectProperties: r,
		})
		return err
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ClassroomUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminClassrooms, func(user *models.User) (err error) {
		r := models.ClassroomRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		id := c.Param("id")
		r.ID = &id
		if id == "" {
			return app.ErrRequired.SetKey("id")
		}
		err = app_validation.ValidateClassroomsCreate(&ses, r)
		if err != nil {
			return err
		}

		if ok, err := classroomAvailableCheck(&ses, models.ClassroomFilterRequest{ID: r.ID}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		m, err := app.ClassroomsUpdate(&ses, &r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &m.ID,
			Subject:           models.LogSubjectClassrooms,
			SubjectAction:     models.LogActionUpdate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"classroom": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ClassroomCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminClassrooms, func(user *models.User) (err error) {
		r := models.ClassroomRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}
		err = app_validation.ValidateClassroomsCreate(&ses, r)
		if err != nil {
			return err
		}
		m, err := app.ClassroomsCreate(&ses, &r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &m.ID,
			Subject:           models.LogSubjectClassrooms,
			SubjectAction:     models.LogActionCreate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"classroom": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ClassroomDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminClassrooms, func(user *models.User) error {
		var ids []string = c.QueryArray("ids")
		if len(ids) == 0 {
			return app.ErrRequired.SetKey("id")
		}

		if ok, err := classroomAvailableCheck(&ses, models.ClassroomFilterRequest{Ids: &ids}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		l, err := app.ClassroomsDelete(&ses, ids)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"classrooms": l,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}
