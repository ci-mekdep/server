package app

import (
	"context"
	"errors"
	"slices"
	"time"

	apiUtils "github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func PaymentTransactionsList(ses *apiUtils.Session, dto models.PaymentTransactionFilterRequest) ([]*models.PaymentTransactionResponse, int, map[string]int, map[string]int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "PaymentTransactionsList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	payments, total, totalAmount, totalTransactions, err := store.Store().PaymentTransactionsFindBy(ses.Context(), dto)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	err = store.Store().PaymentTransactionsLoadRelations(ses.Context(), &payments)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	res := []*models.PaymentTransactionResponse{}
	for _, m := range payments {
		item := models.PaymentTransactionResponse{}
		item.FromModel(m)
		res = append(res, &item)
	}
	return res, total, totalAmount, totalTransactions, nil
}

func PaymentTransactionDetail(ses *apiUtils.Session, id string) (*models.PaymentTransactionResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "PaymentTransactionDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	m, err := store.Store().PaymentTransactionsFindById(ses.Context(), id)
	if err != nil {
		return nil, err
	}
	err = store.Store().PaymentTransactionsLoadRelations(ses.Context(), &[]*models.PaymentTransaction{m})
	if err != nil {
		return nil, err
	}
	res := &models.PaymentTransactionResponse{}
	res.FromModel(m)
	return res, nil
}

func (App) PaymentTransactions(data models.PaymentTransactionFilterRequest) ([]models.PaymentTransactionResponse, int, error) {
	l, total, _, _, err := store.Store().PaymentTransactionsFindBy(context.Background(), data)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().PaymentTransactionsLoadRelations(context.Background(), &l)
	if err != nil {
		return nil, 0, err
	}
	res := []models.PaymentTransactionResponse{}
	for _, m := range l {
		item := models.PaymentTransactionResponse{}
		item.FromModel(m)
		res = append(res, item)
	}
	return res, total, nil
}

func paymentAvailableChildIds(ses *apiUtils.Session) []string {
	ids := []string{}
	sessionUser, _, err := store.Store().UsersFindBy(ses.Context(), models.UserFilterRequest{
		ID: &ses.GetUser().ID,
	})
	if err != nil {
		return nil
	}
	err = store.Store().UsersLoadRelations(ses.Context(), &sessionUser, true)
	if err != nil {
		return nil
	}
	for _, child := range sessionUser[0].Children {
		ids = append(ids, child.ID)
	}
	if ses.GetRole() != nil && *ses.GetRole() == models.RoleStudent {
		ids = append(ids, ses.GetUser().ID)
	}
	return ids
}

func paymentAvailableClassromIds(ses *apiUtils.Session) []string {
	ids := []string{}
	user := ses.GetUser()
	err := store.Store().UsersLoadRelationsChildren(ses.Context(), &[]*models.User{user})
	if err != nil {
		return nil
	}
	err = store.Store().UsersLoadRelationsClassrooms(ses.Context(), &user.Children)
	if err != nil {
		return nil
	}
	for _, child := range user.Children {
		for _, classroom := range child.Classrooms {
			ids = append(ids, classroom.ClassroomId)
		}
	}
	if ses.GetRole() != nil && *ses.GetRole() == models.RoleStudent {
		for _, classroom := range user.Classrooms {
			ids = append(ids, classroom.ClassroomId)
		}
	}
	return ids
}

func PaymentCalculationGet(ses *apiUtils.Session) (map[string]map[string]*models.PaymentTariff, error) {
	// Get available child_ids
	availableChildIds := paymentAvailableChildIds(ses)

	// Response structure
	response := map[string]map[string]*models.PaymentTariff{
		string(models.PaymentPlus):  {"school": nil, "center": nil},
		string(models.PaymentTrial): {"school": nil, "center": nil},
	}

	// Default tariff plus for school
	var defaultPlusTariff models.PaymentTariff
	for _, v := range models.DefaultTariff {
		if v.Code == models.PaymentPlus {
			defaultPlusTariff = v
			break
		}
	}

	// Free trial
	trialChildIds := []string{}
	for _, id := range availableChildIds {
		tt := string(models.PaymentTrial)
		st := string(models.PaymentStatusCompleted)
		_, tCount, _, _, err := store.Store().PaymentTransactionsFindBy(ses.Context(), models.PaymentTransactionFilterRequest{
			StudentId:  &id,
			TariffType: &tt,
			Status:     &st,
		})
		if err != nil {
			return nil, err
		}
		if tCount == 0 {
			trialChildIds = append(trialChildIds, id)
		}
	}

	if len(trialChildIds) > 0 {
		var defaultFreeTariff models.PaymentTariff
		for _, v := range models.DefaultTariff {
			if v.Code == models.PaymentTrial {
				defaultFreeTariff = v
				break
			}
		}
		schoolTrialTariff := defaultFreeTariff
		schoolTrialTariff.AvailableChildIDs = trialChildIds
		response["trial"]["school"] = &schoolTrialTariff

		centerTrialTariff := defaultFreeTariff
		centerTrialTariff.AvailableChildIDs = trialChildIds
		response["trial"]["center"] = &centerTrialTariff
	}

	// Assigning school tariff
	schoolPlusTariff := defaultPlusTariff
	response["plus"]["school"] = &schoolPlusTariff

	// Assigning center tariff
	centerPlusTariff := defaultPlusTariff
	centerPlusTariff.MonthPrice = map[int]float64{3: centerPlusTariff.Price * 3}
	response["plus"]["center"] = &centerPlusTariff

	return response, nil
}

