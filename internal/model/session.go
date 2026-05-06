package model

import (
	"time"
)

type Session struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	ClientId              string    `json:"client_id"`
	ClientSecret          string    `json:"client_secret"`
	ClientSecretExpiresAt time.Time `json:"client_secret_expires_at"`
	DeviceCode            string    `json:"device_code"`
	DeviceExpiresAt       time.Time `json:"device_expires_at"`
	VerificationUrl       string    `json:"verification_url"`
}
