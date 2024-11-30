package app

import (
	"context"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func TeacherExcuseList(ses *utils.Session, query models.TeacherExcuseQueryDto) (response *models.TeacherExcusesResponse, err error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TeacherExcuseList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	list, err := store.Store().TeacherExcusesFindBy(ses.Context(), models.ConvertTeacherExcuseQueryToMap(query))
	if err != nil {
		return nil, err
	}
	response = &models.TeacherExcusesResponse{
		TeacherExcuses: []models.TeacherExcuseResponse{},
	}
	response.Total = list.Total
	for _, v := range list.TeacherExcuses {
		response.TeacherExcuses = append(response.TeacherExcuses, models.ConvertTeacherExcuseToResponse(*v))
	}

	return
}

func TeacherExcuseDetail(ses *utils.Session, id string) (response *models.TeacherExcuseResponse, err error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TeacherExcuseDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	teacherExcuse, err := store.Store().TeacherExcusesFindById(ses.Context(), id, true)
	if err != nil {
		return nil, err
	}
	res := models.ConvertTeacherExcuseToResponse(*teacherExcuse)
	return &res, nil
}

func TeacherExcuseCreate(ses *utils.Session, query models.TeacherExcuseCreateDto) (teacherExcuse *models.TeacherExcuseResponse, err error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TeacherExcuseCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	now := time.Now()
	teacherExcuseCreate, err := store.Store().TeacherExcuseInsert(ses.Context(), &models.TeacherExcuse{
		TeacherId:     query.TeacherId,
		SchoolId:      *query.SchoolId,
		StartDate:     *query.StartDate,
		EndDate:       *query.EndDate,
		Reason:        query.Reason,
		Note:          query.Note,
		DocumentFiles: query.DocumentFiles,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	})
	if err != nil {
		return nil, err
	}
	limit := 100
	teacherExcusesList, err := store.Store().TeacherExcusesFindBy(ses.Context(), models.ConvertTeacherExcuseQueryToMap(models.TeacherExcuseQueryDto{
		TeacherId: &teacherExcuseCreate.TeacherId,
		Limit:     limit,
	}))
	err = syncTeacherExcuses(ses, teacherExcuseCreate.TeacherId, teacherExcusesList.TeacherExcuses)
	if err != nil {
		return nil, err
	}
	teacherExcuseRes := models.ConvertTeacherExcuseToResponse(*teacherExcuseCreate)
	teacherExcuse = &teacherExcuseRes
	return teacherExcuse, nil
}

func TeacherExcuseUpdate(ses *utils.Session, id string, query models.TeacherExcuseCreateDto) (teacherExcuse *models.TeacherExcuseResponse, err error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TeacherExcuseUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	now := time.Now()
	teacherExcuseUpdate, err := store.Store().TeacherExcuseUpdate(ses.Context(), &models.TeacherExcuse{
		ID:            id,
		TeacherId:     query.TeacherId,
		Note:          query.Note,
		DocumentFiles: query.DocumentFiles,
		UpdatedAt:     &now,
	})
	if err != nil {
		return nil, err
	}
	limit := 100
	teacherExcusesList, err := store.Store().TeacherExcusesFindBy(ses.Context(), models.ConvertTeacherExcuseQueryToMap(models.TeacherExcuseQueryDto{
		TeacherId: &teacherExcuseUpdate.TeacherId,
		Limit:     limit,
	}))
	err = syncTeacherExcuses(ses, teacherExcuseUpdate.TeacherId, teacherExcusesList.TeacherExcuses)
	if err != nil {
		return nil, err
	}

	resTeacherExcuse := models.ConvertTeacherExcuseToResponse(*teacherExcuseUpdate)
	teacherExcuse = &resTeacherExcuse
	return teacherExcuse, nil
}

func TeacherExcuseDelete(ses *utils.Session, ids []string) (resp *models.TeacherExcusesResponse, err error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TeacherExcuseDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	limit := 100
	teacherExcusesList, err := store.Store().TeacherExcusesFindBy(ses.Context(), models.ConvertTeacherExcuseQueryToMap(models.TeacherExcuseQueryDto{
		Ids:   ids,
		Limit: limit,
	}))
	for _, v := range teacherExcusesList.TeacherExcuses {
		err = syncTeacherExcuses(ses, v.TeacherId, teacherExcusesList.TeacherExcuses)
		if err != nil {
			return nil, err
		}
	}
	teacherExcuses, err := store.Store().TeacherExcusesDelete(ses.Context(), ids)
	if err != nil {
		return nil, err
	}
	resp = &models.TeacherExcusesResponse{
		TeacherExcuses: []models.TeacherExcuseResponse{},
	}
	for _, v := range teacherExcuses.TeacherExcuses {
		resp.TeacherExcuses = append(resp.TeacherExcuses, models.ConvertTeacherExcuseToResponse(*v))
	}
	return
}

func syncTeacherExcuses(ses *utils.Session, teacherId string, excuses []*models.TeacherExcuse) error {
	subjectFilter := models.SubjectFilterRequest{
		TeacherId: &teacherId,
		SchoolId:  ses.GetSchoolId(),
	}
	subjectFilter.Limit = new(int)
	*subjectFilter.Limit = 5000
	subjects, _, err := store.Store().SubjectsListFilters(context.Background(), &subjectFilter)
	if err != nil {
		return err
	}
	subjectIds := []string{}
	for _, v := range subjects {
		subjectIds = append(subjectIds, v.ID)
	}
	isTeacherExcused := true
	lessonFilter := models.LessonFilterRequest{
		SubjectIds:       &subjectIds,
		DateRange:        &[]string{time.Now().Format(time.DateOnly)},
		IsTeacherExcused: &isTeacherExcused,
	}
	_, err = store.Store().LessonsUpdateBy(context.Background(), lessonFilter, map[string]interface{}{
		"is_teacher_excused": false,
	})
	if err != nil {
		return err
	}

	// TODO: REFACTOR
	for _, v := range excuses {
		if (v.EndDate.Unix()-v.StartDate.Unix())/60/60/24 > 30 {
			continue
		}
		startDate := v.StartDate
		endDate := v.EndDate
		if startDate.Before(time.Now()) {
			startDate = time.Now()
		}
		isTeacherExcused = false
		lessonFilter = models.LessonFilterRequest{
			SubjectIds:       &subjectIds,
			DateRange:        &[]string{startDate.Format(time.DateOnly), endDate.Format(time.DateOnly)},
			IsTeacherExcused: &isTeacherExcused,
		}
		_, err = store.Store().LessonsUpdateBy(context.Background(), lessonFilter, map[string]interface{}{
			"is_teacher_excused": true,
		})
	}
	return err
}
