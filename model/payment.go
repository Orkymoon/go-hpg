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
