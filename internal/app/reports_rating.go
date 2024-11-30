package app

import (
	"encoding/json"
	"errors"
	"log"
	"sort"
	"strconv"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func ReportsRatingCenter(ses *utils.Session, reportId string) (*models.ReportRating, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ReportsRatingCenter", "app")
	ses.SetContext(ctx)
	defer sp.End()
	// fetch
	report, err := store.Store().ReportsFindById(ses.Context(), reportId)
	if err != nil {
		return nil, err
	}
	f := models.ReportItemsFilterRequest{
		ReportId: &report.ID,
	}
	limit := 1000
	offset := 0
	f.Limit = &limit
	f.Offset = &offset
	reportItems, _, err := store.Store().ReportItemsFindBy(ses.Context(), f)
	if err != nil {
		return nil, err
	}
	err = store.Store().ReportItemsLoadRelations(ses.Context(), &reportItems)
	if err != nil {
		return nil, err
	}
	report.ReportItems = reportItems

	// rating
	type valueByIdType struct {
		Value      int
		ReportItem models.ReportItems
	}
	valuesById := []valueByIdType{}
	for _, v := range report.ReportItems {
		value, err := calculateRatingValue(*report, *v)
		if err != nil {
			log.Println(err)
			continue
		}
		valuesById = append(valuesById, valueByIdType{
			ReportItem: *v,
			Value:      value,
		})
	}
	sort.Slice(valuesById, func(i int, j int) bool {
		return valuesById[i].Value > valuesById[j].Value
	})
	reportRes := models.ReportsResponse{}
	reportRes.FromModel(report)
	rating := models.ReportRating{
		Report:           reportRes,
		ReportRatingList: []models.ReportRatingItem{},
	}
	for k, v := range valuesById {
		resItem := models.ReportItemsResponse{}
		resItem.FromModel(&v.ReportItem)
		rating.ReportRatingList = append(rating.ReportRatingList, models.ReportRatingItem{
			ReportItems: &resItem,
			Value:       v.Value,
			Index:       k + 1,
		})
	}

	return &rating, err
}

func calculateRatingValue(report models.Reports, item models.ReportItems) (int, error) {
	sumPoint := 0
	for _, v := range models.DefaultRatingReports {
		num, err := getReportItemNumber(report, item, string(v.Key))
		if err != nil {
			continue
		}
		if v.CalcPoint != nil {
			sumPoint += v.CalcPoint(num)
		}
	}

	pointStudents, _ := getReportItemNumber(report, item, string(models.ReportKeySeasonStudents))
	pointStudentsCompleted, _ := getReportItemNumber(report, item, string(models.ReportKeySeasonStudentsCompleted))
	if pointStudentsCompleted > 0 && pointStudents > 0 {
		sumPoint += int(float64(pointStudentsCompleted) / float64(pointStudents) * 10)
	}

	return sumPoint, nil
}

func getReportItemValue(report models.Reports, item models.ReportItems, key string) (string, models.ReportValueType, error) {
	valueIndex := -1
	valueType := models.ReportValueType("")
	value := ""

	for k, v := range report.ValueTypes {
		if v.Key != nil && *v.Key == key {
			valueIndex = k
			valueType = v.Type
			break
		}
	}
	if valueIndex < 0 {
		return value, valueType, errors.New("key not found in values: " + key)
	}
	if len(item.Values) <= valueIndex {
		return value, valueType, errors.New("value types or keys not found: " + key + "  " + strconv.Itoa(valueIndex))
	}
	valuePtr := item.Values[valueIndex]
	if valuePtr != nil {
		value = *valuePtr
	}
	return value, valueType, nil
}

func getReportItemNumber(report models.Reports, item models.ReportItems, key string) (int, error) {
	value, valueType, err := getReportItemValue(report, item, key)
	if err != nil {
		return 0, err
	}
	// empty
	if value == "" {
		return 0, nil
	}
	// type is number or list
	if valueType == models.ReportValueTypeNumber {
		return strconv.Atoi(value)
	}
	if valueType == models.ReportValueTypeList {
		l := []string{}
		json.Unmarshal([]byte(value), l)
		return len(l), nil
	}
	// none type match
	return 0, nil
}

func getReportItemList(item models.ReportItems, key string) (int, error) {

	return 0, nil
}
