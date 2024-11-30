package app

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	apputils "github.com/mekdep/server/internal/utils"
	"go.elastic.co/apm/v2"
)

// TODO: report list-da admin report_item gelmegi, schools-lary sortlamak, list-da cislo boyunca sortlamak
// TODO: env config add
// TODO: report create-da report-item dorande goroutine etmeli
func ReportsList(ses *utils.Session, f models.ReportsFilterRequest) ([]*models.ReportsResponse, int, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ReportsList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	totalUnfilled := 0
	// set more limit=100, because unfilled counted in App not Sql
	if f.Limit == nil {
		f.Limit = new(int)
		*f.Limit = 12
	}
	limit := *f.Limit
	*f.Limit = 100
	// fetch reports
	reports, total, err := store.Store().ReportsFindBy(ses.Context(), f)
	if err != nil {
		return nil, 0, 0, err
	}
	res := []*models.ReportsResponse{}
	for _, report := range reports {
		// fetch item
		if *ses.GetRole() != models.RoleAdmin {
			isFilled, err := reportItemLoadAndIsFilled(ses, report)
			if err != nil {
				return nil, 0, 0, err
			}
			if !isFilled {
				totalUnfilled++
			}
		}

		// set response
		resItem := models.ReportsResponse{}
		resItem.FromModel(report)
		res = append(res, &resItem)
	}
	if len(res) > limit {
		res = res[:limit]
	}
	return res, total, totalUnfilled, err
}

func reportItemLoadAndIsFilled(ses *utils.Session, report *models.Reports) (bool, error) {
	argsItem := models.ReportItemsFilterRequest{
		ReportId: &report.ID,
		SchoolId: ses.GetSchoolId(),
	}
	argsItem.Limit = new(int)
	*argsItem.Limit = 1000
	if *ses.GetRole() == models.RoleTeacher {
		classroomId := ""
		if ses.GetUser().TeacherClassroom != nil {
			classroomId = ses.GetUser().TeacherClassroom.ID
		}
		argsItem.ClassroomId = &classroomId
	}
	reportItems, _, err := store.Store().ReportItemsFindBy(ses.Context(), argsItem)
	if err != nil {
		return false, err
	}
	if len(reportItems) > 0 {
		if *ses.GetRole() == models.RolePrincipal {
			// Loop through reportItems to find one with classroom_id null and matching school_id
			for _, item := range reportItems {
				if item.ClassroomId == nil && item.SchoolId != nil && *item.SchoolId == *ses.GetSchoolId() {
					report.ReportItem = item
					break
				}
			}
		} else {
			report.ReportItem = reportItems[0]
		}

		// Check if ReportItem was found and if it has Values
		if report.ReportItem != nil {
			if report.ReportItem.Values == nil || len(report.ReportItem.Values) == 0 {
				return false, nil
			}
			return true, nil
		}
	}
	return false, nil
}

func ReportsDetail(ses *utils.Session, id string) (*models.ReportsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ReportsDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	// fetch report
	m, err := store.Store().ReportsFindById(ses.Context(), id)
	if err != nil {
		return nil, err
	}
	// fetch arg by permission
	f := models.ReportItemsFilterRequest{
		ReportId:  &m.ID,
		SchoolIds: ses.GetSchoolsByAdminRoles(),
	}
	f.OnlyClassroom = new(bool)
	*f.OnlyClassroom = false
	limit := 1000
	offset := 0
	f.Limit = &limit
	f.Offset = &offset
	if *ses.GetRole() == models.RolePrincipal {
		*f.OnlyClassroom = true
		f.SchoolId = ses.GetSchoolId()
	}
	if *ses.GetRole() == models.RoleTeacher {
		*f.OnlyClassroom = true
		classroomId := "null"
		if cl := ses.GetUser().TeacherClassroom; cl != nil {
			classroomId = cl.ID
		}
		f.ClassroomId = &classroomId
		f.SchoolId = ses.GetSchoolId()
	}
	// fetch report items
	reportItems, _, err := store.Store().ReportItemsFindBy(ses.Context(), f)
	if err != nil {
		return nil, err
	}
	err = store.Store().ReportItemsLoadRelations(ses.Context(), &reportItems)
	if err != nil {
		return nil, err
	}

	// fetch report item
	_, err = reportItemLoadAndIsFilled(ses, m)
	if err != nil {
		return nil, err
	}

	// counts and order by
	itemsCount := len(reportItems)
	itemsFilledCount := 0
	for _, reportItem := range reportItems {
		if reportItem.Values != nil && len(reportItem.Values) > 0 {
			itemsFilledCount++
		}
	}
	sort.SliceStable(reportItems, func(i, j int) bool {
		return reportItems[i].School.ParentUid == nil && reportItems[j].School.ParentUid != nil
	})

	// set response
	res := &models.ReportsResponse{}
	res.FromModel(m)
	res.ReportItems = []models.ReportItemsResponse{}
	for _, v := range reportItems {
		resItem := models.ReportItemsResponse{}
		resItem.FromModel(v)
		res.ReportItems = append(res.ReportItems, resItem)
	}
	res.ItemsCount = &itemsCount
	res.ItemsCountFilled = &itemsFilledCount
	return res, nil
}

