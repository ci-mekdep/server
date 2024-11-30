package models

import (
	"fmt"
	"time"
)

type ContactType string

const ContactTypeReview = "review"
const ContactTypeComplaint = "complaint"
const ContactTypeSuggestion = "suggestion"
const ContactTypeDataComplaint = "data_complaint"

type ContactStatus string

const ContactStatusWaiting ContactStatus = "waiting"
const ContactStatusTodo ContactStatus = "todo"
const ContactStatusProcessing ContactStatus = "processing"
const ContactStatusDone ContactStatus = "done"
const ContactStatusBacklog ContactStatus = "backlog"
const ContactStatusRejected ContactStatus = "rejected"

type ContactItems struct {
	ID                   string           `json:"id"`
	UserId               *string          `json:"user_id"`
	SchoolId             *string          `json:"school_id"`
	Message              *string          `json:"message"`
	Type                 string           `json:"type"`
	Status               string           `json:"status"`
	Files                *[]string        `json:"files"`
	Note                 *string          `json:"note"`
	ClassroomName        *string          `json:"classroom_name"`
	ParentPhone          *string          `json:"parent_phone"`
	BirthCertNumber      *string          `json:"birth_cert_number"`
	RelatedChildrenCount int              `json:"related_children_count"`
	RelatedId            *string          `json:"related_id"`
	CreatedAt            *time.Time       `json:"created_at"`
	UpdatedAt            *time.Time       `json:"updated_at"`
	UpdatedBy            *string          `json:"updated_by_id"`
	User                 *User            `json:"user"`
	School               *School          `json:"school"`
	Related              *ContactItems    `json:"related"`
	RelatedChildren      *[]*ContactItems `json:"related_children"`
	UpdatedByUser        *User            `json:"updated_by"`
}

func (ContactItems) RelationFields() []string {
	return []string{"User", "School", "Related", "RelatedChildren", "UpdatedByUser"}
}

type ContactItemsRequest struct {
	ID                   *string          `json:"id" form:"id"`
	UserId               *string          `json:"user_id" form:"user_id"`
	SchoolId             *string          `json:"school_id" form:"school_id"`
	Message              *string          `json:"message" form:"message"`
	Type                 string           `json:"type" form:"type"`
	Status               string           `json:"status" form:"status"`
	Note                 *string          `json:"note" form:"note"`
	ParentFirstName      *string          `json:"parent_first_name" form:"parent_first_name"`
	ParentLastName       *string          `json:"parent_last_name" form:"parent_last_name"`
	ParentPhone          *string          `json:"parent_phone" form:"parent_phone"`
	ParentPassportNumber *string          `json:"parent_passport_number" form:"parent_passport_number"`
	ParentBirthday       *string          `json:"parent_birthday" form:"parent_birthday"`
	Children             []ChildrenObject `json:"children" form:"children"`
	RelatedId            *string          `json:"related_id" form:"related_id"`
	RelatedChildrenIds   *[]string        `json:"related_children_ids" form:"related_children_ids"`
	Files                *[]string        ``
}

type ChildrenObject struct {
	FirstName       *string `json:"first_name" form:"first_name"`
	LastName        *string `json:"last_name" form:"last_name"`
	ClassroomName   *string `json:"classroom_name" form:"classroom_name"`
	BirthCertNumber *string `json:"birth_certificate_number" form:"birth_certificate_number"`
	IsDelete        *bool   `json:"is_delete" form:"is_delete"`
}

type ContactItemsResponse struct {
	ID                   string                  `json:"id"`
	User                 *UserResponse           `json:"user"`
	School               *SchoolResponse         `json:"school"`
	Message              *string                 `json:"message"`
	Type                 string                  `json:"type"`
	Status               string                  `json:"status"`
	Files                *[]string               `json:"files"`
	Note                 *string                 `json:"note"`
	ClassroomName        *string                 `json:"classroom_name"`
	ParentPhone          *string                 `json:"parent_phone"`
	BirthCertNumber      *string                 `json:"birth_cert_number"`
	RelatedChildrenCount int                     `json:"related_children_count"`
	RelatedId            *string                 `json:"related_id"`
	CreatedAt            *time.Time              `json:"created_at"`
	UpdatedAt            *time.Time              `json:"updated_at"`
	Related              *ContactItemsResponse   `json:"related"`
	RelatedChildren      *[]ContactItemsResponse `json:"related_children"`
	UpdatedByUser        *UserResponse           `json:"updated_by"`
}

