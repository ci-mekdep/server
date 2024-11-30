package app

import (
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

type StatisticsFormResponseDto struct {
	School    models.SchoolResponse       `json:"school"`
	Settings  []models.SchoolSettingModel `json:"settings"`
	UpdatedAt *time.Time                  `json:"updated_at"` // TODO: handle this
}

type StatisticsFormUpdateRequestDto struct {
	SchoolId string                        `json:"school_id"`
	Items    []models.SchoolSettingRequest `json:"items"`
}

func (a App) StatisticsSchoolForm(ses *utils.Session, schoolIds []string) ([]StatisticsFormResponseDto, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsSchoolForm", "app")
	ses.SetContext(ctx)
	defer sp.End()

	args := models.SchoolFilterRequest{
		Uids: &schoolIds,
	}
	args.Limit = new(int)
	*args.Limit = 500
	args.IsParent = new(bool)
	*args.IsParent = false
	schools, _, err := store.Store().SchoolsFindBy(ses.Context(), args)
	if err != nil {
		return nil, err
	}

	err = store.Store().SchoolsLoadRelations(ses.Context(), &schools)
	if err != nil {
		return nil, err
	}

	schoolSettings, err := store.Store().SchoolSettingsGet(ses.Context(), schoolIds)
	if err != nil {
		return nil, err
	}

	defaultSettings, _ := a.StatisticsSchoolInputOptions(ses)

	res := []StatisticsFormResponseDto{}
	for _, v := range schools {
		resItem := StatisticsFormResponseDto{}
		resItem.School = models.SchoolResponse{}
		resItem.School.FromModel(v)
		resItem.Settings = []models.SchoolSettingModel{}

		for _, defaultSetting := range defaultSettings {
			var schoolSetting *models.SchoolSetting
			for _, vv := range schoolSettings {
				if *vv.SchoolId == v.ID && vv.Key == defaultSetting.Key {
					schoolSetting = &vv
					if resItem.UpdatedAt == nil ||
						vv.UpdatedAt != nil && resItem.UpdatedAt.Before(*vv.UpdatedAt) {
						resItem.UpdatedAt = vv.UpdatedAt
					}
					break
				}
			}
			if schoolSetting != nil {
				defaultSetting.FromModel(*schoolSetting)
			}
			defaultSetting.SchoolId = v.ID
			resItem.Settings = append(resItem.Settings, defaultSetting)
		}
		res = append(res, resItem)
	}
	return res, nil
}

func (a App) StatisticsSchoolFormUpdate(ses *utils.Session, dto []StatisticsFormUpdateRequestDto) ([]StatisticsFormResponseDto, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsSchoolInputUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()

	queryDto := []models.SchoolSettingRequest{}
	for _, v := range dto {
		for _, vv := range v.Items {
			queryDto = append(queryDto, models.SchoolSettingRequest{
				SchoolId: &v.SchoolId,
				Key:      vv.Key,
				Value:    vv.Value,
			})
		}
	}
	err := store.Store().SchoolSettingsUpdate(ses.Context(), "", queryDto)
	if err != nil {
		return nil, err
	}

	return a.StatisticsSchoolForm(ses, ses.SchoolAllIds())
}

func (a App) StatisticsSchoolInput(ses *utils.Session, schoolId string) ([]models.SchoolSetting, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsSchoolInput", "app")
	ses.SetContext(ctx)
	defer sp.End()
	st, err := store.Store().SchoolSettingsGet(ses.Context(), []string{schoolId})
	opt, _ := a.StatisticsSchoolInputOptions(ses)
	for _, v := range opt {
		isExists := false
		for _, vv := range st {
			if vv.Key == v.Key {
				isExists = true
				break
			}
		}
		if !isExists {
			st = append(st, models.SchoolSetting{
				Key: v.Key,
			})
		}
	}
	if err != nil {
		return nil, err
	}
	return st, nil
}

func (a App) StatisticsSchoolInputUpdate(ses *utils.Session, schoolId string, st []models.SchoolSettingRequest) ([]models.SchoolSetting, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsSchoolInputUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	for k := range st {
		st[k].SchoolId = &schoolId
	}
	err := store.Store().SchoolSettingsUpdate(ses.Context(), schoolId, st)
	if err != nil {
		return nil, err
	}
	// get
	stdb, err := store.Store().SchoolSettingsGet(ses.Context(), []string{schoolId})
	if err != nil {
		return nil, err
	}
	return stdb, nil
}

func (a App) StatisticsSchoolInputOptions(ses *utils.Session) ([]models.SchoolSettingModel, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsSchoolInputOptions", "app")
	ses.SetContext(ctx)
	defer sp.End()
	res := []models.SchoolSettingModel{
		{
			Key:     models.SchoolSettingTeachersCount,
			Title:   "Mugallym sany",
			Type:    models.SchoolSettingTypeNumber,
			Options: nil,
		},
		{
			Key:     models.SchoolSettingStudentsCount,
			Title:   "Okuwçy sany",
			Type:    models.SchoolSettingTypeNumber,
			Options: nil,
		},
		{
			Key:     models.SchoolSettingParentsCount,
			Title:   "Ata-ene sany",
			Type:    models.SchoolSettingTypeNumber,
			Options: nil,
		},
		{
			Key:     models.SchoolSettingClassroomsCount,
			Title:   "Synp sany",
			Type:    models.SchoolSettingTypeNumber,
			Options: nil,
		},
		{
			Key:     models.SchoolSettingStudentsPlaceCount,
			Title:   "Näçe orunlyk mekdep (okuwçy sany)",
			Type:    models.SchoolSettingTypeNumber,
			Options: nil,
		},
		{
			Key:     models.SchoolSettingRoomsCount,
			Title:   "Okuw otag sany",
			Type:    models.SchoolSettingTypeNumber,
			Options: nil,
		},
		{
			Key:     models.SchoolSettingBoardsCount,
			Title:   "Akylly tagta sany",
			Type:    models.SchoolSettingTypeNumber,
			Options: nil,
		},
		{
			Key:     models.SchoolSettingComputersCount,
			Title:   "Kompýuter sany",
			Type:    models.SchoolSettingTypeNumber,
			Options: nil,
		},
		{
			Key:     models.SchoolSettingInternetSpeed,
			Title:   "Internet tizligi",
			Type:    models.SchoolSettingTypeOptions,
			Options: &[]string{"Gurnalmadyk", "Internet tölegsiz", "0.5mbit", "1mbit", "2mbit", "4mbit", "6mbit", "6mbit ýokary"},
		},
		{
			Key:     models.SchoolSettingInternetArea,
			Title:   "Internet doly mekdep çäginde tutýarmy?",
			Type:    models.SchoolSettingTypeOptions,
			Options: &[]string{"Ähli otaglarda we binalarda tutýar", "Diňe esasy binada tutýar, beýleki binalarda tutanok", "Käbir otaglarda tutýar", "Hiç hili tutanok"},
		},
		{
			Key:   models.SchoolSettingLocalNetwork,
			Title: "Içki tor gurnalanmy?",
			Type:  models.SchoolSettingTypeBoolean,
		},
	}
	return res, nil
}
