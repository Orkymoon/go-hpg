package controller

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/orkymoon/tripay-golang/config"
	"github.com/orkymoon/tripay-golang/database"
	"github.com/orkymoon/tripay-golang/helper"
	"github.com/orkymoon/tripay-golang/model"
	"github.com/zakirkun/go-tripay/client"
	"github.com/zakirkun/go-tripay/utils"
)

type PaymentForm struct {
	Phone      uint   `form:"phone"`
	Voucher    string `form:"vc"`
	Profile    string `form:"profile"`
	Amount     int    `form:"saldo"`
	MacAddress string `form:"mac"`
	Method     string `form:"rek"`
}

func HandlePayment(c *fiber.Ctx) error {
	context := fiber.Map{
		"success": true,
		"message": "Payment has been made",
	}

	paymentData := new(PaymentForm)

	tripay := client.Client{
		ApiKey:       config.TripayApiKey,
		PrivateKey:   config.TripayPrivateKey,
		MerchantCode: config.TripayMerchantCode,
		Mode:         utils.TRIPAY_MODE(config.TripayMode),
	}

	// callbackSignature := c.Get("User-Agent")
	// log.Println(callbackSignature)

	if err := c.BodyParser(paymentData); err != nil {
		log.Println("Error in parsing request:", err)
		context["success"] = false
		context["message"] = "Invalid request data: " + err.Error()
		return c.Status(fiber.ErrBadRequest.Code).JSON(context)
	} else {
		phoneStr := helper.ValidatePhoneNumber(paymentData.Phone)

		tripayPaymentChannel, err := tripay.MerchantPay()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		var allowedMethod []string
		if tripayPaymentChannel.Success {
			for _, merchant := range tripayPaymentChannel.Data {
				allowedMethod = append(allowedMethod, merchant.Code)
			}
		} else {
			return c.Status(fiber.StatusInternalServerError).SendString(tripayPaymentChannel.Message)
		}

		switch {
		case strings.TrimSpace(phoneStr) == "":
			context["success"] = false
			context["message"] = "Invalid phone number"
			return c.Status(fiber.StatusUnauthorized).JSON(context)
		case strings.TrimSpace(paymentData.Voucher) == "":
			context["success"] = false
			context["message"] = "Voucher is required"
			return c.Status(fiber.StatusUnauthorized).JSON(context)
		case strings.TrimSpace(paymentData.Profile) == "":
			context["success"] = false
			context["message"] = "Profile is required"
			return c.Status(fiber.StatusUnauthorized).JSON(context)
		case paymentData.Amount <= 0:
			context["success"] = false
			context["message"] = "Amount must be greater than 0"
			return c.Status(fiber.StatusUnauthorized).JSON(context)
		case paymentData.MacAddress != "" && !helper.ValidateMacAddress(paymentData.MacAddress):
			context["success"] = false
			context["message"] = "Invalid MAC address format"
			return c.Status(fiber.StatusUnauthorized).JSON(context)
		case !helper.Contains(allowedMethod, paymentData.Method):
			context["success"] = false
			context["message"] = "Invalid payment method. Allowed values are: " + strings.Join(allowedMethod, ", ")
			return c.Status(fiber.StatusUnauthorized).JSON(context)
		}

		merchantReff := helper.GenerateMerchantReference()

		tripay.SetSignature(utils.Signature{
			Amount:       int64(paymentData.Amount),
			PrivateKey:   string(config.TripayPrivateKey),
			MerchantCode: string(config.TripayMerchantCode),
			MerchanReff:  merchantReff,
		})

		bodyReq := client.ClosePaymentBodyRequest{
			Method:        utils.TRIPAY_CHANNEL(paymentData.Method),
			MerchantRef:   merchantReff,
			Amount:        int(paymentData.Amount),
			CustomerName:  "Lindinet",
			CustomerEmail: "lindinet@gmail.com",
			CustomerPhone: string(phoneStr),
			ReturnURL:     "http://" + config.MikrotikDnsHotspot + "/login?dst=&username=" + paymentData.Voucher + "&password=" + paymentData.Voucher,
			ExpiredTime:   client.SetTripayExpiredTime(8), // 8 Hour
			Signature:     tripay.GetSignature(),
			OrderItems: []client.OrderItemClosePaymentRequest{
				{
					SKU:      string(paymentData.Profile),
					Name:     string(paymentData.Voucher),
					Price:    int(paymentData.Amount),
					Quantity: 1,
					// ProductURL: "https://producturl.com",
					// ImageURL:   "https://imageurl.com",
				},
			},
		}

		tripayClosePayment, err := tripay.ClosePaymentRequestTransaction(bodyReq)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		var (
			checkoutUrl   = tripayClosePayment.Data.CheckoutURL
			reference     = tripayClosePayment.Data.Reference
			status        = tripayClosePayment.Data.Status
			customerName  = tripayClosePayment.Data.CustomerName
			customerEmail = tripayClosePayment.Data.CustomerEmail
			customerPhone = tripayClosePayment.Data.CustomerPhone
			merchantRef   = tripayClosePayment.Data.MerchantRef
		)

		record := &model.Transaction{
			Reference:     reference,
			MerchantRef:   merchantRef,
			Voucher:       paymentData.Voucher,
			Profile:       paymentData.Profile,
			PaymentMethod: paymentData.Method,
			Amount:        uint(paymentData.Amount),
			CheckoutURL:   checkoutUrl,
			Status:        status,
			Mac:           paymentData.MacAddress,
			CustomerName:  customerName,
			CustomerEmail: customerEmail,
			CustomerPhone: customerPhone,
		}

		if tripayClosePayment.Success {
			result := database.DBConn.Create(record)
			if result.Error != nil {
				context["success"] = false
				context["message"] = "Error in database "
				return c.Status(fiber.ErrBadRequest.Code).JSON(context)
			}

			return c.Redirect(checkoutUrl, fiber.StatusFound)
		} else {
			context["success"] = false
			context["message"] = "Payment Failed"
			c.Status(400)
			return c.JSON(context)
		}
	}
}
