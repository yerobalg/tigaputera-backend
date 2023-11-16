package model

import (
	"gorm.io/gorm"
)

type ProjectType string
type ProjectStatus string

const (
	Drainage ProjectType = "Drainase"
	Concrete ProjectType = "Beton"
	Ashpalt  ProjectType = "Hotmix"
	Building ProjectType = "Bangunan"
)

const (
	Running   ProjectStatus = "Sedang Berjalan"
	Finished  ProjectStatus = "Selesai"
	Postponed ProjectStatus = "Ditunda"
	Canceled  ProjectStatus = "Dibatalkan"
)

type Project struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `json:"createdBy"`
	UpdatedBy *int64         `json:"updatedBy"`
	DeletedBy *int64         `json:"deletedBy"`

	Name        string `gorm:"not null;unique;type:varchar(255)" json:"name"`
	Description string `gorm:"not null;type:vachar(255)" json:"description"`
	Type        string `gorm:"not null;type:varchar(255)" json:"type"`
	DeptName    string `gorm:"not null;type:varchar(255)" json:"deptName"`
	CompanyName string `gorm:"not null;type:varchar(255)" json:"companyName"`
	Status      string `gorm:"not null;type:varchar(255)" json:"status"`
	Volume      int64  `json:"volume"`
	Length      int64  `json:"length"`
	Width       int64  `json:"width"`
	InspectorID int64  `json:"inspectorId"`
	Inspector   User   `gorm:"foreignKey:InspectorID" json:"inspector"`
}
