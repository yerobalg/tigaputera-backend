package model

import (
	"gorm.io/gorm"
)

type LedgerType string

const (
	Debit  LedgerType = "Pemasukan"
	Credit LedgerType = "Pengeluaran"
)

type Ledger struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `json:"createdBy"`
	UpdatedBy *int64         `json:"updatedBy"`
	DeletedBy *int64         `json:"deletedBy"`

	InspectorID             int64      `json:"inspectorId"`
	ProjectID               int64      `json:"projectId"`
	LedgerType              LedgerType `gorm:"not null;type:varchar(255)" json:"ledgerType"`
	RefID                   *int64     `gorm:"default:0" json:"refId"`
	Ref                     string     `gorm:"default:'Direktur'" json:"ref"`
	Description             *string    `gorm:"type:varchar(255);default:''" json:"description"`
	Amount                  int64      `gorm:"not null" json:"amount"`
	Price                   int64      `gorm:"not null" json:"price"`
	TotalPrice              int64      `gorm:"not null" json:"totalPrice"`
	CurrentInspectorBalance *int64     `gorm:"default:0" json:"currentBalance"`
	FinalInspectorBalance   *int64     `gorm:"default:0" json:"finalBalance"`
	CurrentProjectBalance   *int64     `gorm:"default:0" json:"currentProjectBalance"`
	FinalProjectBalance     *int64     `gorm:"default:0" json:"finalProjectBalance"`
	ReceiptURL              string     `gorm:"type:varchar(255);default:''" json:"receiptUrl"`
	IsCanceled              *bool      `gorm:"default:false" json:"isCanceled"`
	Inspector               User       `gorm:"foreignKey:InspectorID" json:"inspector"`
	Project                 Project    `gorm:"foreignKey:ProjectID" json:"project"`
}

type LedgerParam struct {
	InspectorID   int64 `form:"inspector_id"`
	ProjectID     int64 `uri:"project_id" param:"project_id"`
	RefID         *int64
	IntervalMonth int64 `form:"interval_month"`
	PaginationParam
}

type CreateProjectIncomeBody struct {
	Amount int64  `json:"amount" form:"amount" validate:"required"`
	Ref    string `json:"ref" form:"ref" validate:"required"`
}

type InspectorLedgerResponse struct {
	Account      InspectorLedgerAccount       `json:"account"`
	Transactions []InspectorLedgerTransaction `json:"transactions"`
}

type InspectorLedgerAccount struct {
	InspectorID    int64  `json:"inspectorId"`
	InspectorName  string `json:"inspectorName"`
	CurrentBalance string `json:"currentBalance"`
}

type InspectorLedgerTransaction struct {
	Timestamp     int64  `json:"timestamp"`
	InspectorName string `json:"inspectorName"`
	Type          string `json:"type"`
	RefName       string `json:"refName"`
	Amount        string `json:"amount"`
	RecieptURL    string `json:"receiptUrl"`
}

type ProjectLedgerResponse struct {
	Account      ProjectLedgerAccount         `json:"account"`
	Transactions []InspectorLedgerTransaction `json:"transactions"`
}

type ProjectLedgerAccount struct {
	ProjectID      int64  `json:"projectId"`
	ProjectName    string `json:"projectName"`
	InspectorName  string `json:"inspectorName"`
	CurrentBalance string `json:"currentBalance"`
}

type CreateExpenditureDetailBody struct {
	Name   string `json:"name" form:"name" validate:"required"`
	Price  int64  `json:"price" form:"price" validate:"required"`
	Amount int64  `json:"amount" form:"amount" validate:"required"`
}

type ExpenditureDetailParam struct {
	ID            int64 `uri:"transaction_id" param:"transaction_id"`
	ProjectID     int64 `uri:"project_id" param:"project_id"`
	ExpenditureID int64 `uri:"expenditure_id" param:"expenditure_id"`
	InspectorID   int64 `param:"inspector_id"`
	PaginationParam
}

type ExpenditureDetailList struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Price      string `json:"price"`
	Amount     int64  `json:"amount"`
	TotalPrice string `json:"totalPrice"`
	ReceiptURL string `json:"receiptUrl"`
}

type ExpenditureDetailListResponse struct {
	ExpenditureName string                  `json:"expenditureName"`
	ProjectName     string                  `json:"projectName"`
	InspectorName   string                  `json:"inspectorName"`
	Details         []ExpenditureDetailList `json:"details"`
	SumTotal        string                  `json:"sumTotal"`
}
