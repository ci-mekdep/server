package app

import (
	apiutils "github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func absentMake(ses *apiutils.Session, data *models.AbsentRequest, lesson models.Lesson) ([]*models.Absent, []*models.PeriodGrade, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "absentMake", "app")
	ses.SetContext(ctx)
	defer sp.End()
	absents := []*models.Absent{}
	periodGrades := []*models.PeriodGrade{}
	var err error
	if data != nil {
		// TODO refactor: new function for Principal role?
		if data.StudentId != "" {
			data.StudentIds = []string{data.StudentId}
		}

		absents, _, err = store.Store().AbsentsFindBy(ses.Context(), models.AbsentFilterRequest{
			StudentIds: &data.StudentIds,
			LessonId:   &lesson.ID,
		})
		if err != nil {
			return nil, nil, err
		}

		for _, studentId := range data.StudentIds {
			newAbsent := models.Absent{}
			newAbsent.FromRequest(data)
			newAbsent.StudentId = studentId
			newAbsent.LessonId = lesson.ID
			newAbsent.Lesson = &lesson
			oldAbsent := models.Absent{}
			for _, v := range absents {
				if v.StudentId == newAbsent.StudentId && v.LessonId == newAbsent.LessonId {
					oldAbsent = *v
				}
			}
			if oldAbsent.ID != "" && (oldAbsent.IsUpdateExpired() && *ses.GetRole() != models.RoleAdmin) {
				err = ErrAbsentUpdateExpired
				continue
			}
			if oldAbsent.ID == "" && !data.IsValueDelete() && newAbsent.IsCreateExpired() {
				err = ErrAbsentUpdateExpired
				continue
			}
			if data.IsValueDelete() {
				_, err := store.Store().AbsentsDelete(ses.Context(), []*models.Absent{&newAbsent})
				if err != nil {
					continue
				}
				absents = []*models.Absent{}
			} else {
				newAbsent, err = store.Store().AbsentsCreateOrUpdate(ses.Context(), newAbsent)
				if err != nil {
					continue
				}
				if oldAbsent.ID != newAbsent.ID {
					absents = append(absents, &newAbsent)
				}
			}
			// update period grade
			if lesson.PeriodId != nil && lesson.PeriodKey != nil {
				// pg, err := store.Store().PeriodGradesUpdateValues(ses.Context(), models.PeriodGrade{
				// 	StudentId: &studentId,
				// 	PeriodId:  lesson.PeriodId,
				// 	PeriodKey: *lesson.PeriodKey,
				// 	SubjectId: &lesson.SubjectId,
				// })
				// if err != nil {
				// 	continue
				// }
				// if pg != nil {
				// 	periodGrades = append(periodGrades, pg)
				// }
			}
		}
		if err != nil {
			return nil, nil, err
		}
	}
	return absents, periodGrades, nil
}

var (
	ErrAbsentNotExists     = ErrNotExists.SetKey("absent")
	ErrAbsentUpdateExpired = ErrExpired.SetKey("absent").SetComment("update expired")
)
