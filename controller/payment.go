package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-routeros/routeros/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/orkymoon/tripay-golang/config"
	"github.com/orkymoon/tripay-golang/database"
	"github.com/orkymoon/tripay-golang/helper"
	"github.com/orkymoon/tripay-golang/model"
	"github.com/zakirkun/go-tripay/client"
	"github.com/zakirkun/go-tripay/utils"
)

type PaymentPost struct {
	Hostname      string `form:"hostname" validate:"required,hostname"`
	ServerName    string `form:"server_name" validate:"required"`
	Mac           string `form:"mac_address" validate:"required,mac"`
	IpAddress     string `form:"ip" validate:"required,ip"`
	Voucher       string `form:"voucher" validate:"required"`
	Profile       string `form:"profile" validate:"required"`
	PaymentMethod string `form:"payment_method" validate:"required"`
	CustomerName  string `form:"customer_name" validate:"required"`
	CustomerEmail string `form:"customer_email" validate:"required,email"`
	CustomerPhone string `form:"customer_phone" validate:"required,e164"`
}

type CallbackStruct struct {
	Reference         string  `json:"reference"`
	MerchantRef       string  `json:"merchant_ref"`
	PaymentMethod     string  `json:"payment_method"`
	PaymentMethodCode string  `json:"payment_method_code"`
	TotalAmount       int64   `json:"total_amount"`
	FeeMerchant       int64   `json:"fee_merchant"`
	FeeCustomer       int64   `json:"fee_customer"`
	TotalFee          int64   `json:"total_fee"`
	AmountReceived    int64   `json:"amount_received"`
	IsClosedPayment   int     `json:"is_closed_payment"`
	Status            string  `json:"status"`
	PaidAt            int64   `json:"paid_at"`
	Note              *string `json:"note"`
}

func PaymentCreate(c *fiber.Ctx) error {

	received := new(PaymentPost)
	if err := c.BodyParser(received); err != nil {
		log.Printf("Error parsing request: %v", err)
		return helper.ReturnCustomResponse(c, fiber.StatusBadRequest, false, "Invalid request data: "+err.Error())
	}

	if err := validator.New().Struct(received); err != nil {
		log.Printf("Validation failed: %v", err)
		return helper.ReturnCustomResponse(c, fiber.StatusBadRequest, false, err.Error())
	}

	merchantReff := helper.GenerateMerchantReference()
	var record model.Profile
	if err := database.DBConn.Select("name", "amount").Where("name = ?", received.Profile).First(&record).Error; err != nil {
		log.Printf("Error querying database: %v", err)
		return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, "Profile not found or database error")
	}

	tripay := client.Client{
		ApiKey:       config.TripayApiKey,
		PrivateKey:   config.TripayPrivateKey,
		MerchantCode: config.TripayMerchantCode,
		Mode:         utils.TRIPAY_MODE(config.TripayMode),
	}
	tripay.SetSignature(utils.Signature{
		Amount:       int64(record.Amount),
		PrivateKey:   string(config.TripayPrivateKey),
		MerchantCode: string(config.TripayMerchantCode),
		MerchanReff:  merchantReff,
	})

	bodyReq := client.ClosePaymentBodyRequest{
		Method:        utils.TRIPAY_CHANNEL(received.PaymentMethod),
		MerchantRef:   merchantReff,
		Amount:        int(record.Amount),
		CustomerName:  received.CustomerName,
		CustomerEmail: received.CustomerEmail,
		CustomerPhone: received.CustomerPhone,
		ReturnURL:     "http://" + received.Hostname + "/login?dst=&username=" + received.Voucher + "&password=" + received.Voucher,
		ExpiredTime:   client.SetTripayExpiredTime(config.TripayExpiredTime), // 8 hours expiration
		Signature:     tripay.GetSignature(),
		OrderItems: []client.OrderItemClosePaymentRequest{
			{
				SKU:      record.Name,
				Name:     received.Voucher,
				Price:    int(record.Amount),
				Quantity: 1,
			},
		},
	}

	tripayClosePayment, err := tripay.ClosePaymentRequestTransaction(bodyReq)
	if err != nil {
		log.Printf("Error processing payment with the gateway: %v", err)
		return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, "Error processing payment with the gateway: "+err.Error())
	}

	if !tripayClosePayment.Success {
		log.Println("Tripay payment gateway error: ", tripayClosePayment.Message)
		return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, tripayClosePayment.Message)
	}

	tx := database.DBConn.Begin()
	if tx.Error != nil {
		return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, "Error starting transaction")
	}

	// Goroutine
	errChan := make(chan error, 2)

	go func() {
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

		if err := tx.Create(transaction).Error; err != nil {
			errChan <- fmt.Errorf("error saving transaction: %v", err)
			return
		}
		errChan <- nil
	}()

	go func() {
		payment := &model.Payment{
			Hostname:       received.Hostname,
			ServerName:     received.ServerName,
			MacAddress:     received.Mac,
			IpAddress:      received.IpAddress,
			Voucher:        received.Voucher,
			CustomerName:   received.CustomerName,
			CustomerEmail:  received.CustomerEmail,
			CustomerPhone:  received.CustomerPhone,
			PaymentMethod:  received.PaymentMethod,
			ProfileRef:     received.Profile,
			TransactionRef: tripayClosePayment.Data.Reference,
		}
		time.Sleep(5 * time.Millisecond)

		if err := tx.Create(payment).Error; err != nil {
			errChan <- fmt.Errorf("error saving payment: %v", err)
			return
		}
		errChan <- nil
	}()

	if transactionErr := <-errChan; transactionErr != nil {
		tx.Rollback()
		return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, transactionErr.Error())
	}

	if paymentErr := <-errChan; paymentErr != nil {
		tx.Rollback()
		return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, paymentErr.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, "Error committing transaction: "+err.Error())
	}

	// Redirect to the Tripay checkout URL
	return c.Redirect(tripayClosePayment.Data.CheckoutURL, fiber.StatusFound)
}

