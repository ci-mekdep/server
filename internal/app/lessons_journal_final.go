package app

import (
	"slices"
	"strconv"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func LessonFinal(ses *utils.Session, subjectId string) ([]models.LessonFinalResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "LessonFinal", "app")
	ses.SetContext(ctx)
	defer sp.End()
	// fetch subject
	subject, err := store.Store().SubjectsFindById(ses.Context(), subjectId)
	if err != nil {
		return nil, ErrNotExists.SetKey("subject_id")
	}
	err = store.Store().SubjectsLoadRelations(ses.Context(), &[]*models.Subject{subject}, false)
	if err != nil {
		return nil, ErrNotExists.SetKey("subject_id")
	}
	// fetch students
	args := models.UserFilterRequest{
		ClassroomId: &subject.ClassroomId,
	}
	args.Limit = new(int)
	*args.Limit = 100
	students, _, err := store.Store().UsersFindBy(ses.Context(), args)
	if err != nil {
		return nil, err
	}
	studentIds := []string{}
	for _, s := range students {
		studentIds = append(studentIds, s.ID)
	}
	// fetch current calendar
	schoolDate := time.Now()
	if schoolDate.Month() > 7 {
		schoolDate = schoolDate.AddDate(0, 0, 30)
	} else {
		schoolDate = schoolDate.AddDate(0, 0, -30)
	}
	period, _, err := periodsGetByDate(ses, schoolDate, subject.SchoolId)
	if err != nil {
		return nil, err
	}
	if period == nil {
		return nil, ErrNotSet.SetKey("period")
	}
	// fetch periods grades
	periodKeys := period.GetPeriodKeys()
	subjectId = subject.GetId()
	periodGrades, _, err := store.Store().PeriodGradesFindBy(ses.Context(), models.PeriodGradeFilterRequest{
		// PeriodId:   &period.ID,
		PeriodKeys: &periodKeys,
		StudentIds: &studentIds,
		SubjectId:  &subjectId,
	})
	if err != nil {
		return nil, err
	}
	// fetch exam grades
	PeriodGradeExamKey := models.PeriodGradeExamKey
	examGrades, _, err := store.Store().PeriodGradesFindBy(ses.Context(), models.PeriodGradeFilterRequest{
		PeriodKey:  &PeriodGradeExamKey,
		StudentIds: &studentIds,
		SubjectId:  &subject.ID,
	})
	if err != nil {
		return nil, err
	}
	// calculate final grade, all
	res := []models.LessonFinalResponse{}
	for _, student := range students {
		finalGrade := models.PeriodGrade{}
		finalGrade.StudentId = &student.ID
		sPeriodGrades := []models.PeriodGrade{}
		for _, v := range period.GetPeriodKeys() {
			if len(sPeriodGrades) < v {
				sPeriodGrades = append(sPeriodGrades, models.PeriodGrade{})
			}
		}
		for _, v := range periodGrades {
			if *v.StudentId == student.ID {
				if len(sPeriodGrades) >= v.PeriodKey {
					finalGrade.AppendGrade(v)
					sPeriodGrades[v.PeriodKey-1] = *v
				}
			}
		}

		periods := map[string]*models.PeriodGradeResponse{}
		for k, v := range sPeriodGrades {
			resPeriodGrade := models.PeriodGradeResponse{}
			resPeriodGrade.FromModel(&v)
			resPeriodGrade.SetValueByRules()
			periods[strconv.Itoa(k+1)] = &resPeriodGrade
		}

		var exam *models.PeriodGradeResponse
		if subject.Exams != nil {
			examGrade := models.PeriodGrade{}
			for _, item := range examGrades {
				if *item.StudentId == student.ID {
					finalGrade.AppendPowerGrade(item)
					examGrade = *item
					break
				}
			}

			exam = &models.PeriodGradeResponse{}
			exam.FromModel(&examGrade)
		}
		final := &models.PeriodGradeResponse{}
		final.FromModel(&finalGrade)
		resStudent := models.UserResponse{}
		resStudent.FromModel(student)
		res = append(res, models.LessonFinalResponse{
			Student: resStudent,
			Periods: periods,
			Exam:    exam,
			Final:   final,
		})
	}

	return res, nil
}

