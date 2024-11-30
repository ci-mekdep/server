package app

import (
	"slices"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

type LessonJournalRequest struct {
	SubjectId    *string    `form:"subject_id" validate:"required"`
	Date         *time.Time `form:"date" time_format:"2006-01-02"`
	HourNumber   *int       `form:"hour_number"`
	PeriodNumber *int       `form:"period_number"`
	OnlyLessons  bool       `form:"only_lessons"`
}

func (a App) LessonJournal(ses *utils.Session, data LessonJournalRequest) (*models.JournalResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "LessonJournal", "app")
	ses.SetContext(ctx)
	defer sp.End()
	// fetch subject
	args := models.SubjectFilterRequest{
		ID: data.SubjectId,
	}

	// check access
	if data.SubjectId == nil {
		return nil, ErrNotExists.SetKey("subject_id")
	}
	if *ses.GetRole() == models.RoleAdmin {
		args.SchoolId = ses.GetSchoolId()
	} else if *ses.GetRole() == models.RoleOrganization {
		args.SchoolId = ses.GetSchoolId()
	} else if *ses.GetRole() == models.RolePrincipal {
		args.SchoolId = ses.GetSchoolId()
	} else {
		args.TeacherIds = []string{ses.GetUser().ID}
	}

	// subject find
	ls, _, err := store.Store().SubjectsListFilters(ses.Context(), &args)
	if err != nil || len(ls) < 1 {
		return nil, ErrNotExists.SetKey("subject_id")
	}
	subject := ls[0]
	subjectId := subject.GetId()
	res := models.JournalResponse{}
	if data.OnlyLessons {
		lessons, _, err := fetchJournalItems(ses, subjectId, subject.SchoolId, data.PeriodNumber, data.Date, data.HourNumber)
		if err != nil {
			return nil, err
		}
		res.Lessons = lessons
		return &res, nil
	}
	// fetch students
	students, err := fetchStudents(ses, subject)
	if err != nil {
		return nil, err
	}
	studentIds := []string{}
	for _, s := range students {
		studentIds = append(studentIds, s.ID)
	}
	// fetch lessons
	var lessons []models.JournalItemResponse
	lessons, data.PeriodNumber, err = fetchJournalItems(ses, subjectId, subject.SchoolId, data.PeriodNumber, data.Date, data.HourNumber)
	if err != nil {
		return nil, err
	}
	// fetch student_notes
	studentNotes, err := fetchNotes(ses, subjectId, studentIds)
	if err != nil {
		return nil, err
	}
	// fetch period grades
	periodGrades, err := fetchPeriodGrades(ses, subjectId, studentIds, data.PeriodNumber)
	if err != nil {
		return nil, err
	}

	res.Students = students
	res.Lessons = lessons
	res.StudentNotes = studentNotes
	res.PeriodGrades = periodGrades

	return &res, nil
}

