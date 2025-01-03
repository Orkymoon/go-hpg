package model

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID             uint   `gorm:"primarykey"`
	Voucher        string `json:"voucher" form:"voucher" query:"voucher"`
	Profile        string `json:"profile" form:"profile" query:"profile"`
	MacAddress     string `json:"mac_address" form:"mac_address" query:"mac_address"`
	Phone          uint   `json:"phone" form:"phone" query:"phone"`
	Amount         int    `json:"amount" form:"amount" query:"amount"`
	PaymentMethod  string `json:"payment_method" form:"payment_method" query:"payment_method"`
	TransactionRef string
	Transaction    Transaction `gorm:"foreignKey:TransactionRef;references:Reference"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}