func LessonFinalBySubject(ses *utils.Session, classroomId string, periodNumber int) (*models.LessonFinalBySubjectResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "LessonFinal", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error
	// fetch classroom
	var classroom *models.Classroom
	classroom, err = store.Store().ClassroomsFindById(ses.Context(), classroomId)
	if err != nil {
		return nil, ErrNotExists.SetKey("classroom_id")
	}
	err = store.Store().ClassroomsLoadRelations(ses.Context(), &[]*models.Classroom{classroom}, true)
	if err != nil {
		return nil, err
	}
	// fetch subjects
	sargs := models.SubjectFilterRequest{
		ClassroomId: &classroom.ID,
	}
	sargs.Limit = new(int)
	*sargs.Limit = 100
	subjects, _, err := store.Store().SubjectsListFilters(ses.Context(), &sargs)
	if err != nil {
		return nil, err
	}
	// fetch exams

	argsExam := models.SubjectExamFilterRequest{
		ClassroomId: &classroom.ID,
	}
	argsExam.Limit = new(int)
	*argsExam.Limit = 100
	exams, _, err := store.Store().SubjectExamsFindBy(ses.Context(), &argsExam)
	if err != nil {
		return nil, err
	}
	// fetch students
	args := models.UserFilterRequest{
		ClassroomId: &classroom.ID,
	}
	args.Limit = new(int)
	*args.Limit = 100
	students, _, err := store.Store().UsersFindBy(ses.Context(), args)
	if err != nil {
		return nil, err
	}
	studentIds := []string{}
	for _, s := range students {
		studentIds = append(studentIds, s.ID)
	}
	// fetch current calendar
	schoolDate := time.Now()
	period, _, err := periodsGetByDate(ses, schoolDate, subjects[0].SchoolId)
	if err != nil {
		return nil, err
	}
	if period == nil {
		return nil, ErrNotSet.SetKey("period")
	}

	// check permission
	if *ses.GetRole() == models.RoleTeacher {
		if classroom.TeacherId == nil || *classroom.TeacherId != ses.GetUser().ID {
			return nil, ErrForbidden.SetKey("classroom_id")
		}
	} else {
		if !slices.Contains(ses.GetSchoolIds(), classroom.SchoolId) {
			return nil, ErrForbidden.SetKey("classroom_id")
		}
	}

	// fetch periods grades
	// periodKeys := period.GetPeriodKeys()
	var periodGrades []*models.PeriodGrade
	periodGrades, _, err = store.Store().PeriodGradesFindBy(ses.Context(), models.PeriodGradeFilterRequest{
		PeriodKeys: &[]int{periodNumber},
		StudentIds: &studentIds,
	})
	if err != nil {
		return nil, err
	}
	// calculate final grade, all
	res := models.LessonFinalBySubjectResponse{}
	for _, student := range students {

		resItem := models.LessonFinalBySubjectStudentResponse{
			Student:  &models.UserResponse{},
			Subjects: []models.LessonFinalBySubjectSubjectResponse{},
		}

		for _, subject := range subjects {
			resItemSubject := models.LessonFinalBySubjectSubjectResponse{
				SubjectId: &subject.ID,
			}
			// set period grades
			if periodNumber == models.PeriodGradeExamKey {
				// set exam grades
				resItemSubject.ExamGrades = &[]models.PeriodGradeResponse{}
				for _, periodGrade := range periodGrades {
					for _, exam := range exams {
						if *periodGrade.StudentId == student.ID &&
							*periodGrade.SubjectId == subject.ID &&
							periodGrade.ExamId != nil &&
							*periodGrade.ExamId == exam.ID {
							periodGradeRes := models.PeriodGradeResponse{}
							periodGradeRes.FromModel(periodGrade)
							*resItemSubject.ExamGrades = append(*resItemSubject.ExamGrades, periodGradeRes)
							break
						}
					}
				}
			} else {
				// set period grades
				resItemSubject.PeriodGrade = &models.PeriodGradeResponse{}
				for _, periodGrade := range periodGrades {
					if *periodGrade.StudentId == student.ID && *periodGrade.SubjectId == subject.ID {
						resItemSubject.PeriodGrade.FromModel(periodGrade)
						break
					}
				}
			}
			resItem.Subjects = append(resItem.Subjects, resItemSubject)
		}

		resItem.Student.FromModel(student)
		res.Students = append(res.Students, resItem)
	}
	for _, subject := range subjects {
		subjectRes := models.SubjectResponse{}
		subjectRes.FromModel(subject)
		res.Subjects = append(res.Subjects, &subjectRes)
	}
	for _, exam := range exams {
		examRes := models.SubjectExamResponse{}
		examRes.FromModel(exam)
		res.Exams = append(res.Exams, &examRes)
	}
	return &res, nil
}

