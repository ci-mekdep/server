package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
)

func DiaryRoutes(api *gin.RouterGroup) {
	rs := api.Group("/parent")
	{
		rs.GET("/children", ParentChildren)
		rs.GET("/diary", StudentDiary)
		rs.GET("/subjects", ParentSubjects)
		rs.GET("/diary/list", StudentSubjectGrades)
	}
}

func ParentChildren(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermChildren, func(user *models.User) error {
		uu, err := app.Ap().ParentChildren(&ses, ses.GetSchoolId(), user)
		if err != nil {
			return err
		}

		Success(c, gin.H{
			"children": uu,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
	}
}

func StudentSubjectGrades(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermDiary, func(user *models.User) error {
		student, err := getParentChild(c, &ses, user)
		if err != nil {
			return err
		}

		subjectId := c.Query("subject_id")
		if err != nil {
			return err
		}

		periodNumber, err := strconv.Atoi(c.Query("period_number"))
		if err != nil {
			return err
		}

		res, err := app.Ap().GetAllGradesForSubject(&ses, student, subjectId, periodNumber)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"data": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func StudentDiary(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermDiary, func(user *models.User) error {
		date, err := getDate(c, false)
		if err != nil {
			return err
		}

		student, err := getParentChild(c, &ses, user)
		if err != nil {
			return err
		}

		classroomId := c.Request.FormValue("classroom_id")
		if classroomId == "" {
			return app.ErrRequired.SetKey("classroom_id")
		}
		var classroom *models.Classroom
		for _, v := range student.Classrooms {
			if v.ClassroomId == classroomId {
				classroom = v.Classroom
			}
		}
		if classroom == nil {
			return app.ErrRequired.SetKey("classroom_id").SetComment("not found")
		}

		isReview := c.Query("is_review") == "1" || c.Query("is_review") == "true"
		if *ses.GetRole() != models.RoleParent {
			isReview = false
		}

		res, err := app.Ap().StudentDiary(&ses, *student, *classroom, date, isReview)
		if err != nil {
			return err
		}
		Success(c, gin.H{"data": res})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ParentSubjects(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermDiary, func(user *models.User) error {
		student, err := getParentChild(c, &ses, user)
		if err != nil {
			return err
		}

		date, err := getDate(c, false)
		if err != nil {
			return err
		}

		classroomId := c.Request.FormValue("classroom_id")
		if classroomId == "" {
			return app.ErrRequired.SetKey("classroom_id")
		}

		res, err := app.Ap().StudentSubjects(&ses, student, classroomId, date)
		if err != nil {
			return err
		}
		Success(c, gin.H{"items": res})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func getDate(c *gin.Context, isRequired bool) (time.Time, error) {
	var err error
	date := time.Now()
	dateStr := c.Query("date")
	if dateStr != "" {
		date, err = ParseDate(dateStr)
		if err != nil {
			if isRequired {
				return time.Time{}, app.ErrInvalid.SetKey("date")
			}
			date = time.Now()
		}
	} else if isRequired {
		return time.Time{}, app.ErrRequired.SetKey("date")
	}
	return date, nil
}

func getParentChild(c *gin.Context, ses *utils.Session, user *models.User) (*models.User, error) {
	if *ses.GetRole() == models.RoleStudent {
		return user, nil
	}
	var err error
	user.Children, err = app.UserChildrenGet(ses, user)
	if err != nil {
		return nil, err
	}
	var student *models.User
	if len(user.Children) < 1 {
		return nil, app.ErrNotExists.SetKey("child_id")
	}
	childId := ""
	childIdStr := c.Query("child_id")
	if childIdStr != "" {
		childId = childIdStr
	}
	if childId == "" {
		student = user.Children[0]
	} else {
		for _, v := range user.Children {
			if v.ID == childId {
				student = v
			}
		}
	}
	if student == nil {
		return nil, app.ErrNotExists.SetKey("child_id")
	}
	student.Classrooms = nil
	err = store.Store().UsersLoadRelationsClassrooms(ses.Context(), &[]*models.User{student})
	if err != nil {
		return nil, err
	}
	return student, nil
}