func fetchJournalItems(ses *utils.Session, subjectId string, schoolId string, periodNumber *int, lessonDate *time.Time, hourNumber *int) ([]models.JournalItemResponse, *int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "fetchJournalItems", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error
	if lessonDate == nil {
		lessonDate = new(time.Time)
		*lessonDate = time.Now()
	}
	if hourNumber == nil {
		hourNumber = new(int)
		*hourNumber = 0
	}
	var periodNumberByDate int
	if periodNumber != nil {
		periodNumberByDate = *periodNumber
	} else {
		_, periodNumberByDate, err = periodsGetByDate(ses, *lessonDate, schoolId)
	}

	// get all lessons by period
	args := models.LessonFilterRequest{}
	args.SubjectId = &subjectId
	limit := 300
	args.Limit = &limit
	args.PeriodNumber = &periodNumberByDate
	lessons, _, err := store.Store().LessonsFindBy(ses.Context(), args)
	if err != nil {
		return nil, nil, err
	}
	// sort lessons by date and hour
	lessons = lessonsSort(lessons)

	// select lessons by filters
	if periodNumber == nil {
		for _, v := range lessons {
			vHour := 100
			if v.HourNumber != nil {
				vHour = *v.HourNumber
			}
			if v.Date.Compare(*lessonDate) >= 0 && vHour >= *hourNumber {
				l := *v
				lessons = []*models.Lesson{&l}
				break
			}
		}
	}

	err = store.Store().LessonsLoadRelations(ses.Context(), &lessons)
	if err != nil {
		return nil, nil, err
	}
	lessonIds := []string{}

	resLessons := []models.JournalItemResponse{}
	for _, l := range lessons {
		r := models.JournalItemResponse{}
		r.Lesson = models.LessonResponse{}
		r.Lesson.FromModel(l)

		// batch fetch later grades by lessonIds
		r.Grades = []models.GradeResponse{}
		// batch fetch later absents by lessonIds
		r.Absents = []models.AbsentResponse{}

		resLessons = append(resLessons, r)
		lessonIds = append(lessonIds, l.ID)
	}
	// fetch grades
	grades, _, err := store.Store().GradesFindBy(ses.Context(), models.GradeFilterRequest{
		LessonIds: &lessonIds,
	})
	if err != nil {
		return nil, nil, err
	}
	for _, g := range grades {
		for k, l := range resLessons {
			if g.LessonId == l.Lesson.ID {
				t := models.GradeResponse{}
				t.FromModel(g)
				resLessons[k].Grades = append(resLessons[k].Grades, t)
			}
		}
	}
	// fetch absents
	absents, _, err := store.Store().AbsentsFindBy(ses.Context(), models.AbsentFilterRequest{
		LessonIds: &lessonIds,
	})
	if err != nil {
		return nil, nil, err
	}
	for _, item := range absents {
		for k, l := range resLessons {
			if item.LessonId == l.Lesson.ID {
				t := models.AbsentResponse{}
				t.FromModel(item)
				resLessons[k].Absents = append(resLessons[k].Absents, t)
			}
		}
	}
	return resLessons, &periodNumberByDate, nil
}

func lessonsSort(lessons []*models.Lesson) []*models.Lesson {
	sort.Slice(lessons, func(i, j int) bool {
		iHour := 99
		jHour := 99
		if lessons[i].HourNumber != nil {
			iHour = *lessons[i].HourNumber
		}
		if lessons[j].HourNumber != nil {
			jHour = *lessons[j].HourNumber
		}
		iDate := lessons[i].Date
		jDate := lessons[j].Date

		if iDate.Equal(jDate) {
			return iHour < jHour
		}
		return iDate.Before(jDate)
	})
	return lessons
}

func fetchStudents(ses *utils.Session, subject *models.Subject) ([]models.UserResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "fetchStudents", "app")
	ses.SetContext(ctx)
	defer sp.End()
	argsUsers := models.UserFilterRequest{
		ClassroomId:      &subject.ClassroomId,
		ClassroomType:    subject.ClassroomType,
		ClassroomTypeKey: subject.ClassroomTypeKey,
	}
	lim := 100
	argsUsers.Limit = &lim
	students, _, err := store.Store().UsersFindBy(ses.Context(), argsUsers)
	if err != nil {
		return nil, err
	}
	resStudents := []models.UserResponse{}
	for _, s := range students {
		r := models.UserResponse{}
		r.FromModel(s)
		resStudents = append(resStudents, r)
	}
	return resStudents, nil
}

func fetchNotes(ses *utils.Session, subjectId string, studentIds []string) ([]models.StudentNoteResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "fetchNotes", "app")
	ses.SetContext(ctx)
	defer sp.End()
	studentNotes, _, err := store.Store().StudentNotesFindBy(ses.Context(), models.StudentNoteFilterRequest{
		StudentIds: &studentIds,
		SubjectId:  &subjectId,
	})
	if err != nil {
		return nil, err
	}
	resStudentNotes := []models.StudentNoteResponse{}
	for _, v := range studentNotes {
		tmp := models.StudentNoteResponse{}
		tmp.FromModel(v)
		resStudentNotes = append(resStudentNotes, tmp)
	}
	return resStudentNotes, nil
}