func LessonFinalV2(ses *utils.Session, subjectId, childId, classroomId string) ([]models.LessonFinalResponseV2, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "LessonFinal", "app")
	ses.SetContext(ctx)
	defer sp.End()
	// fetch classroom
	var classroom *models.Classroom
	var err error
	if classroomId != "" {
		classroom, err = store.Store().ClassroomsFindById(ses.Context(), classroomId)
		if err != nil {
			return nil, ErrNotExists.SetKey("classroom_id")
		}
		err = store.Store().ClassroomsLoadRelations(ses.Context(), &[]*models.Classroom{classroom}, true)
		if err != nil {
			return nil, err
		}
	}
	// fetch subject
	argsSubject := models.SubjectFilterRequest{}
	if classroom != nil {
		argsSubject.ClassroomId = &classroom.ID
	}
	if subjectId != "" {
		argsSubject.ID = &subjectId
	}
	subjects, _, err := store.Store().SubjectsListFilters(ses.Context(), &argsSubject)
	if err != nil {
		return nil, ErrNotExists.SetKey("subject_id")
	}
	err = store.Store().SubjectsLoadRelations(ses.Context(), &subjects, false)
	if err != nil {
		return nil, ErrNotExists.SetKey("subject_id")
	}
	// fetch students
	args := models.UserFilterRequest{
		ClassroomId: &subjects[0].ClassroomId,
		ID:          &childId,
	}
	args.Limit = new(int)
	*args.Limit = 100
	students, _, err := store.Store().UsersFindBy(ses.Context(), args)
	if err != nil {
		return nil, err
	}
	studentIds := []string{}
	for _, s := range students {
		studentIds = append(studentIds, s.ID)
	}
	// fetch current calendar
	schoolDate := time.Now()
	period, _, err := periodsGetByDate(ses, schoolDate, subjects[0].SchoolId)
	if err != nil {
		return nil, err
	}
	if period == nil {
		return nil, ErrNotSet.SetKey("period")
	}
	// fetch periods grades
	periodKeys := period.GetPeriodKeys()
	subjectId = subjects[0].GetId()
	periodGrades, _, err := store.Store().PeriodGradesFindBy(ses.Context(), models.PeriodGradeFilterRequest{
		// PeriodId:   &period.ID,
		PeriodKeys: &periodKeys,
		StudentIds: &studentIds,
		SubjectId:  &subjectId,
	})
	if err != nil {
		return nil, err
	}

	//fetch exam
	exams, _, err := store.Store().SubjectExamsFindBy(ses.Context(), &models.SubjectExamFilterRequest{
		SubjectId: &subjectId,
	})
	if err != nil {
		return nil, err
	}
	err = store.Store().SubjectExamLoadRelations(ses.Context(), &exams)
	if err != nil {
		return nil, err
	}
	// fetch exam grades
	PeriodGradeExamKey := models.PeriodGradeExamKey
	examGrades, _, err := store.Store().PeriodGradesFindBy(ses.Context(), models.PeriodGradeFilterRequest{
		PeriodKey:  &PeriodGradeExamKey,
		StudentIds: &studentIds,
		SubjectId:  &subjects[0].ID,
	})
	if err != nil {
		return nil, err
	}
	if subjects[0].School.IsSecondarySchool != nil && *subjects[0].School.IsSecondarySchool == false {
		periodGrades = nil
	}
	// calculate final grade, all
	res := []models.LessonFinalResponseV2{}
	for _, student := range students {
		finalGrade := models.PeriodGrade{}
		finalGrade.StudentId = &student.ID
		sPeriodGrades := []models.PeriodGrade{}
		for _, v := range period.GetPeriodKeys() {
			if len(sPeriodGrades) < v {
				sPeriodGrades = append(sPeriodGrades, models.PeriodGrade{})
			}
		}
		for _, v := range periodGrades {
			if *v.StudentId == student.ID {
				if len(sPeriodGrades) >= v.PeriodKey {
					finalGrade.AppendGrade(v)
					sPeriodGrades[v.PeriodKey-1] = *v
				}
			}
		}

		periods := map[string]*models.PeriodGradeResponse{}
		for k, v := range sPeriodGrades {
			resPeriodGrade := models.PeriodGradeResponse{}
			resPeriodGrade.FromModel(&v)
			resPeriodGrade.SetValueByRules()
			periods[strconv.Itoa(k+1)] = &resPeriodGrade
		}
		if subjects[0].School.IsSecondarySchool != nil && *subjects[0].School.IsSecondarySchool == false {
			periods = nil
		}
		studentExams := make([]*models.ExamWithGrade, 0)
		for k, exam := range exams {
			if exam.ExamWeightPercent == nil {
				exams[k].ExamWeightPercent = new(uint)
				*exams[k].ExamWeightPercent = uint(float64(1/float64(len(exams))) * 100)
			}
			examRes := &models.SubjectExamResponse{}
			examRes.FromModel(exam)

			examToRes := &models.ExamWithGrade{
				Exam: examRes,
			}
			studentExams = append(studentExams, examToRes)
		}
		for _, item := range examGrades {
			if *item.StudentId != student.ID {
				continue
			}

			for k, vv := range studentExams {
				examGradeResponse := models.PeriodGradeResponse{}
				examGradeResponse.FromModel(item)

				if vv.Exam.ID == *item.ExamId {
					studentExams[k].Grade = &examGradeResponse
				}
			}
		}

		examAverage := float64(0)
		for _, v := range examGrades {
			for _, vv := range exams {
				if v.ExamId != nil && *v.ExamId == vv.ID && *v.StudentId == student.ID {
					examAverage += float64(v.GradeIntValue()) * (float64(*vv.ExamWeightPercent) / 100)
				}
			}
		}
		finalGrade.AppendGrade(&models.PeriodGrade{
			GradeSum:   int(examAverage) * 10,
			GradeCount: 10,
		})

		final := &models.PeriodGradeResponse{}
		final.FromModel(&finalGrade)
		resStudent := models.UserResponse{}
		resStudent.FromModel(student)
		res = append(res, models.LessonFinalResponseV2{
			Student: resStudent,
			Periods: periods,
			Exams:   studentExams,
			Final:   final,
		})
	}

	return res, nil
}

