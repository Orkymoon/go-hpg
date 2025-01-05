package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/orkymoon/tripay-golang/controller"
)

func SetupRoutes(app *fiber.App) {

	api := app.Group("/api")

	api.Post("/payment", controller.PaymentCreate)
	api.Post("/payment/callback", controller.PaymentCallback)
}
