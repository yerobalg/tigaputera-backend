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

	Name        string  `gorm:"not null;unique;type:varchar(255)" json:"name"`
	Description string  `gorm:"not null;type:varchar(255)" json:"description"`
	Type        string  `gorm:"not null;type:varchar(255)" json:"type"`
	DeptName    string  `gorm:"not null;type:varchar(255)" json:"deptName"`
	CompanyName string  `gorm:"not null;type:varchar(255)" json:"companyName"`
	Status      string  `gorm:"not null;type:varchar(255)" json:"status"`
	Budget      int64   `gorm:"default:0" json:"budget"`
	PPN         float64 `gorm:"default:0.11" json:"ppn"`
	PPH         float64 `gorm:"default:0.015" json:"pph"`
	Volume      *int64  `json:"volume"`
	Length      *int64  `json:"length"`
	Width       *int64  `json:"width"`
	InspectorID int64   `json:"inspectorId"`
	Inspector   User    `gorm:"foreignKey:InspectorID" json:"inspector"`
}

type ProjectParam struct {
	ID int64 `uri:"project_id" param:"id"`
	PaginationParam
}

type CreateProjectBody struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	Type        string `json:"type" validate:"required"`
	DeptName    string `json:"deptName" validate:"required"`
	CompanyName string `json:"companyName" validate:"required"`
	InspectorID int64  `json:"inspectorId" validate:"required"`
	Volume      *int64 `json:"volume"`
	Length      *int64 `json:"length"`
	Width       *int64 `json:"width"`
}

type ProjectListResponse struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	UpdatedAt     int64  `json:"updatedAt"`
	InspectorName string `json:"inspectorName"`
}

type UpdateProjectBudgetBody struct {
	Budget int64   `json:"budget" validate:"required"`
	PPN    float64 `json:"ppn" validate:"required,min=0,max=1"`
	PPH    float64 `json:"pph" validate:"required,min=0,max=1"`
}

func ValidateProjectType(typeName string) bool {
	typeNames := []string{
		string(Drainage),
		string(Concrete),
		string(Ashpalt),
		string(Building),
	}

	for _, t := range typeNames {
		if t == typeName {
			return true
		}
	}

	return false
}

func ValidateProjectStatus(statusName string) bool {
	statusNames := []string{
		string(Running),
		string(Finished),
		string(Postponed),
		string(Canceled),
	}

	for _, s := range statusNames {
		if s == statusName {
			return true
		}
	}

	return false
}
