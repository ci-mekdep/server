package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/config"
)

func Routes(routes *gin.Engine) {
	if config.Conf.AppEnvIsProd {
		gin.SetMode("release")
	}

	api := routes.Group("/api")
	{
		PublicRoutes(api)
		UserRoutes(api)
		SchoolRoutes(api)
		PeriodRoutes(api)
		ClassroomRoutes(api)
		SubjectRoutes(api)
		TimetableRoutes(api)
		ShiftRoutes(api)
		DiaryRoutes(api)
		JournalRoutes(api)
		TeacherRoutes(api)
		UserNotificationsRoutes(api)
		NotificationsRoutes(api)
		UserLogsRoutes(api)
		ToolRoutes(api)
		ReportRoutes(api)
		ReportFormRoutes(api)
		AnalyticsRoutes(api)
		PaymentRoutes(api)
		SettingsRoutes(api)
		TopicRoutes(api)
		ContactItemsRoutes(api)
		MessageGroupRoutes(api)
		MessageRoutes(api)
		SubjectExamRoutes(api)
		BookRoutes(api)
		BaseSubjectsRoutes(api)
		ReportsRoutes(api)
		ReportItemsRoutes(api)
		TeacherExcuseRoutes(api)
		SchoolTransferRoutes(api)
	}
	routes.Static("/uploads", "./web/uploads")
	if !config.Conf.AppEnvIsProd {
		routes.StaticFile("/docs", "./api/docs.html")
		routes.StaticFile("/openapi.yaml", "./api/openapi.yaml")
		routes.StaticFile("/openapi.json", "./api/openapi.json")
	}
}
