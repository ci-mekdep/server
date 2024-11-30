package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"slices"

	"firebase.google.com/go/v4/messaging"
	"github.com/appleboy/go-fcm"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	apputils "github.com/mekdep/server/internal/utils"
	"go.elastic.co/apm/v2"
)

func NotificationList(ses *utils.Session, f models.NotificationsFilterRequest) ([]*models.NotificationsResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "NotificationList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	notifications, total, err := store.Store().NotificationsFindBy(ses.Context(), f)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().NotificationsLoadRelations(ses.Context(), &notifications)
	if err != nil {
		return nil, 0, err
	}

	notificationsResponse := []*models.NotificationsResponse{}
	for _, notification := range notifications {
		n := models.NotificationsResponse{}
		n.FromModel(notification)
		notificationsResponse = append(notificationsResponse, &n)
	}
	return notificationsResponse, total, err
}

func NotificationDetail(f models.UserNotificationFilterRequest) (*models.NotificationsResponse, int, error) {
	n, err := store.Store().NotificationFindById(context.Background(), *f.NotificationId)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().NotificationsLoadRelations(context.Background(), &[]*models.Notifications{n})
	if err != nil {
		return nil, 0, err
	}

	nr := models.NotificationsResponse{}
	nr.FromModel(n)
	nr.Items = &[]models.UserNotificationResponse{}

	nn, total, err := store.Store().UserNotificationsFindBy(context.Background(), f)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().UserNotificationsLoadRelationUser(&nn)
	if err != nil {
		return nil, 0, err
	}
	for _, v := range nn {
		v.Notifications = nil
		nnr := models.UserNotificationResponse{}
		nnr.FromModel(v)
		*nr.Items = append(*nr.Items, nnr)
	}

	return &nr, total, err
}

func NotificationUpdate(ses *utils.Session, data models.NotificationsRequest) (*models.NotificationsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "NotificationUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	model := &models.Notifications{
		Title:   data.Title,
		Content: data.Content,
		Files:   data.Files,
	}
	data.ToModel(model)

	var err error
	model, err = store.Store().NotificationUpdate(ses.Context(), model)
	if err != nil {
		return nil, err
	}
	err = store.Store().NotificationsLoadRelations(ses.Context(), &[]*models.Notifications{model})
	if err != nil {
		return nil, err
	}

	res := &models.NotificationsResponse{}
	res.FromModel(model)
	return res, nil
}

func NotificationCreate(ses *utils.Session, data models.NotificationsRequest) (*models.NotificationsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "NotificationCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	model := &models.Notifications{}
	data.ToModel(model)

	if model.SchoolIds == nil || len(model.SchoolIds) < 1 {
		model.SchoolIds = []string{}
		if ses.GetSchoolId() != nil {
			model.SchoolIds = []string{*ses.GetSchoolId()}
		}
	} else {
		tmpIds := model.SchoolIds
		model.SchoolIds = []string{}
		for _, v := range tmpIds {
			if slices.Contains(ses.GetSchoolIds(), v) {
				model.SchoolIds = append(model.SchoolIds, v)
			}
		}
		if len(tmpIds) < 1 {
			return nil, ErrForbidden.SetKey("school_id")
		}
	}
	if model.SchoolIds == nil || len(model.SchoolIds) < 1 {
		return nil, ErrRequired.SetKey("school_ids")
	}
	if model.Roles == nil || len(model.Roles) < 1 {
		return nil, ErrRequired.SetKey("roles")
	}
	model.AuthorID = &ses.GetUser().ID

	var err error
	model, err = store.Store().NotificationCreate(ses.Context(), model)
	if err != nil {
		return nil, err
	}

	err = createNotificationItems(*model)
	if err != nil {
		return nil, err
	}

	err = store.Store().NotificationsLoadRelations(ses.Context(), &[]*models.Notifications{model})
	if err != nil {
		return nil, err
	}

	res := &models.NotificationsResponse{}
	res.FromModel(model)
	return res, nil
}

func NotificationDelete(ses *utils.Session, notificationID string) (err error) {
	sp, ctx := apm.StartSpan(ses.Context(), "NotificationDelete", "app")
	defer sp.End()

	id, err := store.Store().NotificationFindById(ctx, notificationID)
	if err != nil {
		return err
	}

	err = store.Store().NotificationDelete(ses.Context(), id.ID)
	if err != nil {
		return err
	}
	return nil
}

