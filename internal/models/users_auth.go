package models

import (
	"time"

	"github.com/mileusna/useragent"
)

type Session struct {
	ID          string
	Token       string
	UserId      string
	DeviceToken *string
	Agent       string
	Ip          string
	Iat         time.Time
	Exp         time.Time
	Lat         time.Time
	User        User
}

func (Session) RelationFields() []string {
	return []string{"User"}
}

type SessionResponse struct {
	ID             string    `json:"id"`
	UserId         string    `json:"user_id"`
	Agent          string    `json:"agent"`
	Ip             string    `json:"ip"`
	Os             string    `json:"os"`
	OsVersion      string    `json:"os_version"`
	Browser        string    `json:"browser"`
	BrowserVersion string    `json:"browser_version"`
	Device         string    `json:"device"`
	Iat            time.Time `json:"iat"`
	Lat            time.Time `json:"lat"`
}

func (r *SessionResponse) FromModel(m *Session) error {
	r.ID = m.ID
	r.UserId = m.UserId
	r.Agent = m.Agent
	r.Ip = m.Ip
	r.Iat = m.Iat
	r.Lat = m.Lat

	agent := useragent.Parse(m.Agent)

	r.Os = agent.OS
	r.OsVersion = agent.OSVersion
	r.Browser = agent.Name
	r.BrowserVersion = agent.Version
	if r.Device == "" {
		if agent.Mobile {
			r.Device = "Mobile"
		} else if agent.Tablet {
			r.Device = "Tablet"
		} else if agent.Desktop {
			r.Device = "Desktop"
		} else if agent.Bot {
			r.Device = "Bot"
		} else {
			r.Device = "Mobile"
		}
	}
	if agent.Mobile {
		r.Browser = agent.Device
	}
	return nil
}

type SessionFilter struct {
	ID          *string
	Token       *string
	DeviceToken *string
	UserId      *string
	Ip          *string
	Exp         *time.Time
	Lat         *time.Time
	PaginationRequest
}
