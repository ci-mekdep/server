package api

import (
	"encoding/json"
	"io/ioutil"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/app"
)

type ChangeLogStruct struct {
	Version string    `json:"version"`
	Text    []string  `json:"text"`
	Date    time.Time `json:"date"`
	IsNew   bool      `json:"is_new"`
}

func (cl *ChangeLogStruct) UnmarshalJSON(data []byte) error {
	var aux struct {
		Version string   `json:"version"`
		Text    []string `json:"text"`
		Date    string   `json:"date"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	cl.Version = aux.Version
	cl.Text = aux.Text

	// Parse the date string
	parsedDate, err := time.Parse("2006-01-02", aux.Date)
	if err != nil {
		return err
	}
	cl.Date = parsedDate

	return nil
}

func loadChangeLogData() ([]ChangeLogStruct, error) {
	data, err := ioutil.ReadFile("changeLog.json")
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []ChangeLogStruct{}, nil
	}

	var changeLogs []ChangeLogStruct
	if err := json.Unmarshal(data, &changeLogs); err != nil {
		return nil, err
	}

	return changeLogs, nil
}

func ChangeLog(c *gin.Context) {
	// Load changelog data from JSON file
	changeLogs, err := loadChangeLogData()
	if err != nil {
		app.NewAppError("Not loaded data from json file", "", "")
		return
	}

	// Sort changelogs by date in descending order
	sort.Slice(changeLogs, func(i, j int) bool {
		return changeLogs[i].Date.After(changeLogs[j].Date)
	})

	// Get the version query parameter
	versionQuery := c.Query("version")
	if versionQuery != "" {
		// Update `IsCurrent` field based on the query
		for i := range changeLogs {
			changeLogs[i].IsNew = changeLogs[i].Version >= versionQuery
		}
	}

	// Return all versions
	Success(c, gin.H{"versions": changeLogs})
}