func LessonFinalMake(ses *utils.Session, subjectId string, req *models.GradeRequest) (models.PeriodGradeResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "LessonFinalMake", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error
	periodGrades := []*models.PeriodGrade{}
	subjects, _, err := store.Store().SubjectsListFilters(ses.Context(), &models.SubjectFilterRequest{
		ID:       &subjectId,
		SchoolId: ses.GetSchoolId(),
	})
	if len(subjects) < 1 {
		return models.PeriodGradeResponse{}, ErrNotfound.SetComment("ID: " + subjectId)
	}
	if err != nil {
		return models.PeriodGradeResponse{}, err
	}
	err = store.Store().SubjectsLoadRelations(ses.Context(), &subjects, false)
	if err != nil {
		return models.PeriodGradeResponse{}, err
	}
	for _, subject := range subjects {
		if subject.Exams != nil {
			for _, exam := range subject.Exams {
				if exam.TeacherId != ses.GetUser().ID {
					return models.PeriodGradeResponse{}, ErrForbidden.SetComment("teacher is not authorized to access this subject")
				}
			}
		}
	}
	if req != nil {
		periodGrades, err = gradeExamMake(ses, req, subjects[0])
		if err != nil {
			return models.PeriodGradeResponse{}, err
		}
		subId := ""
		subjectAction := models.LogActionCreate
		if len(periodGrades) > 0 {
			subId = periodGrades[0].ID
		}
		if req.IsValueDelete() {
			subjectAction = models.LogActionDelete
			if subId == "" {
				subjectAction = ""
			}
		}
		if subjectAction != "" {
			userLog(ses, models.UserLog{
				SchoolId:          ses.GetSchoolId(),
				SessionId:         ses.GetSessionId(),
				UserId:            ses.GetUser().ID,
				SubjectId:         &subId,
				Subject:           models.LogSubjectGrade,
				SubjectAction:     subjectAction,
				SubjectProperties: req,
			})
		}
	}
	res := models.PeriodGradeResponse{}
	for _, periodGrade := range periodGrades {
		if *periodGrade.StudentId == req.StudentId {
			res.FromModel(periodGrade)
			res.ExamId = periodGrade.ExamId
		}
	}
	return res, nil
}

