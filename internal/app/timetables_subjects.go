package app

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func SubjectsList(ses *utils.Session, data *models.SubjectFilterRequest) ([]*models.SubjectResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SubjectsList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	if data.ID == nil {
		data.ID = new(string)
	}
	l, total, err := store.Store().SubjectsListFilters(context.Background(), data)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().SubjectsLoadRelations(context.Background(), &l, true)
	if err != nil {
		return nil, 0, err
	}
	res := []*models.SubjectResponse{}
	for _, m := range l {
		item := models.SubjectResponse{}
		item.FromModel(m)
		res = append(res, &item)
	}
	return res, total, nil
}

func SubjectListValues(ses *utils.Session, data models.SubjectFilterRequest) ([]models.SubjectValueResponse, int, error) {
	res, total, err := SubjectsList(ses, &data)
	if err != nil {
		return nil, 0, err
	}
	values := []models.SubjectValueResponse{}
	for _, v := range res {
		values = append(values, v.ToValues())
	}
	return values, total, nil
}

func SubjectsDetail(ses *utils.Session, id string) (*models.SubjectResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SubjectsDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	m, err := store.Store().SubjectsFindById(context.Background(), id)
	if err != nil {
		return nil, err
	}
	err = store.Store().SubjectsLoadRelations(context.Background(), &[]*models.Subject{m}, true)
	if err != nil {
		return nil, err
	}
	if m.Classroom != nil {
		err = store.Store().ClassroomsLoadRelations(ses.Context(), &[]*models.Classroom{m.Classroom}, true)
		if err != nil {
			return nil, err
		}
	}
	res := &models.SubjectResponse{}
	res.FromModel(m)
	return res, nil
}

