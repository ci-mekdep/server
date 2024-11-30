package cmd

import (
	"context"
	"time"

	apiutils "github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"github.com/mekdep/server/internal/utils"
)

var SendDailySmsLastRun time.Time

// Caganyz Meret eMekdep programmada nyrhnamasynyň gutarmagyna 7 gun galdy. Goşmaça mümkinçilikleri dowam etmek üçin töleg etmeli.
func SendTariffEndsSmsAll(isLate bool) error {
	today := time.Now()
	todayFormatted := today.Format("2006-01-02")
	args := models.UserFilterRequest{
		TariffEndMin: new(time.Time),
		Role:         new(string),
	}
	*args.TariffEndMin = today
	*args.Role = string(models.RoleStudent)
	args.Limit = new(int)
	args.Offset = new(int)
	*args.Limit = 500
	*args.Offset = 0

	uu, total, err := store.Store().UsersFindBy(context.Background(), args)
	if err != nil {
		return err
	}
	err = store.Store().UsersLoadRelationsParents(context.Background(), &uu)
	if err != nil {
		return err
	}
	err = store.Store().UsersLoadRelationsClassrooms(context.Background(), &uu)
	if err != nil {
		return err
	}

	for total > 0 {
		for _, u := range uu {
			for _, c := range u.Classrooms {
				if c.TariffEndAt != nil && today.Before(*c.TariffEndAt) {
					if c.TariffType != nil && (*c.TariffType == string(models.PaymentPlus) || *c.TariffType == string(models.PaymentUnlimited)) {
						if len(u.Parents) < 1 {
							utils.LoggerDesc("SendTariffEndsSmsAll").Info("User has no parents, ID: " + u.ID)
						}
						tariffEndMinus7Days := c.TariffEndAt.AddDate(0, 0, -7).Format("2006-01-02")
						tariffEndMinus1Day := c.TariffEndAt.AddDate(0, 0, -1).Format("2006-01-02")
						if todayFormatted == tariffEndMinus7Days {
							err := app.SendTariffEndsSmsAll(&apiutils.Session{}, isLate, today, u, u.Parents, "7")
							if err != nil {
								utils.LoggerDesc("SendTariffEndsSmsAll").Error(err)
							}
						} else if todayFormatted == tariffEndMinus1Day {
							err := app.SendTariffEndsSmsAll(&apiutils.Session{}, isLate, today, u, u.Parents, "1")
							if err != nil {
								utils.LoggerDesc("SendTariffEndsSmsAll").Error(err)
							}
						}
					}
				}
			}
		}
		total -= *args.Limit
		*args.Offset += *args.Limit
		uu, _, err = store.Store().UsersFindBy(context.Background(), args)
		if err != nil {
			return err
		}
		err = store.Store().UsersLoadRelationsParents(context.Background(), &uu)
		if err != nil {
			return err
		}
		err = store.Store().UsersLoadRelationsClassrooms(context.Background(), &uu)
		if err != nil {
			return err
		}
	}

	return nil
}

func SendDailySms(isLate bool) error {
	today := time.Now()
	args := models.UserFilterRequest{
		TariffEndMin: new(time.Time),
		Role:         new(string),
	}
	*args.TariffEndMin = today
	*args.Role = string(models.RoleStudent)
	args.Limit = new(int)
	args.Offset = new(int)
	*args.Limit = 500
	*args.Offset = 0

	uu, total, err := store.Store().UsersFindBy(context.Background(), args)
	if err != nil {
		return err
	}
	err = store.Store().UsersLoadRelationsParents(context.Background(), &uu)
	if err != nil {
		return err
	}
	err = store.Store().UsersLoadRelationsClassrooms(context.Background(), &uu)
	if err != nil {
		return err
	}
	for total > 0 {
		for _, u := range uu {
			for _, c := range u.Classrooms {
				if c.TariffEndAt != nil && today.Before(*c.TariffEndAt) {
					if c.TariffType != nil && (*c.TariffType == string(models.PaymentPlus) || *c.TariffType == string(models.PaymentUnlimited)) {
						if len(u.Parents) < 1 {
							utils.LoggerDesc("SendDailySms").Info("User has no parents, ID: " + u.ID)
						}
						err := app.SendDailySms(&apiutils.Session{}, isLate, today, u, u.Parents)
						if err != nil {
							utils.LoggerDesc("SendSmsItem").Error(err)
						}
					}
				}
			}
		}
		total -= *args.Limit
		*args.Offset += *args.Limit
		uu, _, err = store.Store().UsersFindBy(context.Background(), args)
		if err != nil {
			return err
		}
		err = store.Store().UsersLoadRelationsParents(context.Background(), &uu)
		if err != nil {
			return err
		}
		err = store.Store().UsersLoadRelationsClassrooms(context.Background(), &uu)
		if err != nil {
			return err
		}
	}

	return nil
}