func (r *ContactItemsRequest) ToModel(m *ContactItems) error {
	if r.ID == nil {
		r.ID = new(string)
	}
	m.ID = *r.ID
	m.UserId = r.UserId
	m.SchoolId = r.SchoolId
	if r.Message == nil {
		r.Message = new(string)
	}
	m.Message = r.Message
	if r.ParentFirstName != nil && *r.ParentFirstName != "" {
		if r.ParentFirstName == nil {
			r.ParentFirstName = new(string)
		}
		if r.ParentLastName == nil {
			r.ParentLastName = new(string)
		}
		if r.ParentPhone == nil {
			r.ParentPhone = new(string)
		}
		if r.ParentPassportNumber == nil {
			r.ParentPassportNumber = new(string)
		}
		if r.ParentBirthday == nil {
			r.ParentBirthday = new(string)
		}
		*m.Message += "Meniň (ata-ene) maglumatlarymy girizmegiňizi haýyş edýärin: "
		*m.Message += fmt.Sprintf("\nAta-enäniň ady: %v", *r.ParentFirstName)
		*m.Message += fmt.Sprintf("\nAta-enäniň familýasy: %v", *r.ParentLastName)
		*m.Message += fmt.Sprintf("\nAta-enäniň telefony: %v", *r.ParentPhone)
		*m.Message += fmt.Sprintf("\nAta-enäniň passport seriýasy: %v", *r.ParentPassportNumber)
		*m.Message += fmt.Sprintf("\nAta-enäniň doglan güni: %v", *r.ParentBirthday)
	}
	if r.Children != nil && len(r.Children) > 0 {
		*m.Message += "Meniň çagalarymyň maglumatlaryny girizmegiňizi haýyş edýärin: "

		for k, v := range r.Children {
			if v.FirstName == nil {
				v.FirstName = new(string)
			}
			if v.LastName == nil {
				v.LastName = new(string)
			}
			if v.ClassroomName == nil {
				v.ClassroomName = new(string)
			}
			if v.BirthCertNumber == nil {
				v.BirthCertNumber = new(string)
			}
			*m.Message += "\n --- "
			if v.IsDelete != nil && *v.IsDelete {
				*m.Message += "POZMALY (meniň çagam däl): "
			}
			*m.Message += fmt.Sprintf("\n Okuwçynyň %v ady: %v", k+1, *v.FirstName)
			*m.Message += fmt.Sprintf("\n Okuwçynyň %v familýasy: %v", k+1, *v.LastName)
			*m.Message += fmt.Sprintf("\n Okuwçynyň %v synpy: %v", k+1, *v.ClassroomName)
			*m.Message += fmt.Sprintf("\n Okuwçynyň %v dog. şah.: %v", k+1, *v.BirthCertNumber)
			if m.BirthCertNumber != nil && *m.BirthCertNumber != "" {
				m.BirthCertNumber = v.BirthCertNumber
			}
			if m.ClassroomName != nil && *m.ClassroomName != "" {
				m.ClassroomName = v.ClassroomName
			}
		}
	}
	m.Type = r.Type
	m.Status = r.Status
	m.Files = r.Files
	m.RelatedId = r.RelatedId
	m.Note = r.Note
	m.ParentPhone = r.ParentPhone
	return nil
}

func (r *ContactItemsResponse) FromModel(m *ContactItems) error {
	r.ID = m.ID
	if m.User != nil {
		r.User = &UserResponse{}
		r.User.FromModel(m.User)
	}
	if m.School != nil {
		r.School = &SchoolResponse{}
		r.School.FromModel(m.School)
	}
	if m.UpdatedByUser != nil {
		r.UpdatedByUser = &UserResponse{}
		r.UpdatedByUser.FromModel(m.UpdatedByUser)
	}
	r.Message = m.Message
	r.Type = m.Type
	r.Status = m.Status
	r.Note = m.Note
	r.ParentPhone = m.ParentPhone
	r.ClassroomName = m.ClassroomName
	r.BirthCertNumber = m.BirthCertNumber
	r.RelatedChildrenCount = m.RelatedChildrenCount
	if m.Related != nil {
		r.Related = &ContactItemsResponse{}
		r.Related.FromModel(m.Related)
	}
	if m.RelatedChildren != nil {
		r.RelatedChildren = &[]ContactItemsResponse{}
		for _, v := range *m.RelatedChildren {
			item := ContactItemsResponse{}
			item.FromModel(v)
			*r.RelatedChildren = append(*r.RelatedChildren, item)
		}
	}
	if m.Files != nil {
		r.Files = &[]string{}
		for _, f := range *m.Files {
			*r.Files = append(*r.Files, fileUrl(&f))
		}
	}
	r.CreatedAt = m.CreatedAt
	r.UpdatedAt = m.UpdatedAt
	return nil
}

type ContactItemsFilterRequest struct {
	ID             *string   `json:"id" form:"id"`
	IDs            *[]string `json:"ids" form:"ids"`
	RelatedIds     *string   `json:"related_ids" form:"related_ids"`
	UserId         *string   `json:"user_id" form:"user_id"`
	SchoolId       *string   `json:"school_id" form:"school_id"`
	SchoolIds      *[]string `json:"school_ids" form:"school_ids"`
	UpdatedBy      *string   `json:"updated_by" form:"updated_by"`
	NotStatus      *string   `json:"not_status" form:"not_status"`
	Status         *string   `json:"status" form:"status"`
	Type           *string   `json:"type" form:"type"`
	Search         *string   `json:"search" form:"search"`
	OnlyNotRelated *bool     `json:"only_not_related" form:"only_not_related"` // TODO @sohbet
	StartDate      *string   `json:"start_date" form:"start_date"`
	EndDate        *string   `json:"end_date" form:"end_date"`
	PaginationRequest
}
