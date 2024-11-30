package app

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
)

type SettingsQueryDto struct {
}

func convertSettingsQueryToMap(query SettingsQueryDto) (m map[string]interface{}) {
	return
}

func SettingsGet(ses *utils.Session, query SettingsQueryDto) (response models.GeneralSetting, err error) {
	settings, err := store.Store().SettingsFindBy(ses.Context(), convertSettingsQueryToMap(query))
	if err != nil {
		return models.GeneralSetting{}, err
	}
	response = models.GeneralSetting{}
	for _, v := range settings {
		// TODO: refactor keys (reflect)
		if v.Value == "" {
			continue
		}
		switch v.Key {
		case "absent_update_minutes":
			json.Unmarshal([]byte(v.Value), &response.AbsentUpdateMinutes)
		case "delayed_grade_update_hours":
			json.Unmarshal([]byte(v.Value), &response.DelayedGradeUpdateHours)
		case "grade_update_minutes":
			json.Unmarshal([]byte(v.Value), &response.GradeUpdateMinutes)
		case "is_archive":
			json.Unmarshal([]byte(v.Value), &response.IsArchive)
		case "timetable_update_week_available":
			json.Unmarshal([]byte(v.Value), &response.TimetableUpdateCurrentWeek)
		case "alert_message":
			response.AlertMessage = v.Value
		}
	}
	return response, nil
}

func SettingsUpdate(ses *utils.Session, query models.GeneralSettingRequest) (response models.GeneralSetting, err error) {
	now := time.Now()
	settingsMap := structToMap(query)
	for key, value := range settingsMap {
		if value == nil {
			continue
		}
		valueStr, err := json.Marshal(value)
		_, err = store.Store().SettingsUpsert(ses.Context(), &models.Settings{
			Key:       key,
			Value:     strings.Trim(string(valueStr), " \""),
			UpdatedAt: &now,
		})
		if err != nil {
			return models.GeneralSetting{}, err
		}
	}
	response, err = SettingsGet(ses, SettingsQueryDto{})
	if err != nil {
		return models.GeneralSetting{}, err
	}
	return response, nil
}

func structToMap(obj interface{}) map[string]interface{} {
	str, _ := json.Marshal(obj)
	m := map[string]interface{}{}
	json.Unmarshal(str, &m)
	return m
}
