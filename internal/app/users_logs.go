package app

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func UserLogsList(ses *utils.Session, f models.UserLogFilterRequest) ([]*models.UserLogResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UserLogsList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	userLogs, total, err := store.Store().UserLogsFindBy(ses.Context(), f)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().UserLogsLoadRelations(ses.Context(), &userLogs)
	if err != nil {
		return nil, 0, err
	}

	userLogsResponse := []*models.UserLogResponse{}
	for _, userLog := range userLogs {
		s := models.UserLogResponse{}
		s.FromModel(userLog)
		userLogsResponse = append(userLogsResponse, &s)
	}
	return userLogsResponse, total, err
}
