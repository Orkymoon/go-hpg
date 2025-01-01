package database

import (
	"log"

	"github.com/orkymoon/tripay-golang/config"
	"github.com/orkymoon/tripay-golang/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DBConn *gorm.DB

func ConnectDB() {

	dsn := config.MysqlUser + ":" + config.MysqlPassword + "@tcp(" + config.MysqlHost + ":" + config.MysqlPort + ")/" + config.MysqlDbname + "?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})

	if err != nil {
		panic("Database connection failed.")
	}

	log.Println("Connection Successfuly.")

	db.AutoMigrate(new(model.Transaction))

	DBConn = db
}