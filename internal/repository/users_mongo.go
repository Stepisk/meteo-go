package repository

import (
	"context"
	"gitlab.com/peleng-meteo/meteo-go/internal/domain"
	"gitlab.com/peleng-meteo/meteo-go/pkg/database/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type UsersRepo struct {
	db *mongo.Collection
}

func NewUsersRepo(db *mongo.Database) *UsersRepo {
	return &UsersRepo{
		db: db.Collection(usersCollection),
	}
}

func (r *UsersRepo) Create(ctx context.Context, user domain.User) error {
	_, err := r.db.InsertOne(ctx, user)
	if mongodb.IsDuplicate(err) {
		return ErrUserAlreadyExists
	}

	return err
}

func (r *UsersRepo) GetByCredentials(ctx context.Context, email, password string) (domain.User, error) {
	var user domain.User
	if err := r.db.FindOne(ctx, bson.M{"email": email, "password": password, "verification.verified": true}).
		Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return domain.User{}, ErrUserNotFound
		}

		return domain.User{}, err
	}

	return user, nil
}

func (r *UsersRepo) GetByRefreshToken(ctx context.Context, refreshToken string) (domain.User, error) {
	var user domain.User
	if err := r.db.FindOne(ctx, bson.M{"session.refreshToken": refreshToken, "session.expiresAt": bson.M{"$gt": time.Now()}}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return domain.User{}, ErrUserNotFound
		}

		return domain.User{}, err
	}

	return user, nil
}

func (r *UsersRepo) GetById(ctx context.Context, id primitive.ObjectID) (domain.User, error) {
	var user domain.User
	err := r.db.FindOne(ctx, bson.M{"_id": id, "verification.verified": true}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return domain.User{}, ErrUserNotFound
		}

		return domain.User{}, err
	}

	return user, nil
}

func (r *UsersRepo) SetSession(ctx context.Context, id primitive.ObjectID, session domain.Session) error {
	_, err := r.db.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"session": session, "lastVisitAt": time.Now()}})
	return err
}

func (r *UsersRepo) Verify(ctx context.Context, code string) error {
	res, err := r.db.UpdateOne(ctx,
		bson.M{"verification.code": code},
		bson.M{"$set": bson.M{"verification.verified": true, "verification.code": ""}})
	if err != nil {
		return err
	}

	if res.ModifiedCount == 0 {
		return ErrVerificationCodeInvalid
	}

	return nil
}
