package api

import (
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/config"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
)

func PublicRoutes(api *gin.RouterGroup) {
	routes := api.Group("")
	{
		routes.GET("/public/schools", PublicSchoolList)
		routes.GET("/public/regions", PublicSchoolRegions)
		routes.GET("/public/regions/values", SchoolRegionsListValues)
		routes.GET("/public/regions/v2", PublicSchoolRegionsV2)
		routes.GET("/permissions", Permissions)
		routes.GET("/dashboards", Dashboards)
		routes.GET("/dashboards/v2", DashboardNumbersV2)
		routes.GET("/dashboards/details", DashboardDetails)
		routes.GET("/dashboards/details/v2", DashboardDetailsV2)
		routes.GET("/public/school/rating", PublicStudentRatingBySchool)
		routes.GET("/rating/school", StudentRatingBySchool)
		routes.GET("/settings", AppSettings)
		routes.GET("/settings/v2", AppSettingsV2)
		routes.POST("/settings", AppSettingsUpdate)
		routes.GET("/settings/online", SettingsOnline)
		routes.GET("/settings/contact", SettingsContact)
		routes.GET("/settings/report", SettingReport)
		routes.GET("/status", StatusFlutter)
		routes.GET("/changelog", ChangeLog)
	}
}

var ContactVideosParent = map[string]string{
	"title": "eMekdep tanyşdyryş",
	"file":  "https://mekdep.edu.tm/uploads/emekdep.mp4",
}
var ContactVideosTeacher = map[string]string{
	"title": "eMekdep tanyşdyryş",
	"file":  "https://mekdep.edu.tm/uploads/emekdep.mp4",
}

func SettingsContact(c *gin.Context) {
	role := c.Query("role")
	Success(c, gin.H{"contact_messages": app.AppSettingsContactMessagesDeprecated(models.Role(role), nil)})
}

func StatusFlutter(c *gin.Context) {
	Success(c, gin.H{})
}

func SettingReport(c *gin.Context) {
	reports := models.DefaultRatingReports

	if schoolIdStr := c.Query("school_id"); schoolIdStr != "" {
		for k, v := range reports {
			if v.Key == models.ReportKeySeasonStudents {
				_, count, err := store.Store().UsersFindBy(c, models.UserFilterRequest{
					SchoolId: &schoolIdStr,
				})
				if err != nil {
					continue
				}
				reports[k].Value = strconv.Itoa(count)
			}
			if v.Key == models.ReportKeySeasonStudentsCompleted {
				_, count, err := store.Store().UsersFindBy(c, models.UserFilterRequest{
					SchoolId: &schoolIdStr,
				})
				if err != nil {
					continue
				}
				reports[k].Value = strconv.Itoa(count)
			}
			if v.Key == models.ReportKeyCourses {
				_, count, err := store.Store().BaseSubjectsFindBy(c, models.BaseSubjectsFilterRequest{
					SchoolId: &schoolIdStr,
				})
				if err != nil {
					continue
				}
				reports[k].Value = strconv.Itoa(count)
			}
			if v.Key == models.ReportKeyGroups {
				_, count, err := store.Store().ClassroomsFindBy(c, models.ClassroomFilterRequest{
					SchoolId: &schoolIdStr,
				})
				if err != nil {
					continue
				}
				reports[k].Value = strconv.Itoa(count)
			}
		}
	}
	reportsData := []gin.H{}
	for _, v := range reports {
		reportsData = append(reportsData, gin.H{
			"header":      v.Title,
			"group":       v.Group,
			"description": v.Description,
			"key":         v.Key,
			"type":        v.Type,
			"value":       v.Value,
		})
	}
	Success(c, gin.H{
		"settings": gin.H{
			"report_keys": reportsData,
		},
	})
}

