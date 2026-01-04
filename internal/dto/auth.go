package dto

import (
	"time"
)

type AccessToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type RefreshToken struct {
	Token     string
	ExpiresAt time.Time
}
