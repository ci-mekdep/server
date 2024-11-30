package app

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func LessonLikes(ses *utils.Session, data *models.LessonLikesRequest) (*models.LessonLikes, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "LessonLikes", "app")
	ses.SetContext(ctx)
	defer sp.End()

	alreadyLiked, err := store.Store().LessonsLikesByUser(ses.Context(), *data.LessonId, *data.UserId)
	if err != nil {
		return nil, err
	}
	if alreadyLiked {
		err = store.Store().LessonsLikesThenUnlike(ses.Context(), *data.LessonId, *data.UserId)
		if err != nil {
			return nil, err
		}
	} else {
		err = store.Store().LessonsLikes(ses.Context(), *data.LessonId, *data.UserId)
		if err != nil {
			return nil, err
		}
	}
	model := models.LessonLikes{}
	model.FromRequest(data)
	err = store.Store().LessonLikesLoadRelations(ses.Context(), &[]*models.LessonLikes{&model})
	return &model, err
}
