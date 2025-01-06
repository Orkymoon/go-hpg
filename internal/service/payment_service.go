package service

import (
	"fmt"
	"log"

	"github.com/go-routeros/routeros/v3"
	"github.com/orkymoon/tripay-golang/internal/config"
	"github.com/orkymoon/tripay-golang/internal/helper"
	"github.com/orkymoon/tripay-golang/internal/model"
	"github.com/orkymoon/tripay-golang/internal/repository"
	"github.com/zakirkun/go-tripay/client"
	"github.com/zakirkun/go-tripay/utils"
)

func BodyRequestService(r *model.PaymentRequest, p *model.Profile, tripay *client.Client) *client.ClosePaymentBodyRequest {

	merchantReff := helper.GenerateMerchantReference()

	tripay.SetSignature(utils.Signature{
		Amount:       int64(p.Amount),
		PrivateKey:   string(config.TripayPrivateKey),
		MerchantCode: string(config.TripayMerchantCode),
		MerchanReff:  merchantReff,
	})

	bodyReq := client.ClosePaymentBodyRequest{
		Method:        utils.TRIPAY_CHANNEL(r.PaymentMethod),
		MerchantRef:   merchantReff,
		Amount:        int(p.Amount),
		CustomerName:  r.CustomerName,
		CustomerEmail: r.CustomerEmail,
		CustomerPhone: r.CustomerPhone,
		ReturnURL:     "http://" + r.Hostname + "/login?dst=&username=" + r.Voucher + "&password=" + r.Voucher,
		ExpiredTime:   client.SetTripayExpiredTime(config.TripayExpiredTime), // 8 hours expiration
		Signature:     tripay.GetSignature(),
		OrderItems: []client.OrderItemClosePaymentRequest{
			{
				SKU:      p.Name,
				Name:     r.Voucher,
				Price:    int(p.Amount),
				Quantity: 1,
			},
		},
	}

	return &bodyReq
}

func PaymentStatus(callback model.Callback) error {

	switch callback.Status {
	case "PAID":
		r, err := routeros.Dial(config.MikrotikIpAddress+helper.ValidateAppPort(config.MikrotikPort), config.MikrotikUsername, config.MikrotikPassword)
		if err != nil {
			return fmt.Errorf("failed to connect to RouterOS: %v", err)
		}
		defer r.Close()

		payment, err := repository.GetPaymentByReferenceWithProfile(callback.Reference)
		if err != nil {
			log.Printf("Database Error: %v", err)
			return fmt.Errorf("database Error: %v", err)
		}

		cmd := []string{
			"/ip/hotspot/user/add",
			"server=all",
			"=profile=" + payment.Profile.Name,
			"=name=" + payment.Voucher,
			"=password=" + payment.Voucher,
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
			Note:   "",
		}
		if callback.Note != nil {
			transaction.Note = *callback.Note
		}
		if err := repository.UpdateTransactionByReference(callback.Reference, transaction); err != nil {
			return fmt.Errorf("error in database: %v", err)
		}

		return nil
	case "REFUND":
		transaction := &model.Transaction{
			Status: callback.Status,
			Note:   *callback.Note,
		}
		if err := repository.UpdateTransactionByReference(callback.Reference, transaction); err != nil {
			return fmt.Errorf("error in database: %v", err)
		}

		return nil
	case "EXPIRED":
		transaction := &model.Transaction{
			Status: callback.Status,
			Note:   *callback.Note,
		}
		if err := repository.UpdateTransactionByReference(callback.Reference, transaction); err != nil {
			return fmt.Errorf("error in database: %v", err)
		}

		return nil
	case "FAILED":
		transaction := &model.Transaction{
			Status: callback.Status,
			Note:   *callback.Note,
		}
		if err := repository.UpdateTransactionByReference(callback.Reference, transaction); err != nil {
			return fmt.Errorf("error in database: %v", err)
		}

		return nil
	default:
		return fmt.Errorf("unrecognized payment status: %s", callback.Status)
	}
}