func SettingsOnline(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) (err error) {

		// count, err := store.Store().UsersOnlineCount(nil)

		if err != nil {
			return err
		}
		Success(c, gin.H{
			"online_count": utils.SessionOnlineCount(nil, 15),
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func Permissions(c *gin.Context) {
	ses := utils.InitSession(c)
	ps, psw, err := app.Ap().GetPermissions(&ses)
	if err != nil {
		handleError(c, err)
		return
	}
	Success(c, gin.H{
		"permissions_write": psw,
		"permissions":       ps,
	})
}

func AppSettingsUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSettings, func(user *models.User) (err error) {
		req := models.GeneralSettingRequest{}
		if errMsg, errKey := BindAndValidate(c, &req); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		req = models.GeneralSettingRequest{
			AlertMessage:               req.AlertMessage,
			AbsentUpdateMinutes:        req.AbsentUpdateMinutes,
			GradeUpdateMinutes:         req.GradeUpdateMinutes,
			IsArchive:                  req.IsArchive,
			DelayedGradeUpdateHours:    req.DelayedGradeUpdateHours,
			TimetableUpdateCurrentWeek: req.TimetableUpdateCurrentWeek,
		}
		settings, err := app.SettingsUpdate(&ses, req)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"settings": models.Setting{
				General: settings,
			},
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func AppSettingsV2(c *gin.Context) {
	gradeReasonNames := []models.GradeReason{}
	for _, v := range models.DefaultGradeReasons {
		gradeReasonNames = append(gradeReasonNames, models.GradeReason{
			Code: v[0],
			Name: models.GradeReasonName{
				Tm: v[1],
				Ru: v[2],
				En: v[3],
			}})
	}
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
	resTT := []models.PaymentTariffResponse{}
	for _, v := range models.DefaultTariff {
		resItem := models.PaymentTariffResponse{}
		resItem.FromModel(v)
		resTT = append(resTT, resItem)
	}
	menuApps := []models.MenuApps{}
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
		menuApps = append(menuApps, menuItem)
	}
	periodModel := struct {
		CurrentNumber int        `json:"current_number"`
		Value         [][]string `json:"value"`
	}{
		CurrentNumber: 1, // TODO: fix, make automatic by today
		Value: [][]string{
			{"2024-09-01", "2024-10-22"},
			{"2024-10-31", "2024-12-29"},
			{"2025-01-12", "2025-03-19"},
			{"2025-03-29", "2025-05-25"},
		},
	}
	type studentCommentType struct {
		Type     string              `json:"type"`
		Comments []map[string]string `json:"comments"`
	}
	studentComments := []struct {
		Name  string               `json:"name"`
		Types []studentCommentType `json:"types"`
	}{
		{
			Name: "Bilim",
			Types: []studentCommentType{
				{
					Type: "Gowy",
					Comments: []map[string]string{
						{
							"tm": "Çagaňyz okuwdan başarjaň! Tüweleme!",
							"ru": "Çagaňyz okuwdan başarjaň! Tüweleme!ru",
							"en": "Çagaňyz okuwdan başarjaň! Tüweleme!en",
						},
						{
							"tm": "Çagaňyz okuwdan başarjaň! Tüweleme!",
							"ru": "Çagaňyz okuwda kyn mesele çözmegi başardy! Tüweleme!",
							"en": "Çagaňyz okuwda kyn mesele çözmegi başardy! Tüweleme!",
						},
					},
				},
				{
					Type: "Erbet",
					Comments: []map[string]string{
						{
							"tm": "Çagaňyz okuwdan ýetişenok! Üns bermeli!",
							"ru": "Çagaňyz okuwdan ýetişenok! Üns bermeli!",
							"en": "Çagaňyz okuwdan ýetişenok! Üns bermeli!",
						},
					},
				},
			},
		},
		{
			Name: "Terbiýe",
			Types: []studentCommentType{
				{
					Type: "Gowy",
					Comments: []map[string]string{
						{
							"tm": "Çagaňyz sapakda ýokary terbiýesini görkezdi, ata-ene sagbolsun!",
							"ru": "Çagaňyz sapakda ýokary terbiýesini görkezdi, ata-ene sagbolsun!",
							"en": "Çagaňyz sapakda ýokary terbiýesini görkezdi, ata-ene sagbolsun!",
						},
					},
				},
				{
					Type: "Erbet",
					Comments: []map[string]string{
						{
							"tm": "Çagaňyz sapakda özüni terbiýesiz alyp barýar! Ata-ene çäre görüň!",
							"ru": "Çagaňyz sapakda özüni terbiýesiz alyp barýar! Ata-ene çäre görüň!",
							"en": "Çagaňyz sapakda özüni terbiýesiz alyp barýar! Ata-ene çäre görüň!",
						},
					},
				},
			},
		},
		{
			Name: "Gatnaşygy",
			Types: []studentCommentType{
				{
					Type: "Gowy",
					Comments: []map[string]string{
						{
							"tm": "Çagaňyz sapakda köp soraglara jogap berýär, tüweleme!",
							"ru": "Çagaňyz sapakda köp soraglara jogap berýär, tüweleme!",
							"en": "Çagaňyz sapakda köp soraglara jogap berýär, tüweleme!",
						},
					},
				},
				{
					Type: "Erbet",
					Comments: []map[string]string{

						{
							"tm": "Çagaňyz sapaklara höwesi ýok! Ata-ene çäre görüň!",
							"ru": "Çagaňyz sapaklara höwesi ýok! Ata-ene çäre görüň!",
							"en": "Çagaňyz sapaklara höwesi ýok! Ata-ene çäre görüň!",
						},
					},
				},
			},
		},
	}

	// Adding payment_map with calculated month_prices
	paymentMap := map[string]interface{}{}
	for _, tariff := range models.DefaultTariff {
		monthlyPrice := tariff.Price
		var annualPrice float64
		if tariff.Code == models.PaymentFree {
			annualPrice = 0
		} else {
			annualPrice = (monthlyPrice * 12) - 40
		}
		paymentMap[string(tariff.Code)] = map[string]interface{}{
			"price": models.MoneyFromFloat(tariff.Price),
			"month_prices": map[int]models.Money{
				1:  models.MoneyFromFloat(monthlyPrice),
				12: models.MoneyFromFloat(annualPrice),
			},
		}
	}
	bookCategories := models.BookCategories
	loginAlert := config.Conf.SettingLoginAlert
	if config.Conf.AppIsReadonly != nil && *config.Conf.AppIsReadonly {
		loginAlert = new(string)
		*loginAlert = "Bu arhiw serwer - üýtgetmek mümkinçiligi ýok, diňe görmek üçin niýetlenen."
	}
	Success(c, gin.H{
		"settings": gin.H{
			"payment": models.PaymentSetting{
				MonthTypes:  []int{1, 3, 9},
				TariffTypes: resTT,
			},
			"payment_map":        paymentMap,
			"bank_types":         models.DefaultPaymentBank,
			"holidays":           models.DefaultHolidays,
			"grade_reasons":      gradeReasonNames,
			"lesson_types":       lessonTypeNames,
			"menu_apps":          menuApps,
			"student_comments":   studentComments,
			"period":             periodModel,
			"book_categories":    bookCategories,
			"subject_categories": models.SubjectCategories,
			"login_alert":        loginAlert,
			"app_version":        "v1.0",
			"required_version":   &config.Conf.MobileRequiredVersion,
		},
	},
	)
}

// TODO: move to app
func AppSettings(c *gin.Context) {
	ses := utils.InitSession(c)
	ses.LoadSession()
	schoolId := c.Param("id")
	response, err := app.AppSettings(&ses, schoolId)
	if err != nil {
		handleError(c, err)
		return
	}
	Success(c, response)
}

func PublicSchoolList(c *gin.Context) {
	r := models.SchoolFilterRequest{}
	if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
		handleError(c, app.NewAppError(errMsg, errKey, ""))
		return
	}
	schoolCode := c.Query("code")
	r.Code = &schoolCode
	r.Limit = new(int)
	*r.Limit = 500

	ses := utils.InitSession(c)
	l, total, err := app.Ap().PublicSchools(&ses, r)
	if err != nil {
		handleError(c, err)
		return
	}
	// ll := []*models.SchoolResponse{}
	// for _, v := range l {
	// 	if slices.Contains([]string{"ag76", "ag131", "ag29", "ag72", "ag51", "ag43", "ag55"}, *v.Code) {
	// 		ll = append(ll, v)
	// 	}
	// }
	Success(c, gin.H{
		"total":   total,
		"schools": l,
	})
}

func PublicSchoolRegions(c *gin.Context) {
	r := models.SchoolFilterRequest{}
	if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
		handleError(c, app.NewAppError(errMsg, errKey, ""))
		return
	}

	l, _, total, err := app.Ap().PublicSchoolRegions(r)
	if err != nil {
		handleError(c, err)
		return
	}
	ll := map[string][]*models.SchoolResponse{}
	for k, v := range l {
		label := models.GetStateLabel(k)
		if slices.Contains([]string{"ag", "ark"}, k) {
			if len(v) > 0 {
				ll[label] = []*models.SchoolResponse{v[0]}
			} else {
				ll[label] = []*models.SchoolResponse{}
			}
		} else {
			ll[label] = v
		}
	}

	Success(c, gin.H{
		"total":   total,
		"regions": ll,
	})
}

func PublicSchoolRegionsV2(c *gin.Context) {
	code := c.Query("code")
	ll, _, total, err := app.Ap().PublicSchoolRegions(models.SchoolFilterRequest{})
	if err != nil {
		handleError(c, err)
		return
	}
	var regions []*models.SchoolResponse
	for k, v := range ll {
		if code != "" && k != code {
			continue
		}
		if strings.HasPrefix(k, "ag") || strings.HasPrefix(k, "ark") {
			if len(v) > 0 {
				regions = append(regions, v[0])
			}
		} else {
			regions = append(regions, v...)
		}
	}
	response := gin.H{
		"total":   total,
		"regions": regions,
	}
	Success(c, response)
}

func PublicStudentRatingBySchool(c *gin.Context) {
	schoolId := c.Query("school_id")
	if schoolId == "" {
		handleError(c, app.NewAppError("school_id is required", "missing_parameter", ""))
		return
	}
	ses := utils.InitSession(c)
	rating, err := app.Ap().StudentRatingBySchool(&ses, schoolId)
	if err != nil {
		handleError(c, err)
		return
	}

	Success(c, gin.H{
		"rating": rating,
	})
}

func StudentRatingBySchool(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) error {
		if ses.GetSchoolId() == nil {
			return app.ErrNotSet.SetKey("school_id")
		}

		rating, err := app.Ap().StudentRatingBySchool(&ses, *ses.GetSchoolId())
		if err != nil {
			return err
		}

		Success(c, gin.H{
			"rating": rating,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
	}

}
