package repository

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // подключаемся к БД
	if err != nil {
		return nil, err
	}

	// Возвращаем объект Repository с подключенной базой данных
	return &Repository{
		db: db,
	}, nil
}

func (r *Repository) GetUser() int {
	return 1
}