func (App) PaymentCheckout(ses *apiUtils.Session, payer *models.User, data models.PaymentCheckoutRequest) (models.PaymentTransactionResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "PaymentCheckout", "app")
	ses.SetContext(ctx)
	defer sp.End()
	res := models.PaymentTransactionResponse{}
	err := store.Store().UsersLoadRelationsChildren(ses.Context(), &[]*models.User{payer})
	if err != nil {
		return models.PaymentTransactionResponse{}, err
	}
	// validate
	children := []*models.User{}
	childIds := []string{}
	for _, v := range data.StudentIds {
		for _, vv := range payer.Children {
			if v == vv.ID {
				children = append(children, vv)
				childIds = append(childIds, vv.ID)
			}
		}
		if payer.ID == v {
			childIds = append(childIds, payer.ID)
			children = append(children, payer)
		}
	}
	if len(children) < 1 {
		return res, ErrRequired.SetKey("student_ids")
	}
	if data.TariffType == "" {
		return res, ErrRequired.SetKey("tariff_type")
	}
	if data.BankType == "" {
		return res, ErrRequired.SetKey("bank_type")
	}

	// transaction args
	schoolId := ""
	for _, v := range payer.Schools {
		if v.SchoolUid != nil {
			schoolId = *v.SchoolUid
			break
		}
	}

	if data.TariffType == models.PaymentTrial {
		if len(data.SchoolClassroomIds) > 1 || len(data.CenterClassroomIds) > 1 {
			return models.PaymentTransactionResponse{}, ErrInvalid.SetKey("school_classroom_id").SetComment("available for 1 classroom")
		}
	}
	if data.SchoolMonths != 0 && len(data.SchoolClassroomIds) < 1 {
		data.SchoolMonths = 0
	} else if data.CenterMonths != 0 && len(data.CenterClassroomIds) < 1 {
		data.CenterMonths = 0
	}

	calcResp, err := PaymentCalculationPost(ses, data)
	if err != nil {
		return res, err
	}
	var totalAmount float64
	var unitPrice int
	var schoolPrice int
	var centerPrice int
	var discountPrice int
	var usedDaysPrice int
	if calcResp.Total != "" {
		totalAmount, _ = calcResp.Total.ToFloat64()
		unitPrice = calcResp.UnitPrice.ToInt()
		schoolPrice = calcResp.SchoolPrice.ToInt()
		centerPrice = calcResp.CenterPrice.ToInt()
		discountPrice = calcResp.DiscountPrice.ToInt()
		usedDaysPrice = calcResp.UsedDaysPrice.ToInt()
	}
	if isPhoneDebug(payer, false) {
		totalAmount = 0.1
	}
	m, err := paymentHandle(models.PaymentTransaction{
		SchoolId:           schoolId,
		PayerId:            payer.ID,
		SchoolClassroomIds: data.SchoolClassroomIds,
		CenterClassroomIds: data.CenterClassroomIds,
		UserIds:            childIds,
		TariffType:         data.TariffType,
		SchoolMonths:       data.SchoolMonths,
		CenterMonths:       data.CenterMonths,
		BankType:           data.BankType,
		CardName:           &data.CardName,
		Amount:             totalAmount,
		OriginalAmount:     totalAmount,
		UnitPrice:          &unitPrice,
		SchoolPrice:        &schoolPrice,
		CenterPrice:        &centerPrice,
		DiscountPrice:      &discountPrice,
		UsedDays:           &calcResp.UsedDays,
		UsedDaysPrice:      &usedDaysPrice,
	})
	if err != nil {
		return res, err
	}
	err = store.Store().PaymentTransactionsLoadRelations(ses.Context(), &[]*models.PaymentTransaction{m})
	if err != nil {
		return res, err
	}

	res.FromModel(m)

	return res, nil
}

