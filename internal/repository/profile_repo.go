package repository

import (
	"errors"

	"github.com/orkymoon/tripay-golang/internal/database"
	"github.com/orkymoon/tripay-golang/internal/model"
	"gorm.io/gorm"
)

func GetProfileByName(name string) (*model.Profile, error) {
	var profile model.Profile

	if err := database.DBConn.Select("name", "amount").Where("name = ? ", name).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return &profile, nil
}
