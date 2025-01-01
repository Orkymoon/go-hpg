package model

type Transaction struct {
	Reference     string `json:"reference" xml:"reference" form:"reference" query:"reference" gorm:"primaryKey;not null;size:100"`
	MerchantRef   string `json:"merchant_ref" xml:"merchant_ref" form:"merchant_ref" query:"merchant_ref" gorm:"not null;size:50"`
	Voucher       string `json:"voucher" query:"voucher" gorm:"not null;size:25"`
	Profile       string `json:"profile" query:"profile" gorm:"not null;size:50"`
	PaymentMethod string `json:"payment_method" query:"payment_method" gorm:"not null;size:100"`
	Amount        uint   `json:"amount" query:"amount" gorm:"not null"`
	CheckoutURL   string `json:"checkout_url" query:"checkout_url" gorm:"not null;size:100"`
	Status        string `json:"status" query:"status" gorm:"not null;size:50"`
	Mac           string `json:"mac" query:"mac" gorm:"not null;size:50"`
	CustomerName  string `json:"customer_name" query:"customer_name" gorm:"size:100"`
	CustomerEmail string `json:"customer_email" query:"customer_email" gorm:"size:100"`
	CustomerPhone string `json:"customer_phone" query:"customer_phone" gorm:"size:20"`
}
