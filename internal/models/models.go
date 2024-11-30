package models

type PaginationRequest struct {
	Limit  *int `json:"limit" form:"limit"`
	Offset *int `json:"offset" form:"offset"`
}
