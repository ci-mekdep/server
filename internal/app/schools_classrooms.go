package app

import (
	"errors"
	"strings"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func ClassroomsList(ses *utils.Session, data models.ClassroomFilterRequest) ([]*models.ClassroomResponse, int, error) {
	if data.ID == nil {
		data.ID = new(string)
	}
	l, total, err := store.Store().ClassroomsFindBy(ses.Context(), data)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().ClassroomsLoadRelations(ses.Context(), &l, false)
	if err != nil {
		return nil, 0, err
	}
	res := []*models.ClassroomResponse{}
	for _, m := range l {
		item := models.ClassroomResponse{}
		item.FromModel(m)
		res = append(res, &item)
	}
	return res, total, nil
}

func ClassroomListValues(ses *utils.Session, data models.ClassroomFilterRequest) ([]models.ClassroomValueResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ClassroomListValues", "app")
	ses.SetContext(ctx)
	defer sp.End()
	res, total, err := ClassroomsList(ses, data)
	if err != nil {
		return nil, 0, err
	}
	values := []models.ClassroomValueResponse{}
	for _, v := range res {
		values = append(values, v.ToValues())
	}
	return values, total, nil
}

func ClassroomsDetail(ses *utils.Session, id string) (*models.ClassroomResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ClassroomsDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	m, err := store.Store().ClassroomsFindById(ses.Context(), id)
	if err != nil {
		return nil, err
	}
	err = store.Store().ClassroomsLoadRelations(ses.Context(), &[]*models.Classroom{m}, true)
	if err != nil {
		return nil, err
	}
	res := &models.ClassroomResponse{}
	res.FromModel(m)
	return res, nil
}

func ClassroomsCreate(ses *utils.Session, data *models.ClassroomRequest) (*models.ClassroomResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ClassroomsCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	dataModel := &models.Classroom{}
	data.ToModel(dataModel)
	m, err := store.Store().ClassroomsCreate(ses.Context(), dataModel)
	if err != nil {
		return nil, err
	}
	if ses.GetSchool() != nil && !*ses.GetSchool().IsSecondarySchool {
		if data.Subjects != nil {
			for _, sr := range data.Subjects {
				if sr.WeekHours != nil && *sr.WeekHours < 0 {
					return nil, ErrInvalid.SetKey("week_hours")
				}
				subject := &models.Subject{}
				sr.ToModel(subject)
				subject.ClassroomId = m.ID
				subject.SchoolId = m.SchoolId
				if subject.TeacherId == nil {
					return nil, ErrRequired.SetKey("teacher_id")
				}
				createSubject, err := store.Store().SubjectsCreate(ses.Context(), subject)
				if err != nil {
					return nil, err
				}
				if sr.Exams != nil {
					for _, exam := range sr.Exams {
						if exam.TeacherId == nil {
							return nil, ErrRequired.SetKey("teacher_id")
						}
						subjectExam := &models.SubjectExam{}
						exam.ToModel(subjectExam)
						subjectExam.SubjectId = createSubject.ID
						subjectExam.ClassroomId = m.ID
						subjectExam.SchoolId = m.SchoolId
						if exam.ID != nil && *exam.ID == "" {
							subjectExam, err = store.Store().SubjectExamCreate(ses.Context(), subjectExam)
						} else {
							subjectExam, err = store.Store().SubjectExamUpdate(ses.Context(), subjectExam)
						}
						if err != nil {
							return nil, err
						}
						createSubject.Exams = append(createSubject.Exams, subjectExam)
					}
				}
				err = store.Store().SubjectsLoadRelations(ses.Context(), &[]*models.Subject{createSubject}, true)
				if err != nil {
					return nil, err
				}
				dataModel.Subjects = append(dataModel.Subjects, createSubject)
			}
		}
	}
	err = store.Store().ClassroomsUpdateRelations(ses.Context(), dataModel, m)
	if err != nil {
		return nil, err
	}
	err = store.Store().ClassroomsLoadRelations(ses.Context(), &[]*models.Classroom{m}, true)
	if err != nil {
		return nil, err
	}
	if ses.GetSchool() != nil && !*ses.GetSchool().IsSecondarySchool {
		var timetableValue [][]string
		shiftResponse := models.ShiftResponse{}
		if m.Shift == nil {
			return nil, ErrRequired.SetKey("shift")
		}
		shiftResponse.FromModel(m.Shift)

		for _, v := range shiftResponse.Value {
			if len(v) > 0 && len(v[0]) > 0 {
				for _, sr := range m.Subjects {
					timetableValue = append(timetableValue, []string{sr.ID})
				}
			} else {
				timetableValue = append(timetableValue, []string{})
			}
		}
		// create timetable
		timetableCreate := models.TimetableRequest{
			SchoolId:    m.SchoolId,
			ClassroomId: m.ID,
			ShiftId:     m.ShiftId,
			IsThisWeek:  true,
			Value:       timetableValue,
			UpdatedBy:   ses.GetUser().ID,
		}

		timetableResponse, err := app.TimetableCreate(ses, timetableCreate)
		if err != nil {
			return nil, err
		}
		timetableId := timetableResponse.ID
		// update timetable
		timetableUpdate := models.TimetableRequest{
			ID:    &timetableId,
			Value: timetableValue,
		}
		_, err = app.TimetableUpdate(ses, ses.GetUser(), timetableUpdate)
		if err != nil {
			return nil, err
		}
	}
	res := &models.ClassroomResponse{}
	res.FromModel(m)
	return res, nil
}

func ClassroomsUpdateRelations(ses *utils.Session, classroomId string, data models.ClassroomRequest) error {
	sp, ctx := apm.StartSpan(ses.Context(), "ClassroomsCreateStudents", "app")
	ses.SetContext(ctx)
	defer sp.End()
	classroom, err := store.Store().ClassroomsFindById(ses.Context(), classroomId)
	if err != nil {
		return err
	}
	err = store.Store().ClassroomsLoadRelations(ses.Context(), &[]*models.Classroom{classroom}, true)
	if err != nil {
		return err
	}
	if classroom == nil {
		return ErrNotfound.SetKey("classroom_id")
	}
	dataModel := &models.Classroom{
		ID: classroomId,
	}
	data.ToModel(dataModel)
	err = store.Store().ClassroomsUpdateRelations(ses.Context(), dataModel, classroom)
	if err != nil {
		return err
	}
	return nil
}

func ClassroomsUpdate(ses *utils.Session, data *models.ClassroomRequest) (*models.ClassroomResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ClassroomsUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	dataModel := &models.Classroom{}
	data.ToModel(dataModel)
	m, err := store.Store().ClassroomsUpdate(ses.Context(), dataModel)
	if err != nil {
		return nil, err
	}
	if ses.GetSchool() != nil && !*ses.GetSchool().IsSecondarySchool {
		if data.Subjects != nil {
			for _, sr := range data.Subjects {
				subject := &models.Subject{}
				sr.ToModel(subject)
				subject.ClassroomId = m.ID
				subject.SchoolId = m.SchoolId
				updateSubject, err := store.Store().SubjectsUpdate(ses.Context(), subject)
				if err != nil {
					return nil, err
				}
				if sr.Exams != nil {
					for _, exam := range sr.Exams {
						if exam.TeacherId == nil {
							return nil, ErrRequired.SetKey("teacher_id")
						}
						subjectExam := &models.SubjectExam{}
						exam.ToModel(subjectExam)
						subjectExam.SubjectId = updateSubject.ID
						subjectExam.ClassroomId = m.ID
						subjectExam.SchoolId = m.SchoolId
						if exam.ID != nil && *exam.ID == "" {
							subjectExam, err = store.Store().SubjectExamCreate(ses.Context(), subjectExam)
						} else {
							subjectExam, err = store.Store().SubjectExamUpdate(ses.Context(), subjectExam)
						}
						if err != nil {
							return nil, err
						}
						updateSubject.Exams = append(updateSubject.Exams, subjectExam)
					}
				}
				err = store.Store().SubjectsLoadRelations(ses.Context(), &[]*models.Subject{updateSubject}, true)
				if err != nil {
					return nil, err
				}
				dataModel.Subjects = append(dataModel.Subjects, updateSubject)
			}
		}
	}
	err = store.Store().ClassroomsLoadRelations(ses.Context(), &[]*models.Classroom{m}, true)
	if err != nil {
		return nil, err
	}
	err = store.Store().ClassroomsUpdateRelations(ses.Context(), dataModel, m)
	if err != nil {
		return nil, err
	}
	if ses.GetSchool() != nil && !*ses.GetSchool().IsSecondarySchool {
		var timetableValue [][]string
		shiftResponse := models.ShiftResponse{}
		if m.Shift == nil {
			return nil, ErrRequired.SetKey("shift")
		}
		shiftResponse.FromModel(m.Shift)

		for _, v := range shiftResponse.Value {
			if len(v) > 0 && len(v[0]) > 0 {
				for _, sr := range m.Subjects {
					timetableValue = append(timetableValue, []string{sr.ID})
				}
			} else {
				timetableValue = append(timetableValue, []string{})
			}
		}
		// create timetable
		timetableCreate := models.TimetableRequest{
			SchoolId:    m.SchoolId,
			ClassroomId: m.ID,
			ShiftId:     m.ShiftId,
			IsThisWeek:  true,
			Value:       timetableValue,
			UpdatedBy:   ses.GetUser().ID,
		}

		timetableResponse, err := app.TimetableCreate(ses, timetableCreate)
		if err != nil {
			return nil, err
		}
		timetableId := timetableResponse.ID
		// update timetable
		timetableUpdate := models.TimetableRequest{
			ID:    &timetableId,
			Value: timetableValue,
		}
		_, err = app.TimetableUpdate(ses, ses.GetUser(), timetableUpdate)
		if err != nil {
			return nil, err
		}
	}
	res := &models.ClassroomResponse{}
	res.FromModel(m)
	return res, nil
}

func ClassroomsDelete(ses *utils.Session, ids []string) ([]*models.Classroom, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ClassroomDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	l, err := store.Store().ClassroomsFindByIds(ses.Context(), ids)
	if err != nil {
		return nil, err
	}
	if len(l) < 1 {
		return nil, errors.New("model not found: " + strings.Join(ids, ","))
	}
	return store.Store().ClassroomsDelete(ses.Context(), l)
}
