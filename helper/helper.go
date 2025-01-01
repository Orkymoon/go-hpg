package helper

import (
	"crypto/hmac"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func ValidateMacAddress(mac string) bool {
	re := regexp.MustCompile("^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$")
	return re.MatchString(mac)
}
func ValidatePhoneNumber(phone uint) string {
	// Mengonversi uint ke string untuk validasi
	phoneStr := strconv.Itoa(int(phone))

	// Validasi panjang nomor telepon (harus 11 digit atau lebih)
	if len(phoneStr) < 11 {
		return ""
	}

	// Validasi awalan nomor telepon
	if !strings.HasPrefix(phoneStr, "8") && !strings.HasPrefix(phoneStr, "628") {
		return ""
	}

	// Jika nomor dimulai dengan 628, ganti dengan 08
	if strings.HasPrefix(phoneStr, "628") {
		phoneStr = "08" + phoneStr[3:]
	} else {
		phoneStr = "0" + phoneStr[:]
	}

	return phoneStr
}
func ValidateAppPort(port string) string {
	vPort := ":" + port
	return vPort
}
func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
func GenerateMerchantReference() string {
	return fmt.Sprintf("MREF-%s", uuid.New().String())
}
func CompareSignature(signature1, signature2 string) bool {
	sign1, _ := hex.DecodeString(signature1)
	sign2, _ := hex.DecodeString(signature2)
	return hmac.Equal(sign1, sign2)
}
