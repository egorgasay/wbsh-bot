package storage

import (
	"bot/internal/constant"
	"bot/internal/entity/table"
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Storage for items.
type Storage struct {
	db *gorm.DB
}

type Config struct {
	DSN string `json:"dsn"`
}

// New returns new storage
func New(config Config) (*Storage, error) {
	db, err := gorm.Open(sqlite.Open(config.DSN), &gorm.Config{
		// TODO: Logger:         logger.GetGormLogger(),
		TranslateError: true,
	})

	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	err = db.AutoMigrate(&table.User{})
	if err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) GetUserByID(ID int) (u table.User, err error) {
	if err := s.db.First(&u, "id = ?", ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return u, constant.ErrUserNotFound
		}
		return u, err
	}

	return u, nil
}

func (s *Storage) AddUser(us table.User) error {
	if err := s.db.Create(&us).Error; err != nil {
		return err
	}

	return nil
}
func (s *Storage) SaveUser(us table.User) error {
	if err := s.db.Save(&us).Error; err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetSubscribers() (subs []table.User, err error) {
	if err := s.db.Find(&subs, "subscribed = ?", true).Error; err != nil {
		return subs, err
	}

	if len(subs) == 0 {
		return subs, constant.ErrNoSubscribers
	}

	return subs, nil
}
