package app

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func ReportItemsCreate(ses *utils.Session, data models.ReportItemsRequest) (*models.ReportItemsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ReportItemsCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	res := &models.ReportItemsResponse{}
	var err error
	// find report item by id
	modelItem, err := store.Store().ReportItemsFindById(ses.Context(), *data.ID)
	if err != nil {
		return nil, err
	}
	if modelItem == nil {
		return &models.ReportItemsResponse{}, ErrNotfound.SetKey("id")
	}
	// check permission
	userRole := *ses.GetRole()
	if userRole == models.RolePrincipal && ses.GetSchoolId() != nil && modelItem.SchoolId != nil && *ses.GetSchoolId() != *modelItem.SchoolId {
		return &models.ReportItemsResponse{}, ErrNotfound.SetKey("id")
	}
	if userRole == models.RoleTeacher && ses.GetUser().TeacherClassroom.ID != *modelItem.ClassroomId {
		return &models.ReportItemsResponse{}, ErrNotfound.SetKey("id")
	}
	if userRole == models.RoleOrganization && ses.GetSchoolId() != modelItem.SchoolId {
		return &models.ReportItemsResponse{}, ErrNotfound.SetKey("id")
	}
	// update report item
	modelItem.UpdatedBy = &ses.GetUser().ID
	modelItem.Values = data.Values
	modelItem.IsEditedManually = data.IsEditedManually
	modelItem, err = store.Store().ReportItemsUpdate(ses.Context(), *modelItem)
	if err != nil {
		return nil, err
	}
	err = store.Store().ReportItemsLoadRelations(ses.Context(), &[]*models.ReportItems{modelItem})
	if err != nil {
		return nil, err
	}
	res.FromModel(modelItem)
	return res, nil
}
