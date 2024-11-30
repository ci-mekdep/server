package models

const TeacherPointWeight = 10

type LessonLikes struct {
	ID       string  `json:"id"`
	UserId   *string `json:"user_id"`
	LessonId *string `json:"lesson_id"`
	User     *User   `json:"user"`
	Lesson   *Lesson `json:"lesson"`
}

type LessonLikesRequest struct {
	ID       *string `json:"id"`
	UserId   *string `json:"user_id"`
	LessonId *string `json:"lesson_id"`
}

func (model *LessonLikes) FromRequest(request *LessonLikesRequest) error {
	if request.ID != nil {
		model.ID = *request.ID
	}
	model.LessonId = request.LessonId
	model.UserId = request.UserId
	return nil
}