func createNotificationItems(model models.Notifications) error {
	if model.SchoolIds == nil {
		return ErrRequired.SetKey("school_ids").SetComment("in notification items creation")
	}
	adminCreated := false
	orgCreated := false
	for _, ad := range model.Roles {
		if ad == string(models.RoleAdmin) {
			if !adminCreated {
				err := createNotificationItem(model, "", models.Role(ad), model.UserIds)
				if err != nil {
					return err
				}
			}
			adminCreated = true
		} else if ad == string(models.RoleOrganization) {
			if !orgCreated {
				err := createNotificationItem(model, "", models.Role(ad), model.UserIds)
				if err != nil {
					return err
				}
				orgCreated = true
			}
		} else {
			for _, sid := range model.SchoolIds {
				err := createNotificationItem(model, sid, models.Role(ad), model.UserIds)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func createNotificationItem(n models.Notifications, sid string, r models.Role, ids []string) error {
	rStr := string(r)
	sidP := &sid
	if r == models.RoleAdmin || r == models.RoleOrganization {
		sidP = nil
	} else {
		sidP = &sid
	}
	f := models.UserFilterRequest{
		SchoolId: sidP,
		Role:     &rStr,
	}
	f.Limit = new(int)
	*f.Limit = 10000
	if ids != nil && len(ids) > 0 {
		f.Ids = &ids
	}
	uu, _, err := store.Store().UsersFindBy(context.Background(), f)
	if err != nil {
		return err
	}
	uids := []string{}
	nn := []models.UserNotification{}
	for _, u := range uu {
		n := models.UserNotification{
			NotificationId: n.ID,
			Role:           &rStr,
			UserId:         u.ID,
		}
		nn = append(nn, n)
		uids = append(uids, u.ID)
	}

	// batch insert user notification
	err = store.Store().UserNotificationsCreateBatch(context.Background(), nn)
	if err != nil {
		return err
	}

	// do push
	err = sendNotificationPush(n, uids, PushTypeNotification, n.ID)
	if err != nil {
		return err
	}
	return nil
}

type PushType string

const PushTypeNotification PushType = "NotificationPage"
const PushTypeChat PushType = "ChatsPage"

func sendNotificationPush(n models.Notifications, userIds []string, t PushType, id string) error {
	ss, err := utils.SessionByUserIds(userIds)
	if err != nil {
		return nil
	}
	dts := []string{}
	for _, s := range ss {
		if s.DeviceToken != nil && *s.DeviceToken != "" {
			if !slices.Contains(dts, *s.DeviceToken) {
				dts = append(dts, *s.DeviceToken)
			}
		}
	}
	// log.Println("send push: "+*n.Title, len(userIds), len(dts))

	desc := ""
	if n.Content != nil {
		desc = *n.Content
		if len(desc) > 255 {
			desc = desc[:255]
		}
	} else {
		desc = ""
	}
	go func() {
		err := SendPushByTokens(dts, *n.Title, desc, map[string]string{
			"page": string(t),
			"id":   id,
		})
		if err != nil {
			apputils.LoggerDesc("SendPushByTokens").Error(err)
		}
	}()
	if err != nil {
		return nil
	}
	return nil
}

func SendPushByTokens(tokens []string, title string, desc string, data map[string]string) error {
	ctx := context.Background()
	client, err := fcm.NewClient(
		ctx,
		fcm.WithCredentialsFile("service_account.json"),
		// initial with service account
		// fcm.WithServiceAccount("my-client-id@my-project-id.iam.gserviceaccount.com"),
	)
	if err != nil {
		return err
	}

	// Send multicast message
	// Create a list containing up to 500 registration tokens.
	// This registration tokens come from the client FCM SDKs.
	limit := 500
	for page := 0; page < len(tokens)/limit+1; page++ {
		max := page*limit + limit
		if max > len(tokens) {
			max = len(tokens)
		}
		partTokens := tokens[page*limit : max]
		msg := &messaging.MulticastMessage{
			Notification: &messaging.Notification{
				Title: title,
				Body:  desc,
			},
			Data: data,
			APNS: &messaging.APNSConfig{
				Headers: map[string]string{
					"apns-priority": "10",
				},
			},
			Tokens: partTokens,
		}
		_, err = client.SendMulticast(
			ctx,
			msg,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func SendPushByTokensOld(dts []string, title string, desc string, data map[string]interface{}) error {
	dtsStr, _ := json.Marshal(dts)
	dataStr, _ := json.Marshal(data)
	if string(dataStr) != "" {
		dataStr = []byte(`"data":` + string(dataStr) + `,`)
	}
	body := []byte(`{
		"notification":{
			"title": "` + title + `",
			"body": "` + desc + `"
		},
		` + string(dataStr) + `
		"apns": {
			"headers": {
				"apns-priority": 10
			},
			"payload": {
				"aps": "possible"
			}
		},
		"registration_ids": ` + string(dtsStr) + `
	}`)
	// log.Println(string(body))
	r, err := http.NewRequest("POST", FcmUrl, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	r.Header.Add("Authorization", "key="+FcmApiKey)
	r.Header.Add("Content-Type", "application/json")
	defer r.Body.Close()

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if err != nil {
		return err
	}
	return nil
}

const FcmUrl = "https://fcm.googleapis.com/fcm/send"
const FcmApiKey = "AAAA4GvElcc:APA91bHW2tjaILKzKbHJvHVT5LCvF8SmPm146HsupD90JrJo837NExX3e4SZDlDqWOrK2I2rTP1vJbsQDFPMjUprM9e3V_TeT4OTzNZTIAEIMXz5m4AYziCWICYPhl8KctXNUwsGKnJY"
