package app

import (
	"strconv"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func StatisticsContactItems(ses *utils.Session, dto models.ContactItemsFilterRequest) (StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsStudents", "app")
	ses.SetContext(ctx)
	defer sp.End()
	argsS := models.SchoolFilterRequest{}
	argsS.Limit = new(int)
	*argsS.Limit = 500
	schools, _, err := store.Store().SchoolsFindBy(ses.Context(), argsS)
	if err != nil {
		return StatisticsResponse{}, err
	}
	err = store.Store().SchoolsLoadRelations(ses.Context(), &schools)
	if err != nil {
		return StatisticsResponse{}, err
	}
	contactItemsCount, err := store.Store().ContactItemsCountByType(ses.Context(), dto)
	if err != nil {
		return StatisticsResponse{}, err
	}
	resRows := []*StatisticsRow{}
	resRowsNoPtr := []StatisticsRow{}
	for _, schoolItem := range schools {
		rowItem := StatisticsRow{}
		resRows = append(resRows, &rowItem)
		rowItem.FromSchool(schoolItem)
		var totalCount, reviewCount, complaintCount, suggestionCount, dataComplaintCount int
		for _, contactItem := range contactItemsCount {
			if contactItem.SchoolCode == *schoolItem.Code {
				totalCount = contactItem.TotalCount
				reviewCount = contactItem.ReviewCount
				complaintCount = contactItem.ComplaintCount
				suggestionCount = contactItem.SuggestionCount
				dataComplaintCount = contactItem.DataComplaintCount
			}
		}
		rowItem.Values = []StatisticsCell{
			StatisticsCell(*schoolItem.Code),
			StatisticsCell(*schoolItem.Name),
			StatisticsCell(strconv.Itoa(totalCount)),
			StatisticsCell(strconv.Itoa(reviewCount)),
			StatisticsCell(strconv.Itoa(complaintCount)),
			StatisticsCell(strconv.Itoa(suggestionCount)),
			StatisticsCell(strconv.Itoa(dataComplaintCount)),
		}
	}
	for _, v := range resRows {
		resRowsNoPtr = append(resRowsNoPtr, *v)
	}

	res := StatisticsResponse{
		Headers: []StatisticsHeader{"Kody", "Mekdep", "Jemi sany", "Teswir#review", "Arz-şikaýat#complaint", "Teklip#suggestion", "Maglumaty düzediş arza#data_complaint"},
		Rows:    resRowsNoPtr,
	}
	return res, nil
}
