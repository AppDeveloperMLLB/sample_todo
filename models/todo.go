package models

import (
	"time"
)

type Todo struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Title string `gorm:"not null;size:50" json:"title"`
	Body  string `gorm:"not null;size:300" json:"body"`
	// EmployeeNumber string    `gorm:"uniqueIndex;not null;size:8" json:"employeeNumber"`
	// Email          string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	// Password       string    `gorm:"not null;size:255" json:"-"`
	// Verified       bool      `gorm:"not null" json:"verified"`
	// Role           uint8     `gorm:"not null" json:"role"`
	// Disabled       bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"- autoCreateTime" json:"-"`
	UpdatedAt time.Time `gorm:"- autoUpdateTime" json:"-"`
}