func PaymentRead(c *fiber.Ctx) error {
	return c.JSON("Under construction")
}

func PaymentUpdate(c *fiber.Ctx) error {
	return c.JSON("Under construction")
}

func PaymentDelete(c *fiber.Ctx) error {
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

	var callback CallbackStruct
	if err := json.Unmarshal(received, &callback); err != nil {
		log.Printf("Error failed to decode callback data to JSON : %v ", err)
		return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, "Failed to decode JSON data")
	}

	switch callback.IsClosedPayment {
	case 0:
		log.Println("Invalid Payment Method")
		return helper.ReturnCustomResponse(c, fiber.StatusBadRequest, false, "Only CLOSED PAYMENT")
	case 1:
		var payment model.Payment
		if err := database.DBConn.Preload("Transaction", "reference = ? AND merchant_ref = ?", callback.Reference, callback.MerchantRef).Preload("Profile").Where("transaction_ref = ?", callback.Reference).First(&payment).Error; err != nil {
			log.Printf("Database Error: %v", err)
			return helper.ReturnCustomResponse(c, fiber.StatusInternalServerError, false, "Database Error")
		}
		switch payment.Transaction.Status {
		case "PAID":
			log.Println("Invoice already paid")
			return helper.ReturnCustomResponse(c, fiber.StatusBadRequest, false, "Invoice already paid")
		default:
			if err := paymentStatus(callback, payment); err != nil {
				log.Printf("Error processing payment status: %v", err)
				return helper.ReturnCustomResponse(c, fiber.StatusBadRequest, false, err.Error())
			}
		}
	}

	log.Println("Payment successfully processed")
	return helper.ReturnCustomResponse(c, fiber.StatusCreated, true, "Payment successfully processed")
}

func paymentStatus(callback CallbackStruct, record model.Payment) error {

	switch callback.Status {
	case "PAID":
		r, err := routeros.Dial(config.MikrotikIpAddress+helper.ValidateAppPort(config.MikrotikPort), config.MikrotikUsername, config.MikrotikPassword)
		if err != nil {
			return fmt.Errorf("failed to connect to RouterOS: %v", err)
		}
		defer r.Close()

		cmd := []string{
			"/ip/hotspot/user/add",
			"server=all",
			"=profile=" + record.Profile.Name,
			"=name=" + record.Voucher,
			"=password=" + record.Voucher,
			"=comment=" + "vc-tripay",
		}

		_, err = r.RunArgs(cmd)
		if err != nil {
			if helper.IsErrorMessage(err, "from RouterOS device: failure: already have user with this name for this server") {
				return fmt.Errorf("HALLO BOSS")

				// DISINI BUAT GENERATE VOUCHER ULANG JIKA DI ROUTER OS TERDAPAT NAMA YANG SAMA
			}
			return fmt.Errorf("invalid RouterOS command: %v", err)
		}
		transaction := &model.Transaction{
			Status: callback.Status,
			PaidAt: helper.UnixToTime(callback.PaidAt),
			Note:   *callback.Note,
		}

		if err := database.DBConn.Model(&model.Transaction{}).Where("reference = ?", record.TransactionRef).Updates(transaction).Error; err != nil {
			return fmt.Errorf("error in database: %v", err)
		}

		return nil
	case "REFUND":
		if err := database.DBConn.Model(&model.Transaction{}).Where("reference = ?", record.TransactionRef).Update("status", "REFUND").Error; err != nil {
			return fmt.Errorf("error in database: %v", err)
		}

		return nil
	case "EXPIRED":
		if err := database.DBConn.Model(&model.Transaction{}).Where("reference = ?", record.TransactionRef).Update("status", "EXPIRED").Error; err != nil {
			return fmt.Errorf("error in database: %v", err)
		}

		return nil
	case "FAILED":
		if err := database.DBConn.Model(&model.Transaction{}).Where("reference = ?", record.TransactionRef).Update("status", "FAILED").Error; err != nil {
			return fmt.Errorf("error in database: %v", err)
		}

		return nil
	default:
		return fmt.Errorf("unrecognized payment status: %s", callback.Status)
	}
}
