package repository

import (
	"errors"

	"github.com/minio/minio-go/v7"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	minioClient "web_backend/internal/app/minioClient"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotAllowed    = errors.New("not allowed")
	ErrNoDraft       = errors.New("no draft for this user")
)

type Repository struct {
	db     *gorm.DB
	mc     *minio.Client
	userID int
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	mc, err := minioClient.InitMinio()
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:     db,
		mc:     mc,
		userID: creatorUserID,
	}, nil
}

const creatorUserID = 1

func (r *Repository) GetCreatorID() int {
	return creatorUserID
}

func (r *Repository) GetUserID() int {
	return r.userID
}

func (r *Repository) SetUserID(id int) {
	r.userID = id
}

func (r *Repository) SignOut() {
	r.userID = 0
}
