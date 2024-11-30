package app

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/config"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func (a App) PublicSchools(ses *utils.Session, data models.SchoolFilterRequest) ([]*models.SchoolResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "PublicSchools", "app")
	ses.SetContext(ctx)
	defer sp.End()
	data.IsParent = new(bool)
	*data.IsParent = false
	if data.ParentUid != nil {
		p, _ := store.Store().SchoolsFindById(ses.Context(), *data.ParentUid)
		if p == nil {
			return nil, 0, ErrNotfound
		}
		if p.Code != nil && *p.Code == "ag" {
			data.ParentUid = nil
		}
	}
	l, total, err := store.Store().SchoolsFindBy(ses.Context(), data)
	if err != nil {
		return nil, 0, err
	}
	if data.Regions != nil {
		err = store.Store().SchoolsLoadParents(ses.Context(), &l)
		if err != nil {
			return nil, 0, err
		}
	}
	res := []*models.SchoolResponse{}
	for _, m := range l {
		item := models.SchoolResponse{}
		item.FromModel(m)
		res = append(res, &item)
	}
	return res, total, nil
}

func (a App) PublicSchoolRegions(data models.SchoolFilterRequest) (map[string][]*models.SchoolResponse, []models.SchoolValueResponse, int, error) {
	data.IsParent = new(bool)
	*data.IsParent = true
	l, total, err := store.Store().SchoolsFindBy(context.Background(), data)
	if err != nil {
		return nil, nil, 0, err
	}
	err = store.Store().SchoolsLoadRelations(context.Background(), &l)
	if err != nil {
		return nil, nil, 0, err
	}
	res := map[string][]*models.SchoolResponse{}
	resValues := []models.SchoolValueResponse{}
	for _, m := range l {
		item := models.SchoolResponse{}
		item.FromModel(m)
		for state, regions := range models.Regions {
			if _, ok := res[state]; !ok {
				res[state] = []*models.SchoolResponse{}
			}
			for _, region := range regions {
				if region == *item.Code {
					res[state] = append(res[state], &item)
					resValues = append(resValues, item.ToValues())
				}
			}
		}
	}
	return res, resValues, total, nil
}