func ReportsCreate(ses *utils.Session, data models.ReportsRequest) (*models.ReportsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ReportsCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	model := &models.Reports{}
	data.ToModel(model)
	if model.ValueTypes == nil || len(model.ValueTypes) < 1 {
		return nil, ErrRequired.SetKey("value_types")
	}
	var err error
	// if selected "All"
	if len(model.SchoolIds) == 1 && model.SchoolIds[0] == "" {
		allSchoolIds := ses.GetSchoolIds()
		model.SchoolIds = allSchoolIds
	}
	// filter regions to save ids
	regionSet := make(map[string]bool)
	rawSchools, err := store.Store().SchoolsFindByIds(ses.Context(), model.SchoolIds)
	if err != nil {
		return nil, err
	}
	schools := []*models.School{}
	for _, school := range rawSchools {
		if school.IsSecondarySchool == nil || !*school.IsSecondarySchool {
			continue
		}
		if school.ParentUid != nil {
			schools = append(schools, school)
			regionSet[*school.ParentUid] = true
		}
	}
	model.SchoolIds = nil
	for _, v := range schools {
		model.SchoolIds = append(model.SchoolIds, v.ID)
	}
	model.RegionIds = []string{}
	for regionId := range regionSet {
		model.RegionIds = append(model.RegionIds, regionId)
	}
	// create report with data
	model, err = store.Store().ReportsCreate(ses.Context(), model)
	if err != nil {
		return nil, err
	}
	go func() {
		reportItems := []models.ReportItems{}
		for _, schoolId := range model.SchoolIds {
			// fetch period uid
			if data.IsClassroomsIncluded != nil && *data.IsClassroomsIncluded {
				classroomUids, err := getClassroomIds(context.Background(), schoolId)
				if err != nil {
					apputils.Logger.Error(err)
				}
				for _, classroomId := range classroomUids {
					classroomIdd := classroomId
					reportItems = append(reportItems, models.ReportItems{
						ReportId:    &model.ID,
						SchoolId:    &schoolId,
						ClassroomId: &classroomIdd,
					})
				}
			}
			reportItems = append(reportItems, models.ReportItems{
				ReportId:    &model.ID,
				SchoolId:    &schoolId,
				PeriodId:    nil,
				ClassroomId: nil,
			})
			err = store.Store().ReportItemsCreateBatch(context.Background(), reportItems)
			if err != nil {
				apputils.Logger.Error(err)
			}
			reportItems = []models.ReportItems{}
		}
	}()

	reportItems := []models.ReportItems{}
	// Create report items for each region
	for _, regionId := range model.RegionIds {
		regionIdd := regionId
		reportItems = append(reportItems, models.ReportItems{
			ReportId: &model.ID,
			SchoolId: &regionIdd,
		})
	}
	err = store.Store().ReportItemsCreateBatch(ctx, reportItems)
	if err != nil {
		return nil, err
	}
	reportItems = []models.ReportItems{}

	res := &models.ReportsResponse{}
	res.FromModel(model)
	return res, nil
}

func ReportsUpdate(ses *utils.Session, data models.ReportsRequest) (*models.ReportsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ReportsUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()

	existingReport, err := store.Store().ReportsFindById(ses.Context(), *data.ID)
	if err != nil {
		return nil, err
	}
	data.IsClassroomsIncluded = existingReport.IsClassroomsIncluded
	data.ValueTypes = existingReport.ValueTypes
	for _, newSchoolId := range data.SchoolIds {
		found := false
		for _, existingsSchoolId := range existingReport.SchoolIds {
			if existingsSchoolId == newSchoolId {
				found = true
				continue
			}
		}
		if !found {
			existingReport.SchoolIds = append(existingReport.SchoolIds, newSchoolId)
			if data.IsClassroomsIncluded != nil && *data.IsClassroomsIncluded {
				classroomUids, err := getClassroomIds(ses.Context(), newSchoolId)
				if err != nil {
					return nil, err
				}
				for _, classroomId := range classroomUids {
					itemModel := models.ReportItems{
						ReportId:    data.ID,
						SchoolId:    &newSchoolId,
						ClassroomId: &classroomId,
					}
					_, err = store.Store().ReportItemsCreate(ses.Context(), itemModel)
					if err != nil {
						return nil, err
					}
					err = store.Store().ReportItemsLoadRelations(ses.Context(), &[]*models.ReportItems{&itemModel})
					if err != nil {
						return nil, err
					}
				}
			}
			itemModel := models.ReportItems{
				ReportId:    data.ID,
				SchoolId:    &newSchoolId,
				ClassroomId: nil,
			}
			_, err = store.Store().ReportItemsCreate(ses.Context(), itemModel)
			if err != nil {
				return nil, err
			}

			err = store.Store().ReportItemsLoadRelations(ses.Context(), &[]*models.ReportItems{&itemModel})
			if err != nil {
				return nil, err
			}
		}
	}
	existingReport.Description = data.Description
	existingReport.Title = data.Title
	existingReport.IsCenterRating = data.IsCenterRating
	existingReport, err = store.Store().ReportsUpdate(ses.Context(), existingReport)
	if err != nil {
		return nil, err
	}
	res := &models.ReportsResponse{}
	res.FromModel(existingReport)
	return res, nil
}

func ReportsDelete(ses *utils.Session, ids []string) ([]*models.Reports, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ReportsDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	reports, err := store.Store().ReportsFindByIds(ses.Context(), ids)
	if err != nil {
		return nil, err
	}
	if len(reports) < 1 {
		return nil, errors.New("model not found: " + strings.Join(ids, ","))
	}
	return store.Store().ReportsDelete(ses.Context(), reports)
}

func getClassroomIds(ctx context.Context, schoolId string) ([]string, error) {
	argsC := models.ClassroomFilterRequest{
		SchoolId: &schoolId,
	}
	argsC.Limit = new(int)
	argsC.Offset = new(int)
	*argsC.Limit = 1000
	*argsC.Offset = 0
	classrooms, _, err := store.Store().ClassroomsFindBy(ctx, argsC)
	if err != nil {
		return nil, err
	}
	var classroomUids []string
	for _, classroom := range classrooms {
		classroomUids = append(classroomUids, classroom.ID)
	}
	return classroomUids, nil
}
