package app

import (
	"errors"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	apputils "github.com/mekdep/server/internal/utils"
	"go.elastic.co/apm/v2"
)

func UserTariffUpgrade(ses *utils.Session, payment *models.PaymentTransaction, child models.User, parents []*models.User) (err error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UserTariffUpgrade", "app")
	ses.SetContext(ctx)
	defer sp.End()

	err = store.Store().UsersLoadRelationsClassrooms(ses.Context(), &[]*models.User{&child})
	if err != nil {
		return err
	}

	classroomIds := append(payment.CenterClassroomIds, payment.SchoolClassroomIds...)
	if classroomIds == nil {
		classroomIds = []string{child.Classrooms[0].ClassroomId}
	}
	for _, v := range parents {
		UserCacheClear(ses, v.ID)
	}

	for _, classroomId := range classroomIds {
		apputils.Logger.Info("PAID class: " + classroomId + " student: " + child.ID)
		days := payment.SchoolMonths * 30
		daysCenter := payment.CenterMonths * 30
		// TODO: get user_classroom
		expiresAt, err := store.Store().GetDateUserPayment(ses.Context(), child.ID, classroomId)
		if err != nil {
			return err
		}
		// TODO: add time, and sql update user_classroom
		if expiresAt.IsZero() || expiresAt.Before(time.Now()) {
			expiresAt = time.Now()
		}
		if isPaymentHandleSpecialTariff(payment) {
			expiresAt = expiresAt.AddDate(0, 0, 14)
		} else if days != 0 {
			expiresAt = expiresAt.AddDate(0, 0, days)
		} else if daysCenter != 0 {
			expiresAt = expiresAt.AddDate(0, 0, daysCenter)
		} else {
			err = errors.New("No days provided nor special tariff")
			apputils.LoggerDesc("error bank api checkout").Error(err)
			return err
		}
		_, err = store.Store().UpdateUserPayment(ses.Context(), child.ID, expiresAt, classroomId)
		if err != nil {
			return err
		}
		err = SendWelcomeSms(&child, parents, expiresAt)
		if err != nil {
			return err
		}
	}
	return nil
}
