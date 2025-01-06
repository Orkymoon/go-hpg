package repository

import (
	"errors"

	"github.com/orkymoon/tripay-golang/internal/database"
	"github.com/orkymoon/tripay-golang/internal/model"
	"gorm.io/gorm"
)

func SavePayment(payment model.Payment) error {
	if err := database.DBConn.Create(&payment).Error; err != nil {
		return err
	}
	return nil
}

func GetPaymentByReferenceWithProfile(reference string) (*model.Payment, error) {
	var payment model.Payment

	if err := database.DBConn.Preload("Profile").Where("transaction_ref = ? ", reference).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return &payment, nil
}
