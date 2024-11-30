package config

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv                       string `mapstructure:"app_env"`
	AppUrl                       string `mapstructure:"app_url"`
	ApiVersion                   string `mapstructure:"api_version"`
	MobileRequiredVersion        int    `mapstructure:"mobile_required_version"`
	OtpValidationEnabled         *bool  `mapstructure:"otp_validation_enabled"`
	AppIsReadonly                *bool  `mapstructure:"app_is_readonly"`
	AdminMultiDevice             bool   `mapstructure:"admin_multi_device"`
	IsForeignCountry             bool   `mapstructure:"is_foreign_country"`
	TimetableUpdateWeekAvailable bool   `mapstructure:"timetable_update_week_available"`
	SupportPhone                 string `mapstructure:"support_phone"`
	SupportEmail                 string `mapstructure:"support_email"`
	ArchiveLink                  string `mapstructure:"archive_link"`

	DbConnection string `mapstructure:"db_connection"`
	DbHost       string `mapstructure:"db_host"`
	DbPort       string `mapstructure:"db_port"`
	DbDatabase   string `mapstructure:"db_database"`
	DbUsername   string `mapstructure:"db_username"`
	DbPassword   string `mapstructure:"db_password"`

	HttpPort     string `mapstructure:"http_port"`
	SmppHost     string `mapstructure:"smpp_host"`
	SmppPort     string `mapstructure:"smpp_port"`
	SmppLogin    string `mapstructure:"smpp_login"`
	SmppPassword string `mapstructure:"smpp_password"`

	SmppServerURL   string `mapstructure:"smpp_server_url"`
	SmppServerToken string `mapstructure:"smpp_server_token"`

	ElasticApmServerUrl   string `mapstructure:"elastic_apm_server_url"`
	ElasticApmSecretToken string `mapstructure:"elastic_apm_secret_token"`

	SentryDsn string `mapstructure:"sentry_dsn"`

	SettingLoginAlert *string `mapstructure:"setting_login_alert"`
	AppEnvIsProd      bool
	DevPhones         []string `mapstructure:"phones"`
}

const APP_ENV_PROD = "prod"
const APP_ENV_DEV = "dev"

var Conf Config
var RequestLocation *time.Location

func LoadConfig() {
	viper.SetConfigType("env")
	viper.SetConfigName(".env")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln(err)
	}
	if err := viper.Unmarshal(&Conf); err != nil {
		log.Fatalln(err)
	}

	if Conf.OtpValidationEnabled == nil {
		otpValidationEnabled := true
		Conf.OtpValidationEnabled = &otpValidationEnabled
	}
	if Conf.AppIsReadonly == nil || !*Conf.AppIsReadonly {
		Conf.AppIsReadonly = nil
	}
	Conf.AppEnvIsProd = (Conf.AppEnv == APP_ENV_PROD)
	if Conf.SettingLoginAlert != nil && *Conf.SettingLoginAlert == "" {
		Conf.SettingLoginAlert = nil
	}

	phones := viper.GetString("phones")
	if phones != "" {
		Conf.DevPhones = strings.Split(phones, ",")
	}

	// // init the loc
	RequestLocation, _ = time.LoadLocation("Asia/Ashgabat")
	UTCLoc, _ := time.LoadLocation("UTC")
	time.Local = UTCLoc
	// // set timezone,
	// os.Setenv("TZ", "Asia/Ashgabat")
}
