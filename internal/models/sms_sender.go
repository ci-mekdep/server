package models

import "time"

type SmsType string

const SmsTypeOTP SmsType = "otp"
const SmsTypeDaily SmsType = "daily"
const SmsTypeReminder SmsType = "reminder"
const SmsTypeOther SmsType = "other"

type SmsSender struct {
	ID          string     `json:"id"`
	Phones      *[]string  `json:"phones"`
	Message     string     `json:"message"`
	Type        string     `json:"type"`
	ErrorMsg    *string    `json:"error_msg"`
	IsCompleted bool       `json:"is_completed"`
	LeftTry     uint       `json:"left_try"`
	TriedAt     *time.Time `json:"tried_at"`
	CreatedAt   *time.Time `json:"created_at"`
}

func (SmsSender) RelationFields() []string {
	return []string{}
}

type SmsSenderFilterRequest struct {
	ID *string `json:"id"`
	PaginationRequest
}
