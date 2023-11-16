package model

import (
	"gorm.io/gorm"
)

type Role string

const (
	Admin      Role = "Admin"
	Supervisor Role = "Supervisor"
)

type User struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt" swaggertype:"string" example:"2020-12-31T00:00:00Z"`
	CreatedBy *int64         `json:"createdBy"`
	UpdatedBy *int64         `json:"updatedBy"`
	DeletedBy *int64         `json:"deletedBy"`

	Username     string `gorm:"not null;unique;type:varchar(255)" json:"username"`
	Name         string `gorm:"not null;type:varchar(255)" json:"name"`
	Password     string `gorm:"not null;type:text" json:"-"`
	IsFirstLogin bool   `gorm:"default:true" json:"isFirstLogin"`
	Role         Role   `gorm:"type:varchar(255);default:Supervisor;index" json:"role"`
}

type UserParam struct {
	Username string `param:"username"`
	PaginationParam
}

type UserLoginBody struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserLoginResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

type ResetPasswordBody struct {
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}
