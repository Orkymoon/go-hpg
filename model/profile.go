package model

import (
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type Profile struct {
	ID               uint   `gorm:"primarykey"`
	Name             string `json:"profile" form:"profile" gorm:"unique;not null;size:100"`
	SessionTimeout   string `json:"session_timeout" form:"session_timeout" gorm:"not null"`
	SharedUser       uint   `json:"shared_user" form:"shared_user" gorm:"not null"`
	RateLimit        string `json:"rate_limit" form:"rate_limit" gorm:"not null"`
	MacCookie        bool   `json:"mac_cookie" form:"mac_cookie" gorm:"not null"`
	MacCookieTimeout string `json:"mac_cookie_timeout" form:"mac_cookie_timeout" gorm:"not null"`
	Amount           uint   `json:"amount" form:"amount" gorm:"not null"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

func (t *Profile) Validate() error {
	validate := validator.New()
	return validate.Struct(t)
}
