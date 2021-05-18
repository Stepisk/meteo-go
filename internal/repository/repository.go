package repository

import (
	"context"
	"gitlab.com/peleng-meteo/meteo-go/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Users interface {
	Create(ctx context.Context, user domain.User) error
	GetByCredentials(ctx context.Context, email, password string) (domain.User, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (domain.User, error)
	SetSession(ctx context.Context, id primitive.ObjectID, session domain.Session) error
	GetById(ctx context.Context, id primitive.ObjectID) (domain.User, error)
	Verify(ctx context.Context, code string) error
}

type Admins interface {
	GetByCredentials(ctx context.Context, email, password string) (domain.Admin, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (domain.Admin, error)
	SetSession(ctx context.Context, id primitive.ObjectID, session domain.Session) error
	GetById(ctx context.Context, id primitive.ObjectID) (domain.Admin, error)
}

type Sensors interface {
	Read() ([]byte, error)
	Write(data []byte) error
	Parse(data []byte) error
}

type Repositories struct {
	Users  Users
	Admins Admins
}

func NewRepositories(db *mongo.Database) *Repositories {
	return &Repositories{
		Users:  NewUsersRepo(db),
		Admins: NewAdminsRepo(db),
	}
}
