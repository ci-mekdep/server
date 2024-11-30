package app

import (
	"strings"

	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	apputils "github.com/mekdep/server/internal/utils"
	"golang.org/x/net/context"
)

func SendSMS(phones []string, text string, smsType models.SmsType) error {
	err := apputils.SendMessageToPhones(phones, text)
	errMsg := new(string)
	if err != nil {
		*errMsg = err.Error()
	}
	if err != nil && strings.Contains(err.Error(), "not set") {
		err = nil
	}
	_, _ = store.Store().SmsSenderCreate(context.Background(), &models.SmsSender{
		Phones:      &phones,
		Message:     text,
		Type:        string(smsType),
		LeftTry:     3,
		IsCompleted: err == nil,
		ErrorMsg:    errMsg,
	})

	return nil // prevent UI "cant complete request"
}
