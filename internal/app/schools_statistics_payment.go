package app

import (
	"strconv"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func StatisticsPayments(ses *utils.Session, dto models.PaymentTransactionFilterRequest) (StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsStudents", "app")
	ses.SetContext(ctx)
	defer sp.End()
	args := models.SchoolFilterRequest{}
	args.IsParent = new(bool)
	*args.IsParent = false
	args.Limit = new(int)
	*args.Limit = 500
	schools, _, err := store.Store().SchoolsFindBy(ses.Context(), args)
	if err != nil {
		return StatisticsResponse{}, err
	}
	err = store.Store().SchoolsLoadRelations(ses.Context(), &schools)
	if err != nil {
		return StatisticsResponse{}, err
	}
	paymentCounts, err := store.Store().PaymentsTransactionsCountBySchool(ses.Context(), dto)
	if err != nil {
		return StatisticsResponse{}, err
	}
	resRows := []*StatisticsRow{}
	for _, schoolItem := range schools {
		rowItem := StatisticsRow{}
		resRows = append(resRows, &rowItem)
		rowItem.FromSchool(schoolItem)
		var totalCount, altynAsyrCount, senagatCount, rysgalCount, tfebCount int
		for _, payment := range paymentCounts {
			if payment.SchoolCode == *schoolItem.Code {
				totalCount = payment.TotalCount
				altynAsyrCount = payment.HalkbankCount
				senagatCount = payment.SenagatbankCount
				rysgalCount = payment.RysgalbankCount
				tfebCount = payment.TfebCount
			}
		}
		rowItem.Values = []StatisticsCell{
			StatisticsCell(*schoolItem.Code),
			StatisticsCell(*schoolItem.Name),
			StatisticsCell(strconv.Itoa(totalCount)),
			StatisticsCell(strconv.Itoa(altynAsyrCount)),
			StatisticsCell(strconv.Itoa(rysgalCount)),
			StatisticsCell(strconv.Itoa(senagatCount)),
			StatisticsCell(strconv.Itoa(tfebCount)),
		}
	}
	resRowsNoPtr := []StatisticsRow{}
	for _, v := range resRows {
		resRowsNoPtr = append(resRowsNoPtr, *v)
	}

	res := StatisticsResponse{
		Headers: []StatisticsHeader{"Kody", "Mekdep", "Jemi sany", "Altyn Asyr Bank#halkbank", "Rysgal Bank#rysgalbank", "Senagat Bank#senagatbank", "TFEB#tfeb"},
		Rows:    resRowsNoPtr,
	}
	return res, nil
}
