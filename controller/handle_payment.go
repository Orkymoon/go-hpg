package controller

import (
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/orkymoon/tripay-golang/config"
	"github.com/orkymoon/tripay-golang/database"
	"github.com/orkymoon/tripay-golang/helper"
	"github.com/orkymoon/tripay-golang/model"
	"github.com/zakirkun/go-tripay/client"
	"github.com/zakirkun/go-tripay/utils"
)

func HandlePayment(c *fiber.Ctx) error {
	context := fiber.Map{
		"success": true,
		"message": "Payment has been made",
	}

	paymentData := new(model.Payment)

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
		case !helper.ValidateMacAddress(paymentData.MacAddress):
			context["success"] = false
			context["message"] = "Invalid MAC address format or empty"
			return c.Status(fiber.StatusUnauthorized).JSON(context)
		case !helper.Contains(allowedMethod, paymentData.PaymentMethod):
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
			Method:        utils.TRIPAY_CHANNEL(paymentData.PaymentMethod),
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

		transactionData := &model.Transaction{
			Reference:      tripayClosePayment.Data.Reference,
			MerchantRef:    tripayClosePayment.Data.MerchantRef,
			PaymentName:    tripayClosePayment.Data.PaymentName,
			PaymentMethod:  tripayClosePayment.Data.PaymentMethod,
			CustomerName:   tripayClosePayment.Data.CustomerName,
			CustomerEmail:  tripayClosePayment.Data.CustomerEmail,
			CustomerPhone:  tripayClosePayment.Data.CustomerPhone,
			Amount:         tripayClosePayment.Data.Amount,
			FeeMerchant:    tripayClosePayment.Data.FeeMerchant,
			FeeCustomer:    tripayClosePayment.Data.FeeCustomer,
			TotalFee:       tripayClosePayment.Data.TotalFee,
			AmountReceived: tripayClosePayment.Data.AmountReceived,
			CheckoutURL:    tripayClosePayment.Data.CheckoutURL,
			Status:         tripayClosePayment.Data.Status,
			ExpiredTime:    time.Unix(int64(tripayClosePayment.Data.ExpiredTime), 0),
		}

		paymentData.TransactionRef = tripayClosePayment.Data.Reference

		if tripayClosePayment.Success {
			result := database.DBConn.Create(transactionData).Error
			result2 := database.DBConn.Create(paymentData).Error
			if result != nil || result2 != nil {
				database.DBConn.Rollback()
				context["success"] = false
				context["message"] = "Error in database "
				return c.Status(fiber.ErrBadRequest.Code).JSON(context)
			}

			return c.Redirect(tripayClosePayment.Data.CheckoutURL, fiber.StatusFound)
		} else {
			context["success"] = false
			context["message"] = "Payment Failed"
			c.Status(400)
			return c.JSON(context)
		}
	}
}
