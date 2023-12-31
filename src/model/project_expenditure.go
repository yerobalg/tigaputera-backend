package model

import "gorm.io/gorm"

type ProjectExpenditure struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `json:"createdBy"`
	UpdatedBy *int64         `json:"updatedBy"`
	DeletedBy *int64         `json:"deletedBy"`

	ProjectID   int64   `json:"projectId" gorm:"index:idx_project_expenditure,unique"`
	Sequence    int64   `gorm:"not null;index:idx_project_expenditure,unique" json:"sequence"`
	Name        string  `gorm:"not null;type:varchar(255)" json:"name"`
	TotalPrice  *int64  `gorm:"default:0" json:"totalPrice"`
	IsFixedCost *bool   `gorm:"default:true" json:"isFixedCost"`
	Project     Project `gorm:"foreignKey:ProjectID" json:"project"`
}

type ProjectExpenditureParam struct {
	ProjectID int64 `uri:"project_id" param:"project_id"`
	ID        int64 `uri:"id" param:"id"`
	PaginationParam
}

type ProjectExpenditureListResponse struct {
	Expenditures []ProjectExpenditureList `json:"expenditures"`
	SumTotal     string                   `json:"sumTotal"`
}

type ProjectExpenditureResponse struct {
	Expenditures []ProjectExpenditureList `json:"expenditures"`
	SumTotal     string                   `json:"sumTotal"`
}

type ProjectExpenditureList struct {
	ID          int64  `json:"id"`
	Sequence    int64  `json:"sequence"`
	Name        string `json:"name"`
	TotalPrice  string `json:"totalPrice"`
	IsFixedCost bool   `json:"isFixedCost"`
}

type CreateProjectExpenditureBody struct {
	Name        string `json:"name" validate:"required"`
	IsFixedCost *bool  `json:"isFixedCost" validate:"required"`
}

var InitialProjectExpenditures = []ProjectExpenditure{
	{
		Sequence: 1,
		Name:     "Operasional",
	},
	{
		Sequence: 2,
		Name:     "Upah Pekerja",
	},
	{
		Sequence: 3,
		Name:     "Fotokopi Berkas + Meterai",
	},
	{
		Sequence:    4,
		Name:        "BPJS Konstruksi",
		IsFixedCost: new(bool), // false
	},
}
