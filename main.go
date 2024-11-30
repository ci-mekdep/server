package main

import (
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/config"
	"github.com/mekdep/server/internal/api"
	"github.com/mekdep/server/internal/api/middleware"
	apiutils "github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/cmd"
	"github.com/mekdep/server/internal/store"
	"github.com/mekdep/server/internal/store/pgx"
	"github.com/mekdep/server/internal/utils"
	"go.elastic.co/apm/module/apmgin/v2"
)

func main() {
	defer utils.InitLogs().Close()
	config.LoadConfig()
	defer store.Init().(*pgx.PgxStore).Close()
	cmd.Init()
	app.Init()

	if config.Conf.AppEnvIsProd {
		gin.SetMode(gin.ReleaseMode)
	}
	routes := gin.Default()
	// CORS
	routes.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"http://localhost"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "RefreshToken", "Authorization"},
		AllowCredentials: true,
		AllowAllOrigins:  true,
		AllowWebSockets:  true,
		// MaxAge:           12 * time.Hour,
	}))
	DebugInit(routes)
	routes.Use(middleware.SetLoggerRequest)
	routes.Use(middleware.StartSession)
	api.Routes(routes)

	go runSchedules()
	port := "8000"
	if config.Conf.HttpPort != "" {
		port = config.Conf.HttpPort
	}
	if err := routes.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func DebugInit(routes *gin.Engine) {

	if config.Conf.AppUrl != "" {
		os.Setenv("ELASTIC_APM_SERVER_URL", config.Conf.ElasticApmServerUrl)
		os.Setenv("ELASTIC_APM_SECRET_TOKEN", config.Conf.ElasticApmSecretToken)
		os.Setenv("ELASTIC_APM_SERVICE_NAME", "api")

		if config.Conf.AppEnvIsProd {
			routes.Use(apmgin.Middleware(routes))
		}
	}
	if config.Conf.SentryDsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			ServerName:         "api",
			Dsn:                config.Conf.SentryDsn,
			EnableTracing:      true,
			TracesSampleRate:   0.2, // 20% of requests to trace
			ProfilesSampleRate: 0.8, // of (20%) them 80% to profile detailed
			// Enable printing of SDK debug messages.
			// Useful when getting started or trying to figure something out.
			Debug: !config.Conf.AppEnvIsProd,
		})
		routes.Use(sentrygin.New(sentrygin.Options{}))
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
	}
}

func runSchedules() {
	ticker := time.NewTicker(60 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				runEveryMinute()
				// ticker.Stop()
			}
		}
	}()
}

func runEveryMinute() {
	now := time.Now().In(config.RequestLocation)
	// utils.LoggerDesc("Running runEveryMinute").Info()
	// SendDailySms
	isEvening := now.Hour() == 18 && now.Minute() == 30
	isAfternoon := now.Hour() == 13 && now.Minute() == 50
	isMidnight := now.Hour() == 00 && now.Minute() == 00
	if isEvening || isAfternoon {
		utils.LoggerDesc("Running SendDailySms... Last ran: " + cmd.SendDailySmsLastRun.Format(time.DateTime)).Info()
		go func() {
			err := cmd.SendDailySms(isEvening)
			if err != nil {
				utils.LoggerDesc("In SendDailySms").Error(err)
			}
		}()
		cmd.SendDailySmsLastRun = now
	}
	if isEvening {
		utils.LoggerDesc("Running SendMessageUnreadReminder...").Info()
		go func() {
			// err := cmd.SendMessageUnreadReminder()
			// if err != nil {
			// 	utils.LoggerDesc("In SendMessageUnreadReminder").Error(err)
			// }
		}()
	}
	if isEvening {
		utils.LoggerDesc("Running SendTariffEndsSmsAll... Last ran: " + cmd.SendDailySmsLastRun.Format(time.DateTime)).Info()
		go func() {
			err := cmd.SendTariffEndsSmsAll(isEvening)
			if err != nil {
				utils.LoggerDesc("In SendTariffEndsSmsAll").Error(err)
			}
		}()
	}
	if isMidnight {
		utils.LoggerDesc("Running UpdatePeriodGrades...").Info()
		go func() {
			err := cmd.UpdatePeriodGrades(&apiutils.Session{}, 0, "", "", false)
			if err != nil {
				utils.LoggerDesc("In SendTariffEndsSmsAll").Error(err)
			}
		}()
		go func() {
			err := cmd.UpdatePaymentStatus(&apiutils.Session{})
			if err != nil {
				utils.LoggerDesc("In SendTariffEndsSmsAll").Error(err)
			}
		}()
	}
}
