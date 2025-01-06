package model

import (
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	ID                   uint      `gorm:"primarykey"`
	Reference            string    `json:"reference" query:"reference" gorm:"unique;not null;size:100"`
	MerchantRef          string    `json:"merchant_ref" query:"merchant_ref" gorm:"not null;size:50"`
	PaymentSelectionType string    `json:"payment_selection_type" query:"payment_selection_type" gorm:"size:10"`
	PaymentName          string    `json:"payment_name" query:"payment_name" gorm:"not null;size:100"`
	TotalAmount          int64     `json:"total_amount" query:"total_amount" gorm:"not null"`
	FeeMerchant          int       `json:"fee_merchant" query:"fee_merchant" gorm:"not null"`
	FeeCustomer          int       `json:"fee_customer" query:"fee_customer" gorm:"not null"`
	TotalFee             int       `json:"total_fee" query:"total_fee" gorm:"not null"`
	AmountReceived       int       `json:"amount_received" query:"amount_received" gorm:"not null"`
	IsClosedPayment      bool      `json:"is_closed_payment" query:"is_closed_payment" gorm:"default:1"`
	CheckoutURL          string    `json:"checkout_url" query:"checkout_url" gorm:"not null;size:100"`
	Status               string    `json:"status" query:"status" gorm:"not null;size:10"`
	PaidAt               time.Time `json:"paid_at" query:"paid_at" gorm:"size:50;default:null"`
	ExpiredTime          time.Time `json:"expired_time" query:"expired_time" gorm:"size:50"`
	Note                 string    `json:"note" query:"note" gorm:"default:null;size:255"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            gorm.DeletedAt `gorm:"index"`
}
