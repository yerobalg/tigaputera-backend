package model

import (
	"gorm.io/gorm"
)

type LedgerType string

const (
	Debit  LedgerType = "Pemasukan"
	Credit LedgerType = "Pengeluaran"
)

type InspectorLedger struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `gorm:"index" json:"createdBy"`
	UpdatedBy *int64         `json:"updatedBy"`
	DeletedBy *int64         `json:"deletedBy"`

	InspectorID    int64      `json:"inspectorId"`
	LedgerType     LedgerType `gorm:"not null;type:varchar(255)" json:"ledgerType"`
	RefID          *int64     `gorm:"default:0" json:"refId"`
	Ref            string     `gorm:"default:'Direktur'" json:"ref"`
	Amount         int64      `gorm:"not null" json:"amount"`
	CurrentBalance int64      `gorm:"default:0" json:"currentBalance"`
	FinalBalance   int64      `gorm:"default:0" json:"finalBalance"`
	ReceiptURL     string     `gorm:"type:varchar(255);default:''" json:"receiptUrl"`
	IsCanceled     *bool      `gorm:"default:false" json:"isCanceled"`
	Inspector      User       `gorm:"foreignKey:InspectorID" json:"inspector"`
}

type InspectorLedgerParam struct {
	InspectorID   int64 `form:"inspector_id"`
	RefID         *int64
	IntervalMonth int64 `form:"interval_month"`
	PaginationParam
}

type CreateInspectorIncomeBody struct {
	Amount int64  `json:"amount" binding:"required"`
	Ref    string `json:"ref" binding:"required"`
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
}
