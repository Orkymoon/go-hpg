package repository

import (
	"errors"

	"github.com/orkymoon/tripay-golang/internal/database"
	"github.com/orkymoon/tripay-golang/internal/model"
	"gorm.io/gorm"
)

func SaveTransaction(transaction model.Transaction) error {
	if err := database.DBConn.Create(&transaction).Error; err != nil {
		return err
	}
	return nil
}

func GetTransactionByReference(reference string) (*model.Transaction, error) {
	var transaction model.Transaction

	if err := database.DBConn.Select("status").Where("reference = ? ", reference).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return &transaction, nil
}
func UpdateTransactionByReference(reference string, transaction *model.Transaction) error {
	if err := database.DBConn.Model(&model.Transaction{}).Where("reference = ?", reference).Updates(&transaction).Error; err != nil {
		return err
	}
	return nil
}
