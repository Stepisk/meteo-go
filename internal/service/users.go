package service

import (
	"context"
	"gitlab.com/peleng-meteo/meteo-go/internal/domain"
	"gitlab.com/peleng-meteo/meteo-go/internal/repository"
	"gitlab.com/peleng-meteo/meteo-go/pkg/auth"
	"gitlab.com/peleng-meteo/meteo-go/pkg/hash"
	"gitlab.com/peleng-meteo/meteo-go/pkg/logger"
	"gitlab.com/peleng-meteo/meteo-go/pkg/otp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type UsersService struct {
	repo           repository.Users
	hasher         hash.PasswordHasher
	tokenManager   auth.TokenManager
	otpGenenerator otp.Generator

	emailService Emails

	accessTokenTTL         time.Duration
	refreshTokenTTL        time.Duration
	verificationCodeLength int
}

func NewUsersService(repo repository.Users, hasher hash.PasswordHasher, tokenManager auth.TokenManager, emailService Emails, accessTTL, refreshTTL time.Duration, otpGenerator otp.Generator, verificationCodeLength int) *UsersService {
	return &UsersService{
		repo:                   repo,
		hasher:                 hasher,
		emailService:           emailService,
		tokenManager:           tokenManager,
		accessTokenTTL:         accessTTL,
		refreshTokenTTL:        refreshTTL,
		otpGenenerator:         otpGenerator,
		verificationCodeLength: verificationCodeLength,
	}
}

func (s *UsersService) SignUp(ctx context.Context, input SignUpInput) error {
	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return err
	}

	// TODO: it's possible to use OTP apps (Google Authenticator, Authy) compatibility mode here, in the future
	verificationCode := s.otpGenenerator.RandomSecret(s.verificationCodeLength)

	user := domain.User{
		Name:         input.Name,
		Password:     passwordHash,
		Email:        input.Email,
		RegisteredAt: time.Now(),
		LastVisitAt:  time.Now(),
		Verification: domain.Verification{
			Code: verificationCode,
		},
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if err == repository.ErrUserAlreadyExists {
			return ErrUserAlreadyExists
		}

		return err
	}

	go func() {
		if err := s.emailService.AddToList(user.Name, user.Email); err != nil {
			logger.Error("Failed to add email to the list:", err)
		}
	}()

	// TODO: If it fails what then?
	return s.emailService.SendVerificationEmail(SendVerificationEmailInput{
		Email:            user.Email,
		Name:             user.Name,
		VerificationCode: verificationCode,
	})
}

func (s *UsersService) SignIn(ctx context.Context, input SignInInput) (Tokens, error) {
	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return Tokens{}, err
	}
	user, err := s.repo.GetByCredentials(ctx, input.Email, passwordHash)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return Tokens{}, ErrUserNotFound
		}
		return Tokens{}, err
	}

	return s.createSession(ctx, user.ID)
}

func (s *UsersService) RefreshTokens(ctx context.Context, refreshToken string) (Tokens, error) {
	user, err := s.repo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return Tokens{}, err
	}

	return s.createSession(ctx, user.ID)
}

func (s *UsersService) Verify(ctx context.Context, hash string) error {
	err := s.repo.Verify(ctx, hash)
	if err != nil {
		if err == repository.ErrVerificationCodeInvalid {
			return ErrVerificationCodeInvalid
		}

		return err
	}

	return nil
}

func (s *UsersService) GetById(ctx context.Context, id primitive.ObjectID) (domain.User, error) {
	return s.repo.GetById(ctx, id)
}

func (s *UsersService) createSession(ctx context.Context, studentId primitive.ObjectID) (Tokens, error) {
	var (
		res Tokens
		err error
	)

	res.AccessToken, err = s.tokenManager.NewJWT(studentId.Hex(), s.accessTokenTTL)
	if err != nil {
		return res, err
	}

	res.RefreshToken, err = s.tokenManager.NewRefreshToken()
	if err != nil {
		return res, err
	}

	session := domain.Session{
		RefreshToken: res.RefreshToken,
		ExpiresAt:    time.Now().Add(s.refreshTokenTTL),
	}

	err = s.repo.SetSession(ctx, studentId, session)
	return res, err
}