func fetchPeriodGrades(ses *utils.Session, subjectId string, studentIds []string, periodNumber *int) ([]models.PeriodGradeResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "fetchPeriodGrades", "app")
	ses.SetContext(ctx)
	defer sp.End()
	resPeriodGrades := []models.PeriodGradeResponse{}
	periodGrades, _, err := store.Store().PeriodGradesFindBy(ses.Context(), models.PeriodGradeFilterRequest{
		StudentIds: &studentIds,
		SubjectId:  &subjectId,
		PeriodKey:  periodNumber,
	})
	for _, v := range periodGrades {
		tmp := models.PeriodGradeResponse{}
		tmp.FromModel(v)
		tmp.SetValueByRules()
		resPeriodGrades = append(resPeriodGrades, tmp)
	}
	if err != nil {
		return nil, err
	}
	return resPeriodGrades, nil
}

func LessonUpdateFormLessonPro(c *gin.Context, data *models.JournalFormRequest) (*models.Lesson, error) {
	ml, err := store.Store().LessonsFindById(c, data.LessonId)
	if err != nil {
		return nil, err
	}
	if ml.ProFiles == nil {
		ml.ProFiles = &[]string{}
	}
	// update existing, first delete
	paths := *ml.ProFiles
	if data.LessonProFilesDelete != nil {
		if len(*data.LessonProFilesDelete) < 1 || (*data.LessonProFilesDelete)[0] == "" {
			*data.LessonProFilesDelete = paths
		}
		for _, v := range *data.LessonProFilesDelete {
			filePath := extractFullUrl(v)
			deleteFile(c, filePath, "lesson_pro")
			k := slices.Index(paths, filePath)
			if k >= 0 {
				paths = slices.Delete(paths, k, k+1)
			}
		}
	}
	// then upload
	tmp, err := handleFilesUpload(c, "lesson_pro_files", "lesson_pro")
	if err != nil {
		return nil, err
	}
	paths = append(paths, tmp...)
	ml.ProFiles = &paths
	ml, err = store.Store().LessonsUpdate(c, ml)
	if err != nil {
		return nil, err
	}
	return &ml, nil
}

func LessonUpdateFormAssignment(c *gin.Context, data *models.JournalFormRequest) (*models.Lesson, error) {
	ml, err := store.Store().LessonsFindById(c, data.LessonId)
	if err != nil {
		return nil, err
	}
	if ml.AssignmentFiles == nil {
		ml.AssignmentFiles = &[]string{}
	}

	// update existing, first delete
	paths := *ml.AssignmentFiles
	if data.AssignmentFilesDelete != nil {
		if len(*data.AssignmentFilesDelete) < 1 || (*data.AssignmentFilesDelete)[0] == "" {
			*data.AssignmentFilesDelete = paths
		}
		for _, v := range *data.AssignmentFilesDelete {
			prefix := extractFullUrl(v)
			deleteFile(c, prefix, "assignments")
			k := slices.Index(paths, prefix)
			if k >= 0 {
				paths = slices.Delete(paths, k, k+1)
			}
		}
	}
	// then upload
	tmp, err := handleFilesUpload(c, "assignment_files", "assignments")
	if err != nil {
		return nil, err
	}
	paths = append(paths, tmp...)
	ml.AssignmentFiles = &paths
	ml, err = store.Store().LessonsUpdate(c, ml)
	if err != nil {
		return nil, err
	}
	return &ml, nil
}

