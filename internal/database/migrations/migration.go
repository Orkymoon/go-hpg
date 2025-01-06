package migrations

import (
	"log"

	"github.com/orkymoon/tripay-golang/internal/database"
	"github.com/orkymoon/tripay-golang/internal/model"
)

func Migration() {
	err := database.DBConn.AutoMigrate(
		&model.Transaction{},
		&model.Payment{},
		&model.Profile{},
	)

	if err != nil {
		log.Panic("Database migration failed")
	}

	log.Println("Migration successfuly.")

}
