package model

import "time"

type Credential struct {
	AccessKeyId     string    `json:"access_key_id"`
	ExpiresAt       time.Time `json:"expires_at"`
	SecretAccessKey string    `json:"secret_access_key"`
	SessionToken    string    `json:"session_token"`
}