// TODO: refactor: validation, pgx.ErrNoRows check
func CreateSubjectsByNames(ses *utils.Session, dto *models.CreateSubjectsByNamesRequestDto) error {
	sp, ctx := apm.StartSpan(ses.Context(), "SubjectsCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()

	schoolId := ses.GetSchoolId()
	if schoolId == nil {
		return NewAppError("required", "", "school is not selected")
	}

	for n, subject := range dto.Subjects {
		row := n + 1
		// validate
		if subject.Name == "" {
			return ErrRequired.
				SetKey("subjects." + strconv.Itoa(row) + ".name")
		}
		if subject.ClassroomName == "" {
			return ErrRequired.
				SetKey("subjects." + strconv.Itoa(row) + ".classroom_name")
		}
		if subject.TeacherFullName == nil || *subject.TeacherFullName == "" {
			return ErrRequired.
				SetKey("subjects." + strconv.Itoa(row) + ".teacher_full_name")
		}
		if subject.WeekHoursStr == nil {
			return ErrRequired.
				SetKey("subjects." + strconv.Itoa(row) + ".week_hours")
		}
		// format
		*subject.TeacherFullName = strings.Trim(strings.ReplaceAll(*subject.TeacherFullName, "  ", " "), " ")
		subject.Name = strings.Trim(strings.ReplaceAll(subject.Name, "  ", " "), " ")
		subject.ClassroomName = strings.Trim(strings.ReplaceAll(subject.ClassroomName, "  ", " "), " ")
		if subject.WeekHoursStr != nil && *subject.WeekHoursStr != "" {
			subject.WeekHours = new(uint)
			num, _ := strconv.Atoi(*subject.WeekHoursStr)
			*subject.WeekHours = uint(num)
		}
		if subject.ClassroomTypeKeyStr != nil && *subject.ClassroomTypeKeyStr != "" {
			subject.ClassroomTypeKey = new(uint)
			num, _ := strconv.Atoi(*subject.ClassroomTypeKeyStr)
			*subject.ClassroomTypeKey = uint(num)
		}

		// db validate
		isSubjectExists := false
		for _, v := range models.DefaultSubjects {
			if subject.Name == v[0] || subject.Name == v[1] {
				isSubjectExists = true
				subject.Name = v[0]
				subject.FullName = v[1]
			}
		}
		if !isSubjectExists {
			return ErrNotExists.
				SetKey("subjects." + strconv.Itoa(row) + ".name").
				SetComment("subject does not exist")
		}

		classrooms, _, err := store.Store().ClassroomsFindBy(
			ses.Context(),
			models.ClassroomFilterRequest{
				SchoolId: schoolId,
				Name:     &subject.ClassroomName,
			},
		)
		if err != nil {
			if err.Error() == "no rows in result set" {
				return ErrNotExists.
					SetKey("subjects." + strconv.Itoa(row) + ".classroom_name").
					SetComment("classroom does not exist")
			}
			return err
		}
		if len(classrooms) < 1 {
			return ErrNotExists.
				SetKey("subjects." + strconv.Itoa(row) + ".classroom_name").
				SetComment("classroom does not exist")
		}
		err = store.Store().ClassroomsLoadRelations(ses.Context(), &classrooms, false)
		classroom := classrooms[0]
		classroomType := ""
		if len(classroom.SubGroups) > 0 && classroom.SubGroups[0].Type != nil {
			classroomType = *classroom.SubGroups[0].Type
		}
		teacherNames := strings.SplitN(*subject.TeacherFullName, " ", 3)
		if len(teacherNames) < 2 {
			return ErrInvalid.
				SetKey("subjects." + strconv.Itoa(row) + ".teacher_full_name").
				SetComment("no last name & first name")
		}

		teacherId, err := store.Store().GetTeacherIdByName(
			ses.Context(),
			models.GetTeacherIdByNameQueryDto{
				SchoolId:  *schoolId,
				LastName:  &teacherNames[0],
				FirstName: teacherNames[1],
			},
		)
		// TOOD: make constant
		if err != nil && err.Error() == "no rows in result set" {
			return ErrNotExists.
				SetKey("subjects." + strconv.Itoa(row) + ".teacher_full_name").
				SetComment("teacher does not exist")
		}
		if err != nil {
			return err
		}
		if teacherId == nil {
			return ErrNotExists.
				SetKey("subjects." + strconv.Itoa(row) + ".teacher_full_name").
				SetComment("teacher does not exist")
		}

		createSubjectDto := models.SubjectRequest{
			SchoolId:    *schoolId,
			ClassroomId: classroom.ID,
			Name:        &subject.Name,
			FullName:    &subject.FullName,
			TeacherId:   teacherId,
			WeekHours:   subject.WeekHours,
		}

		if subject.ClassroomTypeKey != nil && *subject.ClassroomTypeKey == 0 {
			createSubjectDto.ClassroomType = &classroomType
		} else {
			createSubjectDto.ClassroomTypeKey = nil
		}
		SubjectsCreate(ses, &createSubjectDto)
	}

	return nil
}

func SubjectsCreate(ses *utils.Session, data *models.SubjectRequest) (*models.SubjectResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SubjectsCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	if data.SchoolId == "" {
		return nil, ErrRequired.SetKey("school_id")
	}
	if data.TeacherId == nil {
		return nil, ErrRequired.SetKey("teacher_id")
	}
	m := &models.Subject{}
	data.ToModel(m)
	var err error

	m, err = store.Store().SubjectsCreate(context.Background(), m)
	if err != nil {
		return nil, err
	}
	for _, subjectExam := range data.Exams {
		exam := &models.SubjectExam{}
		subjectExam.ToModel(exam)
		exam.SubjectId = m.ID
		exam.ClassroomId = m.ClassroomId
		exam.SchoolId = m.SchoolId
		createExam, err := store.Store().SubjectExamCreate(ses.Context(), exam)
		if err != nil {
			return nil, err
		}
		m.Exams = append(m.Exams, createExam)
	}
	res := &models.SubjectResponse{}
	err = store.Store().SubjectsLoadRelations(context.Background(), &[]*models.Subject{m}, true)
	if err != nil {
		return nil, err
	}
	res.FromModel(m)
	return res, nil
}

func SubjectsUpdate(ses *utils.Session, data *models.SubjectRequest) (*models.SubjectResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SubjectsUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()

	m := &models.Subject{}
	var err error
	if m, err = store.Store().SubjectsFindById(context.Background(), *data.ID); err != nil {
		return nil, err
	}
	// ignore fields
	data.Name = m.Name
	data.FullName = m.FullName
	data.ClassroomId = m.ClassroomId
	data.ToModel(m)
	if err = SubjectsIsExists(ses, m); err != nil {
		return nil, err
	}
	if m, err = store.Store().SubjectsUpdate(context.Background(), m); err != nil {
		return nil, err
	}
	for _, exam := range data.Exams {
		subjectExam := &models.SubjectExam{}
		exam.ToModel(subjectExam)
		subjectExam.SubjectId = m.ID
		subjectExam.ClassroomId = m.ClassroomId
		subjectExam.SchoolId = m.SchoolId
		if exam.ID != nil && *exam.ID == "" {
			subjectExam, err = store.Store().SubjectExamCreate(ses.Context(), subjectExam)
		} else {
			subjectExam, err = store.Store().SubjectExamUpdate(ses.Context(), subjectExam)
		}
		if err != nil {
			return nil, err
		}
		m.Exams = append(m.Exams, subjectExam)
	}
	if err = store.Store().SubjectsLoadRelations(context.Background(), &[]*models.Subject{m}, true); err != nil {
		return nil, err
	}
	res := &models.SubjectResponse{}
	res.FromModel(m)
	return res, nil
}

func SubjectsIsExists(ses *utils.Session, data *models.Subject) error {
	_, t, err := store.Store().SubjectsListFilters(ses.Context(), &models.SubjectFilterRequest{
		NotIds:           &[]string{data.ID},
		SubjectNames:     []string{*data.Name},
		ClassroomIds:     []string{data.ClassroomId},
		ClassroomTypeKey: data.ClassroomTypeKey,
	})
	if err != nil {
		return err
	}
	if t > 0 {
		return ErrUnique.SetKey("subject")
	}
	return nil
}

func SubjectsDelete(ses *utils.Session, ids []string) ([]*models.Subject, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SubjectsDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	l, err := store.Store().SubjectsFindByIds(context.Background(), ids)
	if err != nil {
		return nil, err
	}
	if len(l) < 1 {
		return nil, errors.New("model not found: " + strings.Join(ids, ","))
	}
	return store.Store().SubjectsDelete(context.Background(), l)
}
