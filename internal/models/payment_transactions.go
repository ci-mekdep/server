package models

import (
	"time"
)

type PaymentTariffType string

const PaymentFree PaymentTariffType = "free"
const PaymentPlus PaymentTariffType = "plus"
const PaymentTrial PaymentTariffType = "trial"
const PaymentUnlimited PaymentTariffType = "unlimited"

type PaymentStatus string

const PaymentStatusProcess PaymentStatus = "processing"
const PaymentStatusCompleted PaymentStatus = "completed"
const PaymentStatusFailed PaymentStatus = "failed"

type PaymentBank string

const PaymentBankHalkBank PaymentBank = "halkbank"
const PaymentBankRysgalBank PaymentBank = "rysgalbank"
const PaymentBankSenagatBank PaymentBank = "senagatbank"
const PaymentBankTfebBank PaymentBank = "tfeb"

var DefaultPaymentBank []PaymentBank = []PaymentBank{
	PaymentBankHalkBank,
	PaymentBankRysgalBank,
	PaymentBankSenagatBank,
	PaymentBankTfebBank,
}

func GetTariffName(code *string) *string {
	if code == nil {
		return nil
	}
	for _, v := range DefaultTariff {
		if v.Code == PaymentTariffType(*code) {
			return &v.Name
		}
	}
	return nil
}

var DefaultTariff []PaymentTariff = []PaymentTariff{
	{
		Code:           PaymentFree,
		Name:           "Adaty",
		Price:          (0.0),
		Description:    "Gündelik we çärýekler",
		WithoutPayment: true,
	},
	{
		Code:  PaymentPlus,
		Name:  "Göreldeli",
		Price: 5,
		MonthPrice: map[int]float64{
			1: float64(5),
			9: float64(36),
		},
		Description:    "Gündelik+ SMS habarlar, okuwçynyň gyzyklanýan dersleri, bilim analitikasy",
		WithoutPayment: false,
	},
	{
		Code:  PaymentTrial,
		Name:  "Göreldeli",
		Price: 5,
		MonthPrice: map[int]float64{
			1: float64(0),
		},
		Description: `Ilkinji 14 güni mugt, synap gör!

Akylly gündelik
SMS habarlar
Okuwçynyň gyzyklanýan dersleri
Bilim analitikasy`,
		WithoutPayment: true,
	},
	// {
	// 	Code:  PaymentUnlimited,
	// 	Name:  "Premium",
	// 	Price: (9.0),
	// 	MonthPrice: map[int]float64{
	// 		9: float64(72),
	// 	},
	// 	Description: "Gündelik+ SMS+ Widýo-sapaklar, okuw-oýunlary",
	// },
}

type PaymentTariff struct {
	Code              PaymentTariffType `json:"code"`
	Name              string            `json:"name"`
	Price             float64           `json:"price"`
	MonthPrice        map[int]float64   `json:"month_prices"`
	Description       string            `json:"description"`
	WithoutPayment    bool              `json:"without_payment"`
	AvailableChildIDs []string          `json:"available_child_ids"`
}

type PaymentTariffResponse struct {
	Code           PaymentTariffType `json:"code"`
	Name           string            `json:"name"`
	Price          Money             `json:"price"`
	MonthPrice     map[int]Money     `json:"month_prices"`
	Description    string            `json:"description"`
	WithoutPayment bool              `json:"without_payment"`
}

func (r *PaymentTariffResponse) FromModel(m PaymentTariff) {
	r.Code = m.Code
	r.Name = m.Name
	r.Description = m.Description
	r.WithoutPayment = m.WithoutPayment
	r.Price = MoneyFromFloat(m.Price)
	r.MonthPrice = map[int]Money{}
	for k, v := range m.MonthPrice {
		r.MonthPrice[k] = MoneyFromFloat(v)
	}
}

type PaymentTransaction struct {
	ID                 string            `json:"id"`
	SchoolId           string            `json:"school_id"`
	PayerId            string            `json:"payer_id"`
	UserIds            []string          `json:"user_ids"`
	TariffType         PaymentTariffType `json:"tariff_type"`
	Status             PaymentStatus     `json:"status"`
	Amount             float64           `json:"amount"`
	OriginalAmount     float64           `json:"original_amount"`
	BankType           PaymentBank       `json:"bank_type"`
	CardName           *string           `json:"card_name"`
	OrderNumber        *string           `json:"order_number"`
	OrderUrl           *string           `json:"order_url"`
	Comment            *string           `json:"comment"`
	SystemComment      *string           `json:"system_comment"`
	UnitPrice          *int              `json:"unit_price"`
	SchoolPrice        *int              `json:"school_price"`
	CenterPrice        *int              `json:"center_price"`
	DiscountPrice      *int              `json:"discount_price"`
	UsedDays           *int              `json:"used_days"`
	UsedDaysPrice      *int              `json:"used_days_price"`
	SchoolMonths       int               `json:"school_months"`
	CenterMonths       int               `json:"center_months"`
	SchoolClassroomIds []string          `json:"school_classroom_ids"`
	CenterClassroomIds []string          `json:"center_classroom_ids"`
	UpdatedAt          *time.Time        `json:"updated_at"`
	CreatedAt          *time.Time        `json:"created_at"`
	School             *School           `json:"school"`
	Payer              *User             `json:"payer"`
	Classrooms         []*Classroom      `json:"classrooms"`
	Students           []User            `json:"students"`
}

