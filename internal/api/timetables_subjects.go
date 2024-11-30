package api

import (
	"encoding/csv"
	"net/http"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
)

func SubjectRoutes(api *gin.RouterGroup) {
	r := api.Group("/subjects")
	{
		r.GET("", SubjectList)
		r.GET("/csv", SubjectDownloadCsv)
		r.GET(":id", SubjectDetail)
		r.PUT(":id", SubjectUpdate)
		r.POST("batch", CreateSubjectsByNames)
		r.POST("", SubjectCreate)
		r.DELETE("", SubjectDelete)
		r.GET("/values", SubjectListValues)
	}
}

func subjectListQuery(ses *utils.Session, data models.SubjectFilterRequest, IsValue bool) ([]*models.SubjectResponse, []models.SubjectValueResponse, int, error) {
	if data.SchoolIds == nil {
		data.SchoolIds = []string{}
	}
	if len(data.SchoolIds) > 0 {
		sids := data.SchoolIds
		data.SchoolIds = []string{}
		for _, sid := range sids {
			if slices.Contains(ses.GetSchoolIds(), sid) {
				data.SchoolIds = append(data.SchoolIds, sid)
			}
		}
	}
	if len(data.SchoolIds) < 1 {
		data.SchoolIds = ses.GetSchoolIds()
	}
	if IsValue {
		res, total, err := app.SubjectListValues(ses, data)
		return nil, res, total, err
	} else {
		res, total, err := app.SubjectsList(ses, &data)
		return res, nil, total, err
	}
}

