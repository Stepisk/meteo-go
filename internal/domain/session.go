package domain

import "time"

type Session struct {
	RefreshToken string `json:"refresh_token" bson:"refresh_token"`
	ExpiresAt time.Time `json:"expires_at" bson:"expires_at"`
}
