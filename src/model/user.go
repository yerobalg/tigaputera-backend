package model

import (
	"gorm.io/gorm"
)

type Role string

const (
	Admin     Role = "Admin"
	Inspector Role = "Inspector"
)

type User struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `json:"createdBy"`
	UpdatedBy *int64         `json:"updatedBy"`
	DeletedBy *int64         `json:"deletedBy"`

	Username     string `gorm:"not null;unique;type:varchar(255)" json:"username"`
	Name         string `gorm:"not null;type:varchar(255)" json:"name"`
	Password     string `gorm:"not null;type:text" json:"-"`
	IsFirstLogin bool   `gorm:"default:true" json:"isFirstLogin"`
	Role         Role   `gorm:"type:varchar(255);default:Inspector;index" json:"role"`
}

type UserParam struct {
	Username string `param:"username"`
	Role     string `param:"role"`
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

type CreateInspectorBody struct {
	Username string `json:"username" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}
