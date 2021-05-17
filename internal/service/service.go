package service

import (
	"context"
	"gitlab.com/peleng-meteo/meteo-go/internal/config"
	"gitlab.com/peleng-meteo/meteo-go/internal/domain"
	"gitlab.com/peleng-meteo/meteo-go/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

// TODO handle "not found" errors

type SignUpInput struct {
	Name     string
	Email    string
	Password string
}

type SignInInput struct {
	Email    string
	Password string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type Users interface {
	SignUp(ctx context.Context, input SignUpInput) error
	SingIn(ctx context.Context, input SignInInput) (Tokens, error)
	RefreshTokens(ctx context.Context, refreshToken string) (Tokens, error)
	Verify(ctx context.Context, hash string) error
	GetById(ctx context.Context, id primitive.ObjectID) (domain.User, error)
}

type Admins interface {
	SignIn(ctx context.Context, input SignInInput) (Tokens, error)
	RefreshTokens(ctx context.Context, refreshToken string) (Tokens, error)
}

type SendVerificationEmailInput struct {
	Email            string
	Name             string
	VerificationCode string
}

type Emails interface {
	AddToList(name, email string) error
	SendVerificationEmail(input SendVerificationEmailInput) error
}

type Services struct {
	Users  Users
	Admins Admins
}

type Deps struct {
	Repos                  *repository.Repositories
	Cache                  cache.Cache
	Hasher                 hash.PasswordHasher
	TokenManager           auth.TokenManager
	EmailProvider          email.Provider
	EmailSender            email.Sender
	EmailConfig            config.EmailConfig
	AccessTokenTTL         time.Duration
	RefreshTokenTTL        time.Duration
	CacheTTL               int64
	OtpGenerator           otp.Generator
	VerificationCodeLength int
	FrontendURL            string
	Environment            string
}

func NewServices(deps Deps) *Services {
	emailsService := NewEmailsService(deps.EmailProvider, deps.EmailSender, deps.EmailConfig, deps.FrontendURL)
	usersService := NewUsersService(deps.Repos.Users, deps.Hasher, deps.TokenManager, emailsService, deps.AccessTokenTTL, deps.RefreshTokenTTL, deps.OtpGenerator, deps.VerificationCodeLength)

	return &Services {
		Users: usersService,

	}
}
