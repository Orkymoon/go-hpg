package model

import (
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	Reference      string    `json:"reference" query:"reference" gorm:"primaryKey;not null;size:100"`
	MerchantRef    string    `json:"merchant_ref" query:"merchant_ref" gorm:"not null;size:50"`
	PaymentName    string    `json:"payment_name" query:"payment_name" gorm:"not null;size:100"`
	PaymentMethod  string    `json:"payment_method" query:"payment_method" gorm:"not null;size:100"`
	CustomerName   string    `json:"customer_name" query:"customer_name" gorm:"size:100"`
	CustomerEmail  string    `json:"customer_email" query:"customer_email" gorm:"size:100"`
	CustomerPhone  string    `json:"customer_phone" query:"customer_phone" gorm:"size:20"`
	Amount         int       `json:"amount" query:"amount" gorm:"not null"`
	FeeMerchant    int       `json:"fee_merchant" query:"fee_merchant" gorm:"not null"`
	FeeCustomer    int       `json:"fee_customer" query:"fee_customer" gorm:"not null"`
	TotalFee       int       `json:"total_fee" query:"total_fee" gorm:"not null"`
	AmountReceived int       `json:"amount_received" query:"amount_received" gorm:"not null"`
	CheckoutURL    string    `json:"checkout_url" query:"checkout_url" gorm:"not null;size:100"`
	Status         string    `json:"status" query:"status" gorm:"not null;size:50"`
	ExpiredTime    time.Time `json:"expired_time" query:"expired_time" gorm:"size:50"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}