func gradeExamMake(ses *utils.Session, data *models.GradeRequest, subject *models.Subject) ([]*models.PeriodGrade, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "gradeExamMake", "app")
	ses.SetContext(ctx)
	defer sp.End()

	schoolDate := time.Now()
	period, _, err := periodsGetByDate(ses, schoolDate, *ses.GetSchoolId())
	if err != nil {
		return nil, err
	}
	if period == nil {
		return nil, ErrNotSet.SetKey("period")
	}
	periodGrades := []*models.PeriodGrade{}
	if data != nil {
		if data.StudentId != "" {
			data.StudentIds = []string{data.StudentId}
		}
		//TODO student filter in subject
		if err != nil {
			return nil, err
		}

		for _, studentId := range data.StudentIds {
			var gradeCount int
			var gradeSum int
			if data.Values != nil && len(*data.Values) > 0 {
				gradeCount = len(*data.Values)
				for _, v := range *data.Values {
					gradeSum += v
				}
			} else if data.Value != nil {
				gradeCount = 1
				gradeSum = *data.Value
			} else {
				continue
			}
			newGrade := models.PeriodGrade{
				StudentId:   &studentId,
				PeriodId:    &period.ID,
				PeriodKey:   models.PeriodGradeExamKey,
				SubjectId:   &subject.ID,
				LessonCount: 1,
				ExamId:      data.ExamId,
				GradeCount:  gradeCount,
				GradeSum:    gradeSum,
			}
			oldGrade := models.PeriodGrade{}
			for _, v := range periodGrades {
				if v.StudentId == newGrade.StudentId && v.ExamId == newGrade.ExamId {
					oldGrade = *v
				}
			}
			if oldGrade.ID != "" && oldGrade.IsUpdateExpired() {
				err = ErrGradeUpdateExpired
				continue
			}
			if oldGrade.ID == "" && !data.IsValueDelete() && newGrade.IsCreateExpired(subject.Exams) {
				err = ErrGradeUpdateExpired
				continue
			}
			if data.IsValueDelete() {
				_, err = store.Store().PeriodGradesDelete(ses.Context(), []*models.PeriodGrade{&newGrade})
				if err != nil {
					continue
				}
				periodGrades = []*models.PeriodGrade{}
			} else {
				newGrade, err = store.Store().PeriodGradesUpdateOrCreate(ses.Context(), &newGrade)
				if err != nil {
					continue
				}
				if oldGrade.ID != newGrade.ID {
					periodGrades = append(periodGrades, &newGrade)
				}
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return periodGrades, nil
}
