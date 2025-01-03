package migrations

import (
	"log"

	"github.com/orkymoon/tripay-golang/database"
	"github.com/orkymoon/tripay-golang/model"
)

func Migration() {
	err := database.DBConn.AutoMigrate(
		&model.Transaction{},
		&model.Payment{},
	)

	if err != nil {
		log.Panic("Database migration failed")
	}

	log.Println("Migration successfuly.")

}