func PaymentCalculationPost(ses *apiUtils.Session, data models.PaymentCheckoutRequest) (models.PaymentCalculationResponse, error) {
	tariffs, _ := PaymentCalculationGet(ses)
	availableClassroomIds := paymentAvailableClassromIds(ses)
	centerMonths := []int{}
	schoolMonths := []int{}
	if centerTariff, ok := tariffs[string(models.PaymentPlus)]["center"]; ok && centerTariff != nil {
		for month := range centerTariff.MonthPrice {
			centerMonths = append(centerMonths, month)
		}
	}
	if schoolTariff, ok := tariffs[string(models.PaymentPlus)]["school"]; ok && schoolTariff != nil {
		for month := range schoolTariff.MonthPrice {
			schoolMonths = append(schoolMonths, month)
		}
	}
	if len(data.SchoolClassroomIds) == 0 && len(data.CenterClassroomIds) == 0 {
		return models.PaymentCalculationResponse{}, ErrRequired.SetKey("school_classroom_id").SetComment("classroom_id is required")
	}
	for _, id := range data.SchoolClassroomIds {
		if !slices.Contains(availableClassroomIds, id) {
			return models.PaymentCalculationResponse{}, ErrRequired.SetKey("school_classroom_id").SetComment("classroom_id is required")
		}
	}
	for _, id := range data.CenterClassroomIds {
		if !slices.Contains(availableClassroomIds, id) {
			return models.PaymentCalculationResponse{}, ErrRequired.SetKey("center_classroom_id").SetComment("classroom_id is required")
		}
	}

	if len(data.StudentIds) == 0 {
		return models.PaymentCalculationResponse{}, ErrRequired.SetKey("student_id").SetComment("student_id is required")
	}
	if data.CenterMonths == 0 && data.SchoolMonths == 0 {
		return models.PaymentCalculationResponse{}, ErrRequired.SetKey("school_months").SetComment("school_months or center_months is required")
	}
	if len(data.CenterClassroomIds) > 0 && !slices.Contains(centerMonths, data.CenterMonths) {
		return models.PaymentCalculationResponse{}, errors.New("center_months must be 3 if center_classroom_ids is provided")
	} else if len(data.SchoolClassroomIds) > 0 && !slices.Contains(schoolMonths, data.SchoolMonths) {
		return models.PaymentCalculationResponse{}, errors.New("school_months must be 1 or 9 if school_classroom_ids is provided")
	}

	defaultTariff := models.PaymentTariff{}
	for _, v := range models.DefaultTariff {
		if v.Code == models.PaymentTariffType(data.TariffType) {
			defaultTariff = v
			break
		}
	}

	// Default prices
	var schoolPrice, centerPrice, discountPrice, usedDaysPrice, total float64
	var usedDays int
	var isDateChanges bool

	classroomIds := append(data.SchoolClassroomIds, data.CenterClassroomIds...)
	argsClassroom := models.ClassroomFilterRequest{}
	// Process each classroom ID
	for _, classroomId := range classroomIds {
		argsClassroom = models.ClassroomFilterRequest{
			ID: &classroomId,
		}
	}
	classroom, _, err := store.Store().ClassroomsFindBy(ses.Context(), argsClassroom)
	if err != nil {
		return models.PaymentCalculationResponse{}, err
	}
	// Get CURRENT period by school_id
	argsPeriod := models.PeriodFilterRequest{
		SchoolId: &classroom[0].SchoolId,
	}
	period, _, err := store.Store().PeriodsListFilters(ses.Context(), argsPeriod)
	if err != nil {
		return models.PaymentCalculationResponse{}, nil
	}
	periodStart, periodEnd, err := period[0].Dates()
	if err != nil {
		return models.PaymentCalculationResponse{}, err
	}

	// Set start date
	today := time.Now()
	if today.Before(periodStart) {
		today = periodStart
		isDateChanges = true
	}

	// Calculate left days
	leftDays := int(periodEnd.Sub(today).Hours() / 24)
	if leftDays <= 14 {
		return models.PaymentCalculationResponse{}, ErrExpired.SetComment("less than 14 days left in the period")
	}

	// Calculate used days
	if data.SchoolMonths > 1 && data.CenterMonths > 1 {
		usedDays = int(today.Sub(periodStart).Hours() / 24)
	}

	// Calculate schoolPrice and usedDaysPrice based on months

	for _, classroomId := range data.SchoolClassroomIds {
		if len(data.SchoolClassroomIds) > 0 && classroomId != "" {
			itemPrice := defaultTariff.Price * float64(data.SchoolMonths)
			schoolPrice += itemPrice
			if discount, ok := defaultTariff.MonthPrice[data.SchoolMonths]; ok {
				if itemPrice > discount {
					discountPrice += discount - itemPrice
				}
			}
			dayPrice := defaultTariff.Price / 30
			usedDaysPrice -= float64(usedDays) * dayPrice
		}
	}

	// Calculate centerPrice and usedDaysPrice based on months
	for _, classroomId := range data.CenterClassroomIds {
		if len(data.CenterClassroomIds) > 0 && classroomId != "" {
			itemPrice := defaultTariff.Price * float64(data.CenterMonths)
			centerPrice += itemPrice

			if discount, ok := defaultTariff.MonthPrice[data.CenterMonths]; ok {
				if itemPrice > discount {
					discountPrice += discount - itemPrice
				}
			}
			dayPrice := defaultTariff.Price / 30
			if leftDays < 30 {
				usedDaysPrice -= float64(usedDays) * dayPrice
			}
		}
	}
	// Total calculation
	total = schoolPrice + centerPrice + discountPrice + usedDaysPrice

	return models.PaymentCalculationResponse{
		UnitPrice:     models.MoneyFromFloat(defaultTariff.Price),
		SchoolPrice:   models.MoneyFromFloat(schoolPrice),
		CenterPrice:   models.MoneyFromFloat(centerPrice),
		DiscountPrice: models.MoneyFromFloat(discountPrice),
		UsedDaysPrice: models.MoneyFromFloat(usedDaysPrice),
		UsedDays:      usedDays,
		Total:         models.MoneyFromFloat(total),
		IsDateChange:  isDateChanges,
	}, nil
}

