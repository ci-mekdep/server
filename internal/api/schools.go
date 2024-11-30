package api

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
)

func SchoolRoutes(api *gin.RouterGroup) {
	routes := api.Group("/schools")
	{
		routes.GET("/regions", SchoolRegions)
		routes.GET("/regions/values", SchoolRegionsListValues)
		routes.GET("", SchoolList)
		routes.GET("/values", SchoolListValues)
		routes.GET(":id", SchoolDetail)
		routes.PUT(":id", SchoolUpdate)
		routes.POST("", SchoolCreate)
		routes.DELETE("", SchoolDelete)
	}
}

func schoolListQuery(ses *utils.Session, data models.SchoolFilterRequest, IsValue bool) ([]*models.SchoolResponse, []models.SchoolValueResponse, int, error) {
	data.IsParent = new(bool)
	*data.IsParent = false
	if data.ParentUid != nil && *data.ParentUid != "" {
		// set data.parentIds validate, add exceptions
		p, _ := store.Store().SchoolsFindById(ses.Context(), *data.ParentUid)
		if p == nil {
			return nil, nil, 0, app.ErrNotfound
		}
		if p.Code != nil && *p.Code == "ag" {
			// TODO: Add parentIds of Ashgabat etraps
			regionIds := []string{"brk", "bgt", "bzm", "kpt"}
			parentSchools, err := store.Store().SchoolsFindByCode(ses.Context(), regionIds)
			if err != nil {
				return nil, nil, 0, app.ErrNotfound
			}
			var parentIds []string
			for _, school := range parentSchools {
				parentIds = append(parentIds, school.ID)
			}
			data.ParentUids = &parentIds
			data.ParentUid = nil
		}
	}
	data.Uids = new([]string)
	if len(*data.Uids) > 0 {
		sids := *data.Uids
		data.Uids = &[]string{}
		for _, sid := range sids {
			if slices.Contains(ses.SchoolsWithCentersIds(), sid) {
				*data.Uids = append(*data.Uids, sid)
			}
		}
	}
	if len(*data.Uids) < 1 {
		*data.Uids = ses.SchoolsWithCentersIds()
	}
	if IsValue {
		res, total, err := app.SchoolsListValues(ses, data)
		return nil, res, total, err
	} else {
		res, total, err := app.Ap().Schools(ses, data)
		return res, nil, total, err
	}
}

func schoolAvailableCheck(ses *utils.Session, data models.SchoolFilterRequest) (bool, error) {
	_, _, t, err := schoolListQuery(ses, data, false)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

func SchoolListValues(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSchools, func(u *models.User) (err error) {
		r := models.SchoolFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}

		_, schools, total, err := schoolListQuery(&ses, r, true)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"total":   total,
			"schools": schools,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SchoolRegions(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSchools, func(u *models.User) error {
		r := models.SchoolFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}

		l, _, total, err := app.Ap().PublicSchoolRegions(r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"total":   total,
			"regions": l,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SchoolRegionsListValues(c *gin.Context) {
	r := models.SchoolFilterRequest{}
	if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
		handleError(c, app.NewAppError(errMsg, errKey, ""))
		return
	}
	_, l, total, err := app.Ap().PublicSchoolRegions(r)
	if err != nil {
		handleError(c, err)
		return
	}
	Success(c, gin.H{
		"total":   total,
		"regions": l,
	})
}

func SchoolList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSchools, func(user *models.User) error {
		r := models.SchoolFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		if *ses.GetRole() == models.RoleTeacher {
			r.ID = ses.GetSchoolId()
		}
		models, _, total, err := schoolListQuery(&ses, r, false)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"total":   total,
			"schools": models,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SchoolDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSchools, func(user *models.User) error {
		id := c.Param("id")
		if ok, err := schoolAvailableCheck(&ses, models.SchoolFilterRequest{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		m, err := app.SchoolsDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"school": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SchoolCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminSchools, func(user *models.User) (err error) {
		r := models.SchoolRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}

		err = app_validation.ValidateSchoolsCreate(&ses, r)
		if err != nil {
			return err
		}

		galleries, err := handleFilesUpload(c, "galleries", "schools")
		if err != nil {
			return app.NewAppError(err.Error(), "galleries", "")
		}

		if galleries != nil {
			r.Galleries = &galleries
		}

		avatar, _, err := handleFileUpload(c, "avatar", "schools", true)
		if err != nil {
			return app.NewAppError(err.Error(), "avatar", "")
		}
		if avatar != "" {
			r.Avatar = &avatar
		}

		m, err := app.SchoolsCreate(&ses, &r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         m.ID,
			Subject:           models.LogSubjectSchools,
			SubjectAction:     models.LogActionCreate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"school": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SchoolUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminUsers, func(user *models.User) (err error) {
		r := models.SchoolRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		id := c.Param("id")
		r.ID = &id
		if id == "" {
			return app.ErrRequired.SetKey("id")
		}
		err = app_validation.ValidateSchoolsCreate(&ses, r)
		if err != nil {
			return err
		}

		var model models.School
		if model.Galleries == nil {
			model.Galleries = new([]string)
		}

		galleries, err := handleFilesUpload(c, "galleries", "schools")
		if err != nil {
			return app.NewAppError(err.Error(), "galleries", "")
		}
		r.Galleries = &galleries

		if v, ok := c.Request.Form["galleries_delete"]; ok {
			r.GalleriesDelete = &v
		}

		avatar, _, err := handleFileUpload(c, "avatar", "schools", true)
		if err != nil {
			return app.NewAppError(err.Error(), "avatar", "")
		}
		if avatar != "" {
			r.Avatar = &avatar
		}

		if r.AvatarDelete != nil && *r.AvatarDelete {
			r.Avatar = new(string)
		}

		m, err := app.SchoolsUpdate(&ses, &r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         r.ID,
			Subject:           models.LogSubjectSchools,
			SubjectAction:     models.LogActionUpdate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"school": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SchoolDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminUsers, func(user *models.User) error {
		var ids []string = c.QueryArray("ids")
		if len(ids) == 0 {
			return app.ErrRequired.SetKey("ids")
		}

		l, err := app.SchoolsDelete(&ses, ids)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			Subject:           models.LogSubjectSchools,
			SubjectAction:     models.LogActionDelete,
			SubjectProperties: ids,
		})
		Success(c, gin.H{
			"schools": l,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
