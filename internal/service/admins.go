package service

import (
	"context"
	"gitlab.com/peleng-meteo/meteo-go/internal/domain"
	"gitlab.com/peleng-meteo/meteo-go/internal/repository"
	"gitlab.com/peleng-meteo/meteo-go/pkg/auth"
	"gitlab.com/peleng-meteo/meteo-go/pkg/hash"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AdminsService struct {
	hasher       hash.PasswordHasher
	tokenManager auth.TokenManager

	repo repository.Admins

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAdminsService(hasher hash.PasswordHasher, tokenManager auth.TokenManager, repo repository.Admins, accessTokenTTL time.Duration, refreshTokenTTL time.Duration) *AdminsService {
	return &AdminsService{
		hasher:          hasher,
		tokenManager:    tokenManager,
		repo:            repo,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *AdminsService) SignIn(ctx context.Context, input SignInInput) (Tokens, error) {
	// TODO: implement password hashing
	user, err := s.repo.GetByCredentials(ctx, input.Email, input.Password)
	if err != nil {
		return Tokens{}, err
	}

	return s.createSession(ctx, user.ID)
}

func (s *AdminsService) RefreshTokens(ctx context.Context, refreshToken string) (Tokens, error) {
	student, err := s.repo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return Tokens{}, err
	}

	return s.createSession(ctx, student.ID)
}

func (s *AdminsService) createSession(ctx context.Context, adminId primitive.ObjectID) (Tokens, error) {
	var (
		res Tokens
		err error
	)

	res.AccessToken, err = s.tokenManager.NewJWT(adminId.Hex(), s.accessTokenTTL)
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

	err = s.repo.SetSession(ctx, adminId, session)
	return res, err
}
