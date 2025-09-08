package model

import (
	"gorm.io/gorm"
)

// UserToken is the GORM model for user tokens
type UserToken struct {
	User       string `gorm:"unique"`
	Token      string
	Comment    string
	Ports      string `gorm:"type:text"` // Stored as JSON string
	Domains    string `gorm:"type:text"` // Stored as JSON string
	Subdomains string `gorm:"type:text"` // Stored as JSON string
	Enable     bool
	Server     string
	CreateDate string
	ExpireDate string
	gorm.Model
}
