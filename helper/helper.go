package helper

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func ValidateAppPort(port string) string {
	vPort := ":" + port
	return vPort
}

func GenerateMerchantReference() string {
	return fmt.Sprintf("%d-%s", time.Now().Unix(), uuid.New())
}

func CompareSignature(signature1, signature2 string) bool {
	sign1, _ := hex.DecodeString(signature1)
	sign2, _ := hex.DecodeString(signature2)
	return hmac.Equal(sign1, sign2)
}

func CallbackSignature(callbackData []byte, Pk string) (signature string) {
	mac := hmac.New(sha256.New, []byte(Pk))
	mac.Write(callbackData)
	s := hex.EncodeToString(mac.Sum(nil))
	return s
}

func ReturnCustomResponse(c *fiber.Ctx, statusCode int, success bool, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": success,
		"message": message,
	})
}
