package model

import (
	"gorm.io/gorm"
)

type role string

const (
	Admin      role = "Admin"
	Supervisor role = "Supervisor"
)

type User struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	CreatedBy *int64         `json:"createdBy"`
	UpdatedBy *int64         `json:"updatedBy"`
	DeletedBy *int64         `json:"deletedBy"`

	Username     string `gorm:"not null;unique;type:varchar(255)" json:"username"`
	Name         string `gorm:"not null;type:varchar(255)" json:"name"`
	Password     string `gorm:"not null;type:text" json:"-"`
	IsFirstLogin bool   `gorm:"default:true" json:"isFirstLogin"`
	Role         role   `gorm:"type:varchar(255);default:Supervisor;index" json:"role"`
}

type UserLoginBody struct {
	
}