func (PaymentTransaction) RelationFields() []string {
	return []string{"school", "payer", "classrooms", "students"}
}

func (pt *PaymentTransaction) IsStatusProcessing() bool {
	return pt.Status == PaymentStatusProcess
}

type PaymentTransactionResponse struct {
	ID             string               `json:"id"`
	TariffType     PaymentTariffType    `json:"tariff_type"`
	Month          int                  `json:"month"`
	Status         PaymentStatus        `json:"status"`
	Amount         float64              `json:"amount"`
	OriginalAmount float64              `json:"original_amount"`
	BankType       PaymentBank          `json:"bank_type"`
	CardName       *string              `json:"card_name"`
	OrderNumber    *string              `json:"order_number"`
	OrderUrl       *string              `json:"order_url"`
	Comment        *string              `json:"comment"`
	CreatedAt      *time.Time           `json:"created_at"`
	Payer          *UserResponse        `json:"payer"`
	School         *SchoolResponse      `json:"school"`
	Classrooms     *[]ClassroomResponse `json:"classrooms"`
	Students       *[]UserResponse      `json:"students"`
}

func (pt *PaymentTransactionResponse) FromModel(m *PaymentTransaction) {
	pt.ID = m.ID
	pt.TariffType = m.TariffType
	if m.SchoolMonths != 0 {
		pt.Month = m.SchoolMonths
	} else {
		pt.Month = m.CenterMonths
	}
	pt.Status = m.Status
	pt.Amount = m.Amount
	pt.OriginalAmount = m.OriginalAmount
	pt.BankType = m.BankType
	pt.CardName = m.CardName
	pt.OrderNumber = m.OrderNumber
	pt.OrderUrl = m.OrderUrl
	pt.Comment = m.Comment
	if m.SystemComment != nil {
		if pt.Comment == nil {
			pt.Comment = new(string)
		}
		*pt.Comment += ` Bank: ` + *m.SystemComment
	}
	pt.CreatedAt = m.CreatedAt
	if m.Payer != nil {
		resItem := UserResponse{}
		resItem.FromModel(m.Payer)
		pt.Payer = &resItem
	}
	if m.School != nil {
		resItem := SchoolResponse{}
		resItem.FromModel(m.School)
		pt.School = &resItem
	}
	if m.Classrooms != nil {
		pt.Classrooms = &[]ClassroomResponse{}
		for _, v := range m.Classrooms {
			resItem := ClassroomResponse{}
			resItem.FromModel(v)
			*pt.Classrooms = append(*pt.Classrooms, resItem)
		}
	}
	if m.Students != nil {
		pt.Students = &[]UserResponse{}
		for _, v := range m.Students {
			resItem := UserResponse{}
			resItem.FromModel(&v)
			*pt.Students = append(*pt.Students, resItem)
		}
	}
}

type PaymentTransactionFilterRequest struct {
	ID         *string   `json:"id" form:"id"`
	Ids        *[]string `json:"ids" form:"ids[]"`
	SchoolId   *string   `json:"school_id" form:"school_id"`
	SchoolIds  *[]string `json:"school_ids" form:"school_ids[]"`
	PayerId    *string   `json:"payer_id" form:"payer_id"`
	StudentId  *string   `json:"student_id" form:"student_id"`
	TariffType *string   `json:"tariff_type" form:"tariff_type"`
	BankType   *string   `json:"bank_type" form:"bank_type"`
	Status     *string   `json:"status" form:"status"`
	StartDate  *string   `json:"start_date" form:"start_date"`
	EndDate    *string   `json:"end_date" form:"end_date"`
	PaginationRequest
}

type PaymentCheckoutRequest struct {
	CenterClassroomIds []string          `json:"center_classroom_ids"`
	SchoolClassroomIds []string          `json:"school_classroom_ids"`
	StudentIds         []string          `json:"student_ids"`
	SchoolMonths       int               `json:"school_months"`
	CenterMonths       int               `json:"center_months"`
	TariffType         PaymentTariffType `json:"tariff_type"`
	BankType           PaymentBank       `json:"bank_type"`
	CardName           string            `json:"card_name"`
	Comment            string            `json:"comment"`
}

type PaymentConfirmRequest struct {
	Code          string `json:"code"`
	TransactionID string `json:"transaction_id"`
}

type PaymentCalculationResponse struct {
	UnitPrice     Money `json:"unit_price"`
	SchoolPrice   Money `json:"school_price"`
	CenterPrice   Money `json:"center_price"`
	DiscountPrice Money `json:"discount_price"`
	UsedDaysPrice Money `json:"used_days_price"`
	UsedDays      int   `json:"used_days"`
	Total         Money `json:"total"`
	IsDateChange  bool  `json:"is_date_changes"`
}
