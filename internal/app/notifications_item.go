package app

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func UserNotificationsList(ses *utils.Session, f models.UserNotificationFilterRequest) ([]*models.UserNotificationResponse, int, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UserNotificationsList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	// get from store
	userNotifications, total, err := store.Store().UserNotificationsFindBy(ses.Context(), f)
	if err != nil {
		return nil, 0, 0, err
	}
	err = store.Store().UserNotificationsLoadRelations(ses.Context(), &userNotifications)
	if err != nil {
		return nil, 0, 0, err
	}
	nn := []*models.Notifications{}
	for _, v := range userNotifications {
		if v.Notifications != nil {
			nn = append(nn, v.Notifications)
		}
	}
	err = store.Store().NotificationsLoadRelations(ses.Context(), &nn)
	if err != nil {
		return nil, 0, 0, err
	}

	// to response
	userNotificationsResponse := []*models.UserNotificationResponse{}
	for _, userNotification := range userNotifications {
		s := models.UserNotificationResponse{}
		s.FromModel(userNotification)
		userNotificationsResponse = append(userNotificationsResponse, &s)
	}

	// set read
	if f.IsRead != nil && *f.IsRead {
		ids := []string{}
		for _, v := range userNotifications {
			ids = append(ids, v.ID)
		}
		err = store.Store().UserNotificationsUpdateRead(ses.Context(), ids)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	totalUnread, err := store.Store().UserNotificationsSelectTotalUnread(ses.Context(), *f.UserId, *f.Role)

	return userNotificationsResponse, total, totalUnread, err
}

func UserNotificationUpdate(ses *utils.Session, data models.UserNotificationRequest) (*models.UserNotificationResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UserNotificationUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	model := &models.UserNotification{
		ID:           *data.ID,
		Comment:      data.Comment,
		CommentFiles: data.CommentFiles,
	}
	data.ToModel(model)
	var err error
	model, err = store.Store().UserNotificationsUpdate(ses.Context(), *model)
	if err != nil {
		return nil, err
	}
	err = store.Store().UserNotificationsLoadRelations(ses.Context(), &[]*models.UserNotification{model})
	if err != nil {
		return nil, err
	}

	res := &models.UserNotificationResponse{}
	res.FromModel(model)
	return res, nil
}
