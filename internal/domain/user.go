package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name         string             `json:"name" bson:"name"`
	Email        string             `json:"email" bson:"email"`
	Password     string             `json:"password" bson:"password"`
	RegisteredAt time.Time          `json:"registeredAt" bson:"registeredAt"`
	LastVisitAt  time.Time          `json:"lastVisitAt" bson:"lastVisitAt"`
	Verification Verification       `json:"verification" bson:"verification"`
	Session      Session            `json:"session" bson:"session,omitempty"`
}

type Verification struct {
	Code     string `json:"code" bson:"code"`
	Verified bool   `json:"verified" bson:"verified"`
}
