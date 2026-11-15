package config

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/resend/resend-go/v4"
)

func NewResendClient() (*resend.Client, error) {
	key := os.Getenv("RESEND_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("environment variable RESEND_API_KEY is required")
	}
	return resend.NewCustomClient(&http.Client{Timeout: 5 * time.Second}, key), nil
}

func MailSupportAddress() (string, error) {
	address := os.Getenv("MAIL_SUPPORT_ADDRESS")
	if address == "" {
		return "", fmt.Errorf("environment variable MAIL_SUPPORT_ADDRESS is required")
	}
	return address, nil
}

func MailAdminAddress() (string, error) {
	address := os.Getenv("MAIL_ADMIN_ADDRESS")
	if address == "" {
		return "", fmt.Errorf("environment variable MAIL_ADMIN_ADDRESS is required")
	}
	return address, nil
}
