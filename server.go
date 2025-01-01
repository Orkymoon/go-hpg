package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/orkymoon/tripay-golang/config"
	"github.com/orkymoon/tripay-golang/database"
	"github.com/orkymoon/tripay-golang/router"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error in loading .env file.")
	}
	config.LoadEnv()
	database.ConnectDB()
}

func main() {

	sqlDB, err := database.DBConn.DB()

	if err != nil {
		panic("Error in sql connection.")
	}

	defer sqlDB.Close()

	app := fiber.New()
	app.Use(logger.New())

	router.SetupRoutes(app)

	log.Printf("running on http://localhost%v", config.AppPort)
	app.Listen(config.AppPort)
}