func paymentHandle(data models.PaymentTransaction) (*models.PaymentTransaction, error) {
	data.Status = models.PaymentStatusProcess
	if isPaymentHandleSpecialTariff(&data) {
		data.Amount = 0
		data.OriginalAmount = 0
	}
	m, err := store.Store().PaymentTransactionCreate(context.Background(), &data)
	if err != nil {
		return nil, err
	}

	err = store.Store().PaymentTransactionsLoadRelations(context.Background(), &[]*models.PaymentTransaction{m})
	if err != nil {
		return nil, err
	}

	if isPaymentHandleSpecialTariff(m) {
		return paymentHandleSpecialTariff(m)
	}

	err = PaymentHandleBankApi(m)
	if err != nil {
		m.Status = models.PaymentStatusFailed
		m.SystemComment = new(string)
		*m.SystemComment = err.Error()
		_, err1 := store.Store().PaymentTransactionUpdate(context.Background(), m)
		if err1 != nil {
			return nil, err1
		}
		return nil, err
	}

	m, err = store.Store().PaymentTransactionUpdate(context.Background(), m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func isPaymentHandleSpecialTariff(m *models.PaymentTransaction) bool {
	return slices.Contains([]models.PaymentTariffType{models.PaymentTrial}, m.TariffType)
}

func paymentHandleSpecialTariff(m *models.PaymentTransaction) (*models.PaymentTransaction, error) {
	if m.TariffType == models.PaymentTrial {
		m.SchoolMonths = 0
		m.CenterMonths = 0
		isOk := true
		for _, studentId := range m.UserIds {
			tt := string(models.PaymentTrial)
			st := string(models.PaymentStatusCompleted)
			_, tCount, _, _, err := store.Store().PaymentTransactionsFindBy(context.Background(), models.PaymentTransactionFilterRequest{
				StudentId:  &studentId,
				TariffType: &tt,
				Status:     &st,
			})
			if err != nil {
				return nil, err
			}
			if tCount > 0 {
				isOk = false
				break
			}
		}
		if !isOk {
			return nil, ErrExpired.SetComment("Siz üçin elýeter däl")
		}

		m.Status = models.PaymentStatusCompleted
		_, err := store.Store().PaymentTransactionUpdate(context.Background(), m)
		if err != nil {
			return nil, err
		}
		// upgrade
		for _, v := range m.Students {
			err = UserTariffUpgrade(&apiUtils.Session{}, m, v, []*models.User{m.Payer})
			if err != nil {
				return nil, err
			}
		}
	}
	return m, nil
}
