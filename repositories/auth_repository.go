package repositories

import (
	"github.com/mllb/sampletodo/models"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return AuthRepository{
		db: db,
	}
}

func (r *AuthRepository) SignUp(email string, password string) error {
	user := models.User{
		Email:    email,
		Password: password,
	}
	if err := r.db.Create(&user).Error; err != nil {
		return err
	}
	return nil
}

func (r *AuthRepository) FindUser(email string) (models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}
