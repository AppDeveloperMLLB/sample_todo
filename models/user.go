package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Password  string    `gorm:"not null;size:255" json:"password"`
	CreatedAt time.Time `gorm:"- autoCreateTime" json:"-"`
	UpdatedAt time.Time `gorm:"- autoUpdateTime" json:"-"`
}
