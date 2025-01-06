package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/orkymoon/tripay-golang/internal/controller"
)

func SetupRoutes(app *fiber.App) {

	api := app.Group("/api")

	api.Post("/payment", controller.CreatePayment)
	api.Post("/payment/callback", controller.PaymentCallback)
}