func (a App) LessonUpdate(ses *utils.Session, data *models.JournalRequest) (models.JournalItemResponse, []models.PeriodGradeResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "LessonUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	// todo: check student
	// update or create lesson
	var err error
	lesson := models.Lesson{}
	lesson.SchoolId = *ses.GetSchoolId()
	err = lesson.FromRequest(&data.Lesson)

	// TODO: remove update create queries, only query with update or create
	lesson, err = store.Store().LessonsUpdate(ses.Context(), lesson)
	if err != nil {
		return models.JournalItemResponse{}, nil, err
	}

	// update assignment
	assignment := models.Assignment{}
	assignment.FromRequest(&data.Lesson.Assignment)

	// update student note
	studentNotes := []models.StudentNote{}
	if data.StudentNote != nil {
		data.StudentNote.SubjectId = &lesson.SubjectId
		data.StudentNote.SchoolId = lesson.SchoolId
		data.StudentNote.TeacherId = ses.GetUser().ID
		studentNoteData := models.StudentNote{}
		studentNoteData.FromRequest(data.StudentNote)
		studentNote, err := lessonsStudentNotesUpdate(ses, studentNoteData)
		if err != nil {
			return models.JournalItemResponse{}, nil, err
		}
		studentNotes = append(studentNotes, studentNote)
	}

	// if both delete other
	if data.Grade != nil && !data.Grade.IsValueDelete() {
		data.Absent = &models.AbsentRequest{
			StudentId: data.Grade.StudentId,
			UpdatedBy: data.Grade.UpdatedBy,
		}
		data.Absent.Reason = nil
	}
	if data.Absent != nil && !data.Absent.IsValueDelete() {
		data.Grade = &models.GradeRequest{
			StudentId: data.Absent.StudentId,
			UpdatedBy: data.Absent.UpdatedBy,
		}
		data.Grade.Value = nil
		data.Grade.Values = nil
	}

	// update grade
	grades := []*models.Grade{}
	periodGrades := []*models.PeriodGrade{}
	if data.Grade != nil {
		grades, periodGrades, err = gradeMake(ses, data.Grade, lesson)
		if err != nil {
			return models.JournalItemResponse{}, nil, err
		}

		subId := ""
		subAct := models.LogActionCreate
		if len(grades) > 0 {
			subId = grades[0].ID
		}
		if data.Grade.IsValueDelete() {
			subAct = models.LogActionDelete
			if subId == "" {
				subAct = ""
			}
		}
		if subAct != "" {
			if lesson.Subject == nil {
				err = store.Store().LessonsLoadSubject(ctx, &[]*models.Lesson{&lesson})
				if err != nil {
					return models.JournalItemResponse{}, nil, err
				}
			}
			label := lesson.Label()
			userLog(ses, models.UserLog{
				SchoolId:           ses.GetSchoolId(),
				SessionId:          ses.GetSessionId(),
				UserId:             ses.GetUser().ID,
				SubjectId:          &subId,
				Subject:            models.LogSubjectGrade,
				SubjectDescription: &label,
				SubjectAction:      subAct,
				SubjectProperties:  data.Grade,
			})
		}
	}

	// update absent
	absents := []*models.Absent{}
	if data.Absent != nil {
		absents, periodGrades, err = absentMake(ses, data.Absent, lesson)
		if err != nil {
			return models.JournalItemResponse{}, nil, err
		}

		subId := ""
		subAct := models.LogActionCreate
		if len(absents) > 0 {
			subId = absents[0].ID
		}
		if data.Absent.IsValueDelete() {
			subAct = models.LogActionDelete
			if subId == "" {
				subAct = ""
			}
		}

		if subAct != "" {
			if lesson.Subject == nil {
				err = store.Store().LessonsLoadSubject(ctx, &[]*models.Lesson{&lesson})
				if err != nil {
					return models.JournalItemResponse{}, nil, err
				}
			}
			label := lesson.Label()
			userLog(ses, models.UserLog{
				SchoolId:           ses.GetSchoolId(),
				SessionId:          ses.GetSessionId(),
				UserId:             ses.GetUser().ID,
				SubjectId:          &subId,
				Subject:            models.LogSubjectAbsent,
				SubjectDescription: &label,
				SubjectAction:      subAct,
				SubjectProperties:  data.Absent,
			})
			sp.End()
		}
	}

	// return journal item
	res := models.JournalItemResponse{
		Lesson:  models.LessonResponse{},
		Grades:  []models.GradeResponse{},
		Absents: []models.AbsentResponse{},
	}
	res.Lesson.FromModel(&lesson)
	if assignment.ID != "" {
		res.Lesson.Assignment = &models.AssignmentResponse{}
		res.Lesson.Assignment.FromModel(&assignment)
	}
	for _, model := range grades {
		resGrade := models.GradeResponse{}
		resGrade.FromModel(model)
		res.Grades = append(res.Grades, resGrade)
	}
	for _, model := range absents {
		resAbsent := models.AbsentResponse{}
		resAbsent.FromModel(model)
		res.Absents = append(res.Absents, resAbsent)
	}
	for _, model := range studentNotes {
		resStudentNote := models.StudentNoteResponse{}
		resStudentNote.FromModel(&model)
	}
	resPeriodGrades := []models.PeriodGradeResponse{}
	for _, item := range periodGrades {
		resPeriodGrade := models.PeriodGradeResponse{}
		resPeriodGrade.FromModel(item)
		resPeriodGrade.SetValueByRules()
		resPeriodGrades = append(resPeriodGrades, resPeriodGrade)
	}
	return res, resPeriodGrades, nil
}