func (a App) StudentRatingBySchool(ses *utils.Session, schoolId string) ([]models.UserRating, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StudentRatingBySchool", "app")
	ses.SetContext(ctx)
	defer sp.End()
	date := time.Now().Truncate(time.Hour*24*2).AddDate(0, 0, -7)
	startDate := date
	endDate := startDate.AddDate(0, 0, 6)

	userList, userValues, err := store.Store().StudentRatingBySchool(context.Background(), schoolId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	err = store.Store().UsersLoadRelationsClassrooms(context.Background(), &userList)
	if err != nil {
		return nil, err
	}

	res := []models.UserRating{}
	for k, v := range userList {
		resItem := models.UserRating{
			User:      &models.UserResponse{},
			Classroom: &models.ClassroomResponse{},
			Value:     userValues[k],
			Rating:    k + 1,
		}
		if len(v.Classrooms) > 0 {
			resItem.Classroom.FromModel(v.Classrooms[0].Classroom)
		}
		v.Classrooms = nil
		resItem.User.FromModel(v)
		res = append(res, resItem)
	}

	return res, nil
}

func AppSettings(ses *utils.Session, schoolId string) (map[string]interface{}, error) {
	// Create GradeReason instances
	gradeReasonNames := []models.GradeReason{}
	for _, v := range models.DefaultGradeReasons {
		gradeReasonNames = append(gradeReasonNames, models.GradeReason{
			Code: v[0],
			Name: models.GradeReasonName{
				Tm: v[1],
				Ru: v[2],
				En: v[3],
			},
		})
	}

	// Create LessonType instances
	lessonTypeNames := []models.LessonType{}
	for _, v := range models.DefaultLessonTypes {
		lessonTypeNames = append(lessonTypeNames, models.LessonType{
			Code: v[0],
			Name: models.LessonTypeName{
				Tm: v[1],
				Ru: v[2],
				En: v[3],
			}})
	}

	// Create MenuApp instances
	menuAdd := []models.MenuApps{}
	for _, v := range models.FavouriteResource {
		menuItem := models.MenuApps{
			Title: models.StringLocale{
				Tm: v[0],
				En: v[0],
				Ru: v[0],
			},
			Link: v[1],
			Icon: v[2],
		}
		menuAdd = append(menuAdd, menuItem)
	}
	// Create DefaultPeriod instance

	// Create SubjectSetting instances
	subjectNames := []models.SubjectElement{}
	for _, v := range models.DefaultSubjects {
		subjectNames = append(subjectNames, models.SubjectElement{Name: v[0], FullName: v[1], Code: v[0], Area: v[2], Color: v[3]})
	}

	req := SettingsQueryDto{}
	settings, err := SettingsGet(ses, req)
	if err != nil {
		return nil, err
	}
	if config.Conf.SettingLoginAlert != nil {
		settings.AlertMessage = *config.Conf.SettingLoginAlert
	}
	role := models.Role("")
	if ses.GetRole() != nil {
		role = *ses.GetRole()
	}
	school := &models.School{}
	if schoolId != "" {
		school, err = store.Store().SchoolsFindById(ses.Context(), schoolId)
		if err != nil {
			return nil, err
		}
	} else if schoolId == "" {
		school = ses.GetSchool()
	} else {
		school = nil
	}
	if school != nil {
		store.Store().SchoolsLoadParents(ses.Context(), &[]*models.School{school})
	}
	// Construct the final response
	response := gin.H{
		"general": models.GeneralSetting{
			APIVersion:                 &config.Conf.ApiVersion,
			MobileRequiredVersion:      &config.Conf.MobileRequiredVersion,
			APIInstances:               models.States,
			AlertMessage:               settings.AlertMessage,
			LoginAlert:                 config.Conf.SettingLoginAlert,
			ContactMessages:            AppSettingsContactMessagesDeprecated(role, school),
			Contact:                    appSettingsContactMessages(role),
			BankTypes:                  models.DefaultPaymentBank,
			BookCategories:             models.BookCategories,
			Holidays:                   models.DefaultHolidays,
			MenuApps:                   menuAdd,
			DefaultPeriod:              appSettingsPeriods(),
			GradeUpdateMinutes:         settings.GradeUpdateMinutes,
			AbsentUpdateMinutes:        settings.AbsentUpdateMinutes,
			DelayedGradeUpdateHours:    settings.DelayedGradeUpdateHours,
			TimetableUpdateCurrentWeek: config.Conf.TimetableUpdateWeekAvailable,
			IsArchive:                  settings.IsArchive,
			IsForeignCountry:           config.Conf.IsForeignCountry,
			UserDocumentKeys:           models.UserDocumentKeys,
			ContactPhones:              settings.ContactPhones,
		},
		"lesson": models.LessonSetting{
			StudentComments: models.StudentComments,
			LessonTypes:     lessonTypeNames,
			GradeReasons:    gradeReasonNames,
		},
		"subject": models.SubjectSetting{
			BaseSubjectSetting: models.SubjectCategories,
			ClassroomGroupKeys: models.ClassroomGroupKeys,
			SubjectSetting:     subjectNames,
			TopicTags:          models.TopicTags,
		},
	}
	return response, nil
}

func appSettingContactMEssagesPhonesDeprecated(school *models.School) []map[string]string {
	var res []map[string]string
	if school != nil && school.Phone != nil {
		log.Println(res)
		res = append(res, map[string]string{
			"name":  "Mekdep admin",
			"value": *school.Phone,
		})
		if school.Parent != nil && school.Parent.Phone != nil {
			res = append(res, map[string]string{
				"name":  "Bilim müdirlik",
				"value": *school.Parent.Phone,
			})
		}
	}
	res = append(res, map[string]string{
		"name":  "Goldaw merkezi",
		"value": config.Conf.SupportPhone,
	})
	res = append(res, map[string]string{
		"name":  "E-poçta",
		"value": config.Conf.SupportEmail,
	})
	return res
}

func AppSettingsContactMessagesDeprecated(role models.Role, school *models.School) map[string][]map[string]string {
	contactMessages := make(map[string][]map[string]string)

	contactMessages["contact_videos"] = make([]map[string]string, 0)
	contactMessages["contact_phones"] = appSettingContactMEssagesPhonesDeprecated(school)
	if role == models.RoleTeacher {
		contactMessages["suggestions"] = []map[string]string{}
		for _, v := range models.DefaultSuggestionsTeacher {
			contactMessages["suggestions"] = append(contactMessages["suggestions"], settingToMap(v))
		}
		contactMessages["complaints"] = []map[string]string{}
		for _, v := range models.DefaultComplaintsTeacher {
			contactMessages["complaints"] = append(contactMessages["complaints"], settingToMap(v))
		}
		contactMessages["contact_videos"] = []map[string]string{}
		for _, v := range models.ContactVideosTeacher {
			contactMessages["contact_videos"] = append(contactMessages["contact_videos"], map[string]string{
				"title":     v.Title,
				"file":      v.FileUrl,
				"image_url": v.ImageUrl,
			})
		}
	} else {
		contactMessages["suggestions"] = []map[string]string{}
		for _, v := range models.DefaultSuggestionsParent {
			contactMessages["suggestions"] = append(contactMessages["suggestions"], settingToMap(v))
		}
		contactMessages["complaints"] = []map[string]string{}
		for _, v := range models.DefaultComplaintsParent {
			contactMessages["complaints"] = append(contactMessages["complaints"], settingToMap(v))
		}
		contactMessages["contact_videos"] = []map[string]string{}
		for _, v := range models.ContactVideosParent {
			contactMessages["contact_videos"] = append(contactMessages["contact_videos"], map[string]string{
				"title":     v.Title,
				"file":      v.FileUrl,
				"image_url": v.ImageUrl,
			})
		}
	}
	return contactMessages
}

func settingToMap(str models.StringLocale) map[string]string {
	return map[string]string{
		"tm": str.Tm,
		"en": str.En,
		"ru": str.Ru,
	}
}

func appSettingsContactMessages(role models.Role) []models.SettingContactMessagesByRole {
	contactMessages := []models.SettingContactMessagesByRole{}

	contactMessages = append(contactMessages, models.SettingContactMessagesByRole{
		Role: models.RoleTeacher,
		Contact: models.SettingContactMessages{
			Complaints:  models.DefaultComplaintsTeacher,
			Suggestions: models.DefaultSuggestionsTeacher,
			Videos:      models.ContactVideosTeacher,
		},
	})
	contactMessages = append(contactMessages, models.SettingContactMessagesByRole{
		Role: models.RoleParent,
		Contact: models.SettingContactMessages{
			Complaints:  models.DefaultComplaintsParent,
			Suggestions: models.DefaultSuggestionsParent,
			Videos:      models.ContactVideosParent,
		},
	})
	contactMessages = append(contactMessages, models.SettingContactMessagesByRole{
		Role: models.RoleAdmin,
		Contact: models.SettingContactMessages{
			Complaints:  models.DefaultComplaintsTeacher,
			Suggestions: models.DefaultSuggestionsTeacher,
			Videos:      models.ContactVideosAdmin,
		},
	})
	return contactMessages
}

func appSettingsPeriods() *models.DefaultPeriod {
	return &models.DefaultPeriod{
		CurrentNumber: 1, // TODO: fix, make automatic by today
		Value: [][]string{
			{"2024-09-01", "2024-10-22"},
			{"2024-10-31", "2024-12-29"},
			{"2025-01-12", "2025-03-19"},
			{"2025-03-29", "2025-05-25"},
		},
	}
}
