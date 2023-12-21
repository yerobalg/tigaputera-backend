package model

import (
	"gorm.io/gorm"
)

type ProjectType string
type ProjectStatus string
type Color string

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

const (
	White     Color = "FFFFFF"
	LightGrey Color = "DEE2E6"
	Blue      Color = "3A57E8"
	Black     Color = "001129"
	Orange    Color = "F16A1B"
	Green     Color = "1AA053"
	DarkGrey  Color = "6C757D"
	Turquoise Color = "079AA2"
	Red       Color = "C03221"
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
	Budget      *int64  `gorm:"default:0" json:"budget"`
	StartDate   int64   `gorm:"default:0" json:"startDate"`
	FinalDate   int64   `gorm:"default:0" json:"finalDate"`
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
	StartDate   int64  `json:"starDate" validate:"required"`
	FinalDate   int64  `json:"finalDate" validate:"required"`
	Volume      *int64 `json:"volume"`
	Length      *int64 `json:"length"`
	Width       *int64 `json:"width"`
}

type ProjectListResponse struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Type          LabelStyle `json:"type"`
	Status        LabelStyle `json:"status"`
	UpdatedAt     int64      `json:"updatedAt"`
	InspectorName string     `json:"inspectorName"`
}

type ProjectDetailResponse struct {
	ID                 int64                      `json:"id"`
	Name               string                     `json:"name"`
	Description        string                     `json:"description"`
	Type               LabelStyle                 `json:"type"`
	Status             LabelStyle                 `json:"status"`
	DeptName           string                     `json:"deptName"`
	CompanyName        string                     `json:"companyName"`
	Volume             *int64                     `json:"volume"`
	Length             *int64                     `json:"length"`
	Width              *int64                     `json:"width"`
	StartDate          int64                      `json:"startDate"`
	FinalDate          int64                      `json:"finalDate"`
	InspectorName      string                     `json:"inspectorName"`
	ProjectBudget      ProjectBudget              `json:"projectBudget"`
	ProjectExpenditure ProjectExpenditureResponse `json:"projectExpenditure"`
	Margin             string                     `json:"margin"`
}

type ProjectBudget struct {
	Budgets       []Budget `json:"budgets"`
	PPNPercentage float64  `json:"ppnPercentage"`
	PPHPercentage float64  `json:"pphPercentage"`
	Total         string   `json:"total"`
}

type Budget struct {
	Name  string `json:"name"`
	Price string `json:"price"`
}

type LabelStyle struct {
	Name         string `json:"name"`
	BGColorHex   string `json:"bgColorHex"`
	TextColorHex string `json:"textColorHex"`
}

type UpdateProjectBudgetBody struct {
	Budget int64   `json:"budget" validate:"required"`
	PPN    float64 `json:"ppn" validate:"required,min=0,max=1"`
	PPH    float64 `json:"pph" validate:"required,min=0,max=1"`
}

type UpdateProjectStatusBody struct {
	Status string `json:"status" validate:"required"`
}

func IsProjectTypeCorrect(typeName string) bool {
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

func IsProjectStatusCorrect(statusName string) bool {
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

func GetProjectTypeStyle(projectType string) LabelStyle {
	labelStyle := LabelStyle{
		Name:         projectType,
		TextColorHex: string(White),
	}
	switch projectType {
	case string(Drainage):
		labelStyle.BGColorHex = string(Blue)
	case string(Concrete):
		labelStyle.BGColorHex = string(Black)
	case string(Ashpalt):
		labelStyle.BGColorHex = string(LightGrey)
		labelStyle.TextColorHex = string(Black)
	case string(Building):
		labelStyle.BGColorHex = string(Orange)
	}

	return labelStyle
}

func GetProjectStatusStyle(projectStatus string) LabelStyle {
	labelStyle := LabelStyle{
		Name:         projectStatus,
		TextColorHex: string(White),
	}
	switch projectStatus {
	case string(Running):
		labelStyle.BGColorHex = string(Turquoise)
	case string(Finished):
		labelStyle.BGColorHex = string(Green)
	case string(Postponed):
		labelStyle.BGColorHex = string(DarkGrey)
	case string(Canceled):
		labelStyle.BGColorHex = string(Red)
	}

	return labelStyle
}