func (a App) LessonUpdateV2(ses *utils.Session, data *models.JournalRequestV2) (models.JournalItemResponse, []models.PeriodGradeResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "LessonUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	// todo: check student
	// update or create lesson
	var err error
	lesson := models.Lesson{}
	data.SetKeys()
	err = lesson.FromRequest(&data.Lesson)
	if err != nil {
		return models.JournalItemResponse{}, nil, err
	}

	lesson, err = store.Store().LessonsUpdate(ses.Context(), lesson)
	if err != nil {
		return models.JournalItemResponse{}, nil, err
	}
	err = store.Store().LessonsLoadRelations(ctx, &[]*models.Lesson{&lesson})
	if err != nil {
		return models.JournalItemResponse{}, nil, err
	}
	lessonDesc := ""
	if lesson.Subject != nil {
		if lesson.Subject.Classroom != nil {
			lessonDesc = *lesson.Subject.Name + " " + *lesson.Subject.Classroom.Name + " " + lesson.Date.Format(time.DateOnly)
		} else {
			lessonDesc = *lesson.Subject.Name + " ? " + lesson.Date.Format(time.DateOnly)
		}
	}

	// update assignment
	assignment := models.Assignment{}
	assignment.FromRequest(&data.Lesson.Assignment)

	// update student note
	studentNotes := []models.StudentNote{}
	if data.StudentNote != nil {
		data.StudentNote.SubjectId = &lesson.SubjectId
		data.StudentNote.SchoolId = lesson.SchoolId
		data.StudentNote.TeacherId = ses.GetUser().ID
		studentNoteData := models.StudentNote{}
		studentNoteData.FromRequest(data.StudentNote)
		studentNote, err := lessonsStudentNotesUpdate(ses, studentNoteData)
		if err != nil {
			return models.JournalItemResponse{}, nil, err
		}
		studentNotes = append(studentNotes, studentNote)
	}

	// if both delete other
	if data.Grade != nil && !data.Grade.IsValueDelete() {
		data.Absent = &models.AbsentRequest{
			StudentId: data.Grade.StudentId,
			UpdatedBy: data.Grade.UpdatedBy,
		}
		data.Absent.Reason = nil
	}
	if data.Absent != nil && !data.Absent.IsValueDelete() {
		data.Grade = &models.GradeRequest{
			StudentId: data.Absent.StudentId,
			UpdatedBy: data.Absent.UpdatedBy,
		}
		data.Grade.Value = nil
		data.Grade.Values = nil
	}

	// update grade
	grades := []*models.Grade{}
	periodGrades := []*models.PeriodGrade{}
	if data.Grade != nil {
		grades, periodGrades, err = gradeMake(ses, data.Grade, lesson)
		if err != nil {
			return models.JournalItemResponse{}, nil, err
		}

		subId := ""
		subAct := models.LogActionCreate
		if len(grades) > 0 {
			subId = grades[0].ID
		}
		if data.Grade.IsValueDelete() {
			subAct = models.LogActionDelete
			if subId == "" {
				subAct = ""
			}
		}
		if subAct != "" {
			userLog(ses, models.UserLog{
				SchoolId:           ses.GetSchoolId(),
				SessionId:          ses.GetSessionId(),
				UserId:             ses.GetUser().ID,
				SubjectId:          &subId,
				Subject:            models.LogSubjectGrade,
				SubjectDescription: &lessonDesc,
				SubjectAction:      subAct,
				SubjectProperties:  data.Grade,
			})
		}
	}

	// update absent
	absents := []*models.Absent{}
	if data.Absent != nil {
		absents, periodGrades, err = absentMake(ses, data.Absent, lesson)
		if err != nil {
			return models.JournalItemResponse{}, nil, err
		}

		subId := ""
		subAct := models.LogActionCreate
		if len(absents) > 0 {
			subId = absents[0].ID
		}
		if data.Absent.IsValueDelete() {
			subAct = models.LogActionDelete
			if subId == "" {
				subAct = ""
			}
		}

		if subAct != "" {
			sp, ctx := apm.StartSpan(ctx, "userLog", "app")
			ses.SetContext(ctx)
			userLog(ses, models.UserLog{
				SchoolId:           ses.GetSchoolId(),
				SessionId:          ses.GetSessionId(),
				UserId:             ses.GetUser().ID,
				SubjectId:          &subId,
				Subject:            models.LogSubjectAbsent,
				SubjectDescription: &lessonDesc,
				SubjectAction:      subAct,
				SubjectProperties:  data.Absent,
			})
			sp.End()
		}
	}
	// return journal item
	res := models.JournalItemResponse{
		Lesson:  models.LessonResponse{},
		Grades:  []models.GradeResponse{},
		Absents: []models.AbsentResponse{},
	}
	res.Lesson.FromModel(&lesson)
	if assignment.ID != "" {
		res.Lesson.Assignment = &models.AssignmentResponse{}
		res.Lesson.Assignment.FromModel(&assignment)
	}
	for _, model := range grades {
		resGrade := models.GradeResponse{}
		resGrade.FromModel(model)
		res.Grades = append(res.Grades, resGrade)
	}
	for _, model := range absents {
		resAbsent := models.AbsentResponse{}
		resAbsent.FromModel(model)
		res.Absents = append(res.Absents, resAbsent)
	}
	for _, model := range studentNotes {
		resStudentNote := models.StudentNoteResponse{}
		resStudentNote.FromModel(&model)
	}
	resPeriodGrades := []models.PeriodGradeResponse{}
	for _, item := range periodGrades {
		resPeriodGrade := models.PeriodGradeResponse{}
		resPeriodGrade.FromModel(item)
		resPeriodGrade.SetValueByRules()
		resPeriodGrades = append(resPeriodGrades, resPeriodGrade)
	}
	return res, resPeriodGrades, nil
}

func lessonsStudentNotesUpdate(ses *utils.Session, ms models.StudentNote) (models.StudentNote, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "lessonsStudentNotesUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	ms, err := store.Store().StudentNotesUpdateOrCreate(ses.Context(), &ms)
	if err != nil {
		return models.StudentNote{}, err
	}
	return ms, nil
}

func lessonsDelete(ses *utils.Session, data models.Lesson) error {
	_, err := store.Store().LessonsDelete(ses.Context(), []*models.Lesson{&data})
	return err
}
