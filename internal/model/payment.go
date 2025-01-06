package model

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID             uint   `gorm:"primarykey"`
	Hostname       string `gorm:"not null;size:50"`
	ServerName     string `gorm:"not null;size:25"`
	MacAddress     string `json:"mac_address" form:"mac_address" query:"mac_address" validate:"mac"`
	IpAddress      string `gorm:"not null"`
	Voucher        string `json:"voucher" form:"voucher" query:"voucher" validate:"required"`
	CustomerName   string `gorm:"not null"`
	CustomerEmail  string `gorm:"not null"`
	CustomerPhone  string `gorm:"not null"`
	PaymentMethod  string `json:"payment_method" form:"payment_method" query:"payment_method" gorm:"not null" validate:"required"`
	ProfileRef     string
	TransactionRef string
	Profile        Profile     `gorm:"foreignKey:ProfileRef;references:Name"`
	Transaction    Transaction `gorm:"foreignKey:TransactionRef;references:Reference"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

type PaymentRequest struct {
	Hostname      string `form:"hostname" validate:"required,hostname"`
	ServerName    string `form:"server_name" validate:"required"`
	Mac           string `form:"mac_address" validate:"required,mac"`
	IpAddress     string `form:"ip" validate:"required,ip"`
	Voucher       string `form:"voucher" validate:"required"`
	Profile       string `form:"profile" validate:"required"`
	PaymentMethod string `form:"payment_method" validate:"required"`
	CustomerName  string `form:"customer_name" validate:"required"`
	CustomerEmail string `form:"customer_email" validate:"required,email"`
	CustomerPhone string `form:"customer_phone" validate:"required,e164"`
}

type Callback struct {
	Reference         string  `json:"reference"`
	MerchantRef       string  `json:"merchant_ref"`
	PaymentMethod     string  `json:"payment_method"`
	PaymentMethodCode string  `json:"payment_method_code"`
	TotalAmount       int64   `json:"total_amount"`
	FeeMerchant       int64   `json:"fee_merchant"`
	FeeCustomer       int64   `json:"fee_customer"`
	TotalFee          int64   `json:"total_fee"`
	AmountReceived    int64   `json:"amount_received"`
	IsClosedPayment   int     `json:"is_closed_payment"`
	Status            string  `json:"status"`
	PaidAt            int64   `json:"paid_at"`
	Note              *string `json:"note"`
}
