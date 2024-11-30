package app

import (
	apiutils "github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func gradeMake(ses *apiutils.Session, data *models.GradeRequest, lesson models.Lesson) ([]*models.Grade, []*models.PeriodGrade, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "gradeMake", "app")
	ses.SetContext(ctx)
	defer sp.End()
	grades := []*models.Grade{}
	periodGrades := []*models.PeriodGrade{}
	var err error
	if data != nil {
		if data.StudentId != "" {
			data.StudentIds = []string{data.StudentId}
		}

		// TODO: dine grade update or create or delete
		// remove findby, update, create queries
		grades, _, err = store.Store().GradesFindBy(ses.Context(), models.GradeFilterRequest{
			StudentIds: &data.StudentIds,
			LessonId:   &lesson.ID,
		})
		if err != nil {
			return nil, nil, err
		}

		var listErr error
		for _, studentId := range data.StudentIds {
			newGrade := models.Grade{}
			newGrade.FromRequest(data)
			newGrade.StudentId = studentId
			newGrade.LessonId = lesson.ID
			newGrade.Lesson = &lesson
			oldGrade := models.Grade{}
			for _, v := range grades {
				if v.StudentId == newGrade.StudentId && v.LessonId == newGrade.LessonId {
					oldGrade = *v
				}
			}
			if oldGrade.ID != "" && (oldGrade.IsUpdateExpired() && *ses.GetRole() != models.RoleAdmin) {
				err = ErrGradeUpdateExpired
				listErr = err
				continue
			}
			phone, _ := ses.GetUser().FormattedPhone()
			if oldGrade.ID == "" && !data.IsValueDelete() && newGrade.IsCreateExpired(phone) {
				err = ErrGradeUpdateExpired
				listErr = err
				continue
			}
			if data.IsValueDelete() {
				_, err = store.Store().GradesDelete(ses.Context(), []*models.Grade{&newGrade})
				if err != nil {
					listErr = err
					continue
				}
				grades = []*models.Grade{}
			} else {
				newGrade, err := store.Store().GradesCreateOrUpdate(ses.Context(), newGrade)
				if err != nil {
					listErr = err
					continue
				}
				if oldGrade.ID != newGrade.ID {
					grades = append(grades, &newGrade)
				}
			}
			// update period grade
			if lesson.PeriodId != nil && lesson.PeriodKey != nil {
				// var pg *models.PeriodGrade
				// pg, err = store.Store().PeriodGradesUpdateValues(ses.Context(), models.PeriodGrade{
				// 	StudentId: &studentId,
				// 	SubjectId: &lesson.SubjectId,
				// 	PeriodId:  lesson.PeriodId,
				// 	PeriodKey: *lesson.PeriodKey,
				// })
				// if err != nil {
				// 	listErr = err
				// 	continue
				// }
				// if pg != nil {
				// 	periodGrades = append(periodGrades, pg)
				// }
			}
		}
		if listErr != nil {
			return nil, nil, err
		}
	}
	return grades, periodGrades, nil
}

var (
	ErrGradeNotExists     = ErrNotExists.SetKey("grade")
	ErrGradeUpdateExpired = ErrExpired.SetKey("grade").SetComment("grade update expired")
)
