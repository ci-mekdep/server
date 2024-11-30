package cmd

import (
	"context"
	"log"
	"time"

	apiutils "github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"github.com/mekdep/server/internal/utils"
)

func SendMessageUnreadReminder() error {
	args := models.UserFilterRequest{}
	args.Limit = new(int)
	args.Offset = new(int)
	*args.Limit = 50000
	*args.Offset = 0

	uu, total, err := store.Store().UsersFindBy(context.Background(), args)
	if err != nil {
		return err
	}
	userIds := []string{}
	for _, v := range uu {
		userIds = append(userIds, v.ID)
	}
	falsePtr := false
	truePtr := true
	before30days := time.Now().AddDate(0, 0, -30)
	unreadsCount, err := store.Store().GetMessageReadsQuery(context.Background(), models.GetMessageReadsQueryDto{
		UserIds:       &userIds,
		IsRead:        &falsePtr,
		NotifiedAtMax: &before30days,
		SetNotified:   &truePtr,
	})
	if err != nil {
		return err
	}

	for total > 0 {
		log.Println("checked", len(uu), total)
		for _, u := range uu {
			if unreadsCount[u.ID] > 0 {
				err := app.SendMessageUnreadReminder(&apiutils.Session{}, unreadsCount[u.ID], u)
				if err != nil {
					utils.LoggerDesc("SendMessageUnreadReminder").Error(err)
				}
			}
		}
		total -= *args.Limit
		*args.Offset += *args.Limit
		uu, _, err = store.Store().UsersFindBy(context.Background(), args)
		if err != nil {
			return err
		}
		log.Println("load")
		userIds := []string{}
		for _, v := range uu {
			userIds = append(userIds, v.ID)
		}
		falsePtr := false
		truePtr := false
		unreadsCount, err = store.Store().GetMessageReadsQuery(context.Background(), models.GetMessageReadsQueryDto{
			UserIds:     &userIds,
			IsRead:      &falsePtr,
			SetNotified: &truePtr,
		})
		if err != nil {
			return err
		}
		log.Println("unreadsCount")
	}

	return nil
}
