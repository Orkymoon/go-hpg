package controller

import (
	"encoding/json"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/orkymoon/tripay-golang/config"
	"github.com/orkymoon/tripay-golang/helper"
	"github.com/orkymoon/tripay-golang/model"
	"github.com/orkymoon/tripay-golang/repository"
	"github.com/orkymoon/tripay-golang/service"
	"github.com/zakirkun/go-tripay/client"
	"github.com/zakirkun/go-tripay/utils"
)

func CreatePayment(c *fiber.Ctx) error {

	paymentRequest := new(model.PaymentRequest)

	if err := c.BodyParser(paymentRequest); err != nil {
		log.Printf("Error parsing request: %v", err)
		return helper.ReturnCustomResponse(c, fiber.StatusBadRequest, false, "Invalid request data: "+err.Error())
	}

	if err := validator.New().Struct(paymentRequest); err != nil {
		log.Printf("Validation failed: %v", err)
		return helper.ReturnCustomResponse(c, fiber.StatusBadRequest, false, err.Error())
	}

	profile, err := repository.GetProfileByName(paymentRequest.Profile)
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, "Profile not found or database error")
	}

	tripay := client.Client{
		ApiKey:       config.TripayApiKey,
		PrivateKey:   config.TripayPrivateKey,
		MerchantCode: config.TripayMerchantCode,
		Mode:         utils.TRIPAY_MODE(config.TripayMode),
	}

	bodyReq := service.BodyRequestService(paymentRequest, profile, &tripay)
	tripayClosePayment, err := tripay.ClosePaymentRequestTransaction(*bodyReq)
	if err != nil {
		log.Printf("Error processing payment with the gateway: %v", err)
		return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, "Error processing payment with the gateway: "+err.Error())
	}

	if !tripayClosePayment.Success {
		log.Println("Tripay payment gateway error: ", tripayClosePayment.Message)
		return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, tripayClosePayment.Message)
	}

	transaction := &model.Transaction{
		Reference:            tripayClosePayment.Data.Reference,
		MerchantRef:          tripayClosePayment.Data.MerchantRef,
		PaymentSelectionType: tripayClosePayment.Data.PaymentSelectionType,
		PaymentName:          tripayClosePayment.Data.PaymentName,
		TotalAmount:          int64(tripayClosePayment.Data.Amount),
		FeeMerchant:          tripayClosePayment.Data.FeeMerchant,
		FeeCustomer:          tripayClosePayment.Data.FeeCustomer,
		TotalFee:             tripayClosePayment.Data.TotalFee,
		AmountReceived:       tripayClosePayment.Data.AmountReceived,
		CheckoutURL:          tripayClosePayment.Data.CheckoutURL,
		Status:               tripayClosePayment.Data.Status,
		ExpiredTime:          helper.UnixToTime(int64(tripayClosePayment.Data.ExpiredTime)),
	}

	if err := repository.SaveTransaction(*transaction); err != nil {
		log.Printf("error in querying database err: %v", err)
		return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, err.Error())
	}

	payment := &model.Payment{
		Hostname:       paymentRequest.Hostname,
		ServerName:     paymentRequest.ServerName,
		MacAddress:     paymentRequest.Mac,
		IpAddress:      paymentRequest.IpAddress,
		Voucher:        paymentRequest.Voucher,
		CustomerName:   paymentRequest.CustomerName,
		CustomerEmail:  paymentRequest.CustomerEmail,
		CustomerPhone:  paymentRequest.CustomerPhone,
		PaymentMethod:  paymentRequest.PaymentMethod,
		ProfileRef:     paymentRequest.Profile,
		TransactionRef: tripayClosePayment.Data.Reference,
	}
	if err := repository.SavePayment(*payment); err != nil {
		log.Printf("error in querying database err: %v", err)
		return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, err.Error())
	}

	// Redirect to the Tripay checkout URL
	return c.Redirect(tripayClosePayment.Data.CheckoutURL, fiber.StatusFound)
}

func GetPayment(c *fiber.Ctx) error {
	return c.JSON("Under construction")
}

func UpdatePayment(c *fiber.Ctx) error {
	return c.JSON("Under construction")
}

func DeletePayment(c *fiber.Ctx) error {
	return c.JSON("Under construction")
}

func PaymentCallback(c *fiber.Ctx) error {

	received := c.Body()

	signature1 := helper.CallbackSignature(received, config.TripayPrivateKey)
	signature2 := c.Get("X-Callback-Signature")

	if !helper.CompareSignature(signature1, signature2) {
		log.Printf("Invalid signature: %v != %v", signature1, signature2)
		return helper.ReturnCustomResponse(c, fiber.StatusUnauthorized, false, "Invalid signature")
	}

	if c.Get("X-Callback-Event") != "payment_status" {
		log.Println("Unrecognized callback event")
		return helper.ReturnCustomResponse(c, fiber.StatusUnauthorized, false, "Unrecognized callback event")
	}

	var callback model.Callback
	if err := json.Unmarshal(received, &callback); err != nil {
		log.Printf("Error failed to decode callback data to JSON : %v ", err)
		return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, "Failed to decode JSON data")
	}

	switch callback.IsClosedPayment {
	case 0:
		log.Println("Error someone attempted to use an incorrect payment method")
		return helper.ReturnCustomResponse(c, fiber.StatusBadRequest, false, "Invalid Payment Method")
	case 1:
		transaction, err := repository.GetTransactionByReference(callback.Reference)
		if err != nil {
			log.Printf("Database Error: %v", err)
			return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, "Database Error")
		}
		switch transaction.Status {
		case "PAID":
			log.Println("Invoice already paid")
			return helper.ReturnCustomResponse(c, fiber.StatusBadRequest, false, "Invoice already paid")
		default:
			if err := service.PaymentStatus(callback); err != nil {
				log.Printf("Error processing payment status: %v", err)
				return helper.ReturnCustomResponse(c, fiber.StatusBadRequest, false, err.Error())
			}
		}
	}

	log.Println("Payment successfully processed")
	return helper.ReturnCustomResponse(c, fiber.StatusCreated, true, "Payment successfully processed")
}
