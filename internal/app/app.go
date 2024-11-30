package app

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	apputils "github.com/mekdep/server/internal/utils"
	"github.com/patrickmn/go-cache"
	"go.elastic.co/apm/v2"
)

var app *App

func Ap() *App {
	return app
}

func Init() *App {
	app = &App{}
	err := utils.SessionStoreInit()
	if err != nil {
		log.Fatal(err)
	}
	app.cache = cache.New(1*time.Hour, 1*time.Hour)

	return app
}

type App struct {
	cache *cache.Cache
}

type AppError struct {
	code    string
	key     string
	comment string
}

func NewAppError(code, key, comment string) *AppError {
	return &AppError{code, key, comment}
}

func (err AppError) Error() string {
	return strings.Trim(err.code+" : "+err.key+" - "+err.comment, " -:")
}

func (err AppError) Code() string {
	return err.code
}

func (err AppError) Key() string {
	return err.key
}

func (err AppError) Comment() string {
	return err.comment
}

func (err *AppError) SetKey(key string) *AppError {
	er := *err
	er.key = key
	return &er
}

func (err *AppError) SetComment(comment string) *AppError {
	er := *err
	er.comment = comment
	return &er
}

// TODO: move this to new Package called app_errors
var (
	ErrInvalid   = NewAppError("invalid", "", "please fill with valid options")
	ErrRequired  = NewAppError("required", "", "please fill required field")
	ErrNotSet    = NewAppError("not_set", "", "")
	ErrNotExists = NewAppError("not_exists", "", "")
	ErrUnique    = NewAppError("unique", "", "")
	ErrExpired   = NewAppError("expired", "", "")
	ErrExceeded  = NewAppError("exceeded", "", "")
	ErrNotfound  = NewAppError("not_found", "", "")
	ErrNotPaid   = NewAppError("not_paid", "", "")

	ErrUnauthorized = NewAppError("unauthorized", "user", "")
	ErrForbidden    = NewAppError("forbidden", "user", "")
)

func userLog(ses *utils.Session, data models.UserLog) error {
	sp, ctx := apm.StartSpan(ses.Context(), "userLog", "app")
	ses.SetContext(ctx)
	defer sp.End()
	go func() {
		// delete security keys
		secKeys := []string{"password", "otp", "device_token", "token"}
		prStr, _ := json.Marshal(data.SubjectProperties)
		pr := map[string]interface{}{}
		_ = json.Unmarshal(prStr, &pr)
		for _, k := range secKeys {
			if _, ok := pr[k]; ok {
				delete(pr, k)
			}
		}
		data.SubjectProperties = pr

		_, err := store.Store().UserLogsCreate(context.Background(), data)
		if err != nil {
			apputils.LoggerDesc("in user log worker").Error(err)
		}
	}()
	return nil
}