func subjectAvailableCheck(ses *utils.Session, data models.SubjectFilterRequest) (bool, error) {
	_, _, t, err := subjectListQuery(ses, data, false)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

func SubjectListValues(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSubjects, func(user *models.User) (err error) {
		r := models.SubjectFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}

		_, subjects, total, err := subjectListQuery(&ses, r, true)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"total":    total,
			"subjects": subjects,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SubjectDownloadCsv(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSubjects, func(user *models.User) error {
		r := models.SubjectFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		if r.Limit == nil {
			r.Limit = new(int)
			*r.Limit = 2000
		}
		if r.Offset == nil {
			r.Offset = new(int)
			*r.Offset = 0
		}
		if r.Sort == nil {
			r.Sort = new(string)
			*r.Sort = "teacher_uid~"
		}
		if ses.GetRole() != nil && *ses.GetRole() == models.RoleAdmin {
		} else if ses.GetRole() != nil && *ses.GetRole() == models.RoleOrganization {
		} else if ses.GetRole() != nil && *ses.GetRole() == models.RolePrincipal {
		} else if ses.GetRole() != nil && *ses.GetRole() == models.RoleTeacher {
			r.TeacherIds = []string{user.ID}
		} else if ses.GetRole() != nil && *ses.GetRole() == models.RoleStudent {
			r.ClassroomIds = []string{}
			for _, v := range ses.GetUser().Classrooms {
				r.ClassroomIds = append(r.ClassroomIds, v.ClassroomId)
			}
		} else {
			return app.ErrForbidden
		}
		r.SchoolId = ses.GetSchoolId()
		if r.SchoolId == nil || *r.SchoolId == "" {
			return app.ErrForbidden.SetComment("Mekdep saylanmadyk")
		}

		l, _, err := store.Store().SubjectsListFilters(ses.Context(), &r) // (&ses, r, false)
		if err != nil {
			return err
		}
		err = store.Store().SubjectsLoadRelations(ses.Context(), &l, false)
		if err != nil {
			return err
		}

		items := [][]string{}
		items = append(items, []string{
			"Mugallym",
			"Synp",
			"Ders",
			"Yzygider sagat",
			"Hepdelik sagat",
		})
		for _, v := range l {
			if v.WeekHours == nil || *v.WeekHours == 0 {
				v.WeekHours = new(uint)
				*v.WeekHours = 1
			}
			if v.Teacher == nil {
				v.Teacher = &models.User{}
			}
			if v.Classroom == nil {
				v.Classroom = &models.Classroom{}
			}
			items = append(items, []string{
				v.Teacher.FullName(),
				*v.Classroom.Name,
				*v.Name,
				strconv.Itoa(1),
				strconv.Itoa(int(*v.WeekHours)),
			})
		}

		w := c.Writer
		// Set our headers so browser will download the file
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment;filename=subjects.csv")
		// Create a CSV writer using our HTTP response writer as our io.Writer
		wr := csv.NewWriter(w)
		// Write all items and deal with errors
		if err := wr.WriteAll(items); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SubjectList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSubjects, func(user *models.User) error {
		r := models.SubjectFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		if r.Limit == nil {
			r.Limit = new(int)
			*r.Limit = 100
		}
		if r.Offset == nil {
			r.Offset = new(int)
			*r.Offset = 0
		}
		if ses.GetRole() != nil && slices.Contains([]models.Role{models.RoleAdmin, models.RoleOrganization, models.RoleOperator, models.RolePrincipal}, *ses.GetRole()) {
		} else if ses.GetRole() != nil && *ses.GetRole() == models.RoleTeacher {
			r.TeacherIds = []string{user.ID}
		} else if ses.GetRole() != nil && *ses.GetRole() == models.RoleStudent {
			r.ClassroomIds = []string{}
			for _, v := range ses.GetUser().Classrooms {
				r.ClassroomIds = append(r.ClassroomIds, v.ClassroomId)
			}
		} else {
			return app.ErrForbidden
		}
		l, _, total, err := subjectListQuery(&ses, r, false)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"total":    total,
			"subjects": l,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SubjectDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSubjects, func(user *models.User) error {
		id := c.Param("id")

		if ok, err := subjectAvailableCheck(&ses, models.SubjectFilterRequest{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		m, err := app.SubjectsDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"subject": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SubjectUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminSubjects, func(user *models.User) (err error) {
		r := models.SubjectRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		id := c.Param("id")
		r.ID = &id
		if id == "" {
			return app.ErrRequired.SetKey("id")
		}

		err = app_validation.ValidateSubjectsCreate(&ses, r)
		if err != nil {
			return err
		}
		if ok, err := subjectAvailableCheck(&ses, models.SubjectFilterRequest{ID: r.ID}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		m, err := app.SubjectsUpdate(&ses, &r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         r.ID,
			Subject:           models.LogSubjectSubject,
			SubjectAction:     models.LogActionUpdate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"subject": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func SubjectCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminSubjects, func(user *models.User) (err error) {
		r := models.SubjectRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}

		err = app_validation.ValidateSubjectsCreate(&ses, r)
		if err != nil {
			return err
		}

		m, err := app.SubjectsCreate(&ses, &r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &m.ID,
			Subject:           models.LogSubjectSubject,
			SubjectAction:     models.LogActionCreate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"subject": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func CreateSubjectsByNames(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminSubjects, func(user *models.User) error {
		dto := models.CreateSubjectsByNamesRequestDto{}
		if errMsg, errKey := BindAndValidate(c, &dto); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}

		err := app.CreateSubjectsByNames(&ses, &dto)
		if err != nil {
			return err
		}

		Success(c, gin.H{})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SubjectDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminSubjects, func(user *models.User) error {
		var ids []string = c.QueryArray("ids")
		if len(ids) == 0 {
			return app.ErrRequired.SetKey("ids")
		}

		if ok, err := subjectAvailableCheck(&ses, models.SubjectFilterRequest{Ids: &ids}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		l, err := app.SubjectsDelete(&ses, ids)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			Subject:           models.LogSubjectSubject,
			SubjectAction:     models.LogActionDelete,
			SubjectProperties: ids,
		})
		Success(c, gin.H{
			"subjects": l,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
