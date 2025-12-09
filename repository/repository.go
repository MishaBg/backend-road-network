package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	minioClient "lab4-swag/internal/app/minioClient"

	"github.com/go-redis/redis"
	"github.com/minio/minio-go/v7"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotAllowed    = errors.New("not allowed")
	ErrNoDraft       = errors.New("no draft for this user")
)

type Repository struct {
	db *gorm.DB
	mc *minio.Client
	rd *redis.Client
}

func NewRepository(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // подключаемся к БД
	if err != nil {
		return nil, err
	}

	mc, err := minioClient.InitMinio()
	if err != nil {
		return nil, err
	}

	rd := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: "password",
		DB:       0,
	})

	return &Repository{
		db: db,
		mc: mc,
		rd: rd,
	}, nil
}

func (r *Repository) GetToken(userID string) (string, error) {
	token, err := r.rd.Get(userID).Result()
	if err != nil {
		return "", err
	}
	return token, nil
}

func blacklistKeyForToken(tokenString string) string {
	h := sha256.Sum256([]byte(tokenString))
	return "blacklist:" + hex.EncodeToString(h[:])
}

func (r *Repository) AddTokenToBlacklist(ctx context.Context, tokenString string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	key := blacklistKeyForToken(tokenString)
	return r.rd.Set(key, "1", ttl).Err()
}

func (r *Repository) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	key := blacklistKeyForToken(tokenString)
	n, err := r.rd.Exists(key).Result()
	fmt.Println(err)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
