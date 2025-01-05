package config

import (
	"os"
	"strconv"

	"github.com/orkymoon/tripay-golang/helper"
)

var (
	AppPort            string
	MysqlHost          string
	MysqlPort          string
	MysqlDbname        string
	MysqlUser          string
	MysqlPassword      string
	TripayPrivateKey   string
	TripayApiKey       string
	TripayMerchantCode string
	TripayMode         string
	TripayExpiredTime  int
	MikrotikUsername   string
	MikrotikPassword   string
	MikrotikIpAddress  string
	MikrotikPort       string
)

type Config interface {
	Get(key string) string
}

type config struct{}

func (c config) Get(key string) string {
	return os.Getenv(key)
}

func (c config) ifEmpty(env string, defaultValue string) string {
	if env != "" {
		return env
	}
	return defaultValue
}

func (c config) toInt(env string, defaultValue int) int {
	if env != "" {
		value, err := strconv.Atoi(env)
		if err == nil {
			return value
		}
	}
	return defaultValue
}

var c = config{}

func LoadEnv() {
	// running port for aplication
	AppPort = c.ifEmpty(helper.ValidateAppPort(c.Get("APP_PORT")), ":3000")

	// mysql configuration .env
	MysqlHost = c.ifEmpty(c.Get("MYSQL_HOST"), "localhost")
	MysqlPort = c.ifEmpty(c.Get("MYSQL_PORT"), "3306")
	MysqlDbname = c.ifEmpty(c.Get("MYSQL_DBNAME"), "go_pmb")
	MysqlUser = c.ifEmpty(c.Get("MYSQL_USER"), "root")
	MysqlPassword = c.Get("MYSQL_PASSWORD")

	// tripay
	TripayPrivateKey = c.Get("TRIPAY_PRIVATE_KEY")
	TripayApiKey = c.Get("TRIPAY_API_KEY")
	TripayMerchantCode = c.Get("TRIPAY_MERCHANT_CODE")
	TripayMode = c.Get("TRIPAY_MODE")
	TripayExpiredTime = c.toInt(c.Get("TRIPAY_EXPIRED_TIME"), 24)

	// MIKROTIK
	MikrotikUsername = c.ifEmpty(c.Get("MIKROTIK_USER"), "admin")
	MikrotikPassword = c.Get("MIKROTIK_PASSWORD")
	MikrotikIpAddress = c.ifEmpty(c.Get("MIKROTIK_IP"), "192.168.88.1")
	MikrotikPort = c.ifEmpty(c.Get("MIKROTIK_PORT"), "8728")
}
