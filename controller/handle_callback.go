package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"

	"github.com/go-routeros/routeros/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/orkymoon/tripay-golang/config"
	"github.com/orkymoon/tripay-golang/database"
	"github.com/orkymoon/tripay-golang/helper"
	"github.com/orkymoon/tripay-golang/model"
	"github.com/zakirkun/go-tripay/client"
	"github.com/zakirkun/go-tripay/utils"
	"gorm.io/gorm"
)

type Payment struct {
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
	PaidAt            int64   `json:"paid_at"` // menggunakan Unix timestamp
	Note              *string `json:"note"`    // menggunakan pointer untuk nilai null
}

func callbackSignature(callbackData []byte, Pk string) (signature string) {
	mac := hmac.New(sha256.New, []byte(Pk))
	mac.Write(callbackData)
	s := hex.EncodeToString(mac.Sum(nil))
	return s
}

func HandleCallback(c *fiber.Ctx) error {
	context := fiber.Map{
		"success": true,
		"message": "Successfully received callback ",
	}
	c.Status(200)

	routerosIp := config.MikrotikIpAddress + helper.ValidateAppPort(config.MikrotikPort)
	r, err := routeros.Dial(string(routerosIp), string(config.MikrotikUsername), string(config.MikrotikPassword))
	if err != nil {
		context["success"] = false
		context["message"] = "failed to connect to RouterOS: " + err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	defer r.Close()

	callbackData := c.Body()

	signature := callbackSignature(callbackData, config.TripayPrivateKey)
	callbackSignature := c.Get("X-Callback-Signature")

	result := helper.CompareSignature(signature, callbackSignature)
	if !result {
		context["success"] = false
		context["message"] = "Invalid signaturer" + signature
		return c.Status(fiber.StatusUnauthorized).JSON(context)
	}

	var callback Payment

	err = json.Unmarshal(callbackData, &callback)
	if err != nil {
		// Jika gagal mendekode, kembalikan status 400 Bad Request dengan pesan kesalahan
		context["success"] = false
		context["message"] = "Failed to decode callback data err: " + err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	callbackEvent := c.Get("X-Callback-Event")

	if callbackEvent != "payment_status" {
		context["success"] = false
		context["message"] = "Unrecognized callback event: " + callbackEvent
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var record model.Payment

	if callback.IsClosedPayment == 1 {
		result := database.DBConn.Preload("Transaction", "status = ? AND merchant_ref = ?", "UNPAID", callback.MerchantRef).Where("transaction_ref = ?", callback.Reference).Limit(1).First(&record)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				context["success"] = false
				context["message"] = "Invoice not found or already paid"
				return c.Status(fiber.StatusBadRequest).JSON(context)
			}
			return c.Status(fiber.StatusBadRequest).JSON(result.Error)
		}

		tripay := client.Client{
			ApiKey:       config.TripayApiKey,
			PrivateKey:   config.TripayPrivateKey,
			MerchantCode: config.TripayMerchantCode,
			Mode:         utils.TRIPAY_MODE(config.TripayMode),
		}

		tripay.SetSignature(utils.Signature{
			Amount:       callback.TotalAmount,
			PrivateKey:   config.TripayPrivateKey,
			MerchantCode: config.TripayMerchantCode,
			MerchanReff:  callback.MerchantRef,
		})

		switch callback.Status {
		// handle status paid
		case "PAID":
			cmd := []string{
				"/ip/hotspot/user/add",
				"server=all",
				"=profile=" + record.Profile,
				"=name=" + record.Voucher,
				"=password=" + record.Voucher,
				"=comment=" + "vc-tripay",
			}
			_, err := r.RunArgs(cmd)
			if err != nil {
				context["success"] = false
				context["message"] = "Invalid RouterOS command :" + err.Error()
				return c.Status(fiber.StatusBadRequest).JSON(context)
			}
			result := database.DBConn.Model(&model.Transaction{}).Where("merchant_ref = ?", record.Transaction.MerchantRef).Update("status", "PAID")
			if result.Error != nil {
				log.Println(result.Error)
				context["success"] = false
				context["message"] = "Error in database "
				return c.Status(fiber.ErrBadRequest.Code).JSON(context)
			}

		case "EXPIRED":
			result := database.DBConn.Model(record).Where("merchant_ref = ?", record.Transaction.MerchantRef).Update("status", "EXPIRED")
			if result.Error != nil {
				log.Println(result.Error)
				context["success"] = false
				context["message"] = "Error in database "
				return c.Status(fiber.ErrBadRequest.Code).JSON(context)
			}
		case "FAILED":
			result := database.DBConn.Model(record).Where("merchant_ref = ?", record.Transaction.MerchantRef).Update("status", "FAILED")
			if result.Error != nil {
				log.Println(result.Error)
				context["success"] = false
				context["message"] = "Error in database "
				return c.Status(fiber.ErrBadRequest.Code).JSON(context)
			}
		default:
			context["success"] = false
			context["message"] = "Unrecognized payment status "
			return c.Status(fiber.ErrBadRequest.Code).JSON(context)
		}

	}
	return c.JSON(context)
}
