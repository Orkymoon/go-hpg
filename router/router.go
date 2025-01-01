package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/orkymoon/tripay-golang/controller"
)

func SetupRoutes(app *fiber.App) {

	api := app.Group("/api")

	api.Post("/pay", controller.HandlePayment)
	api.Post("/callback", controller.HandleCallback)
}
