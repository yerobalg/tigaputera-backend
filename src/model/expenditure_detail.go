package model

type ExpenditureDetail struct {
	ID        int64  `gorm:"primaryKey" json:"id"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
	DeletedAt int64  `gorm:"index" json:"-"`
	CreatedBy *int64 `json:"createdBy"`
	UpdatedBy *int64 `json:"updatedBy"`
	DeletedBy *int64 `json:"deletedBy"`

	Name          string             `gorm:"not null;type:varchar(255)" json:"name"`
	Price         int64              `json:"price"`
	Amount        int64              `json:"amount"`
	TotalPrice    int64              `json:"totalPrice"`
	ReceiptURL    string             `json:"receiptUrl"`
	ExpenditureID int64              `json:"expenditureId"`
	Expenditure   ProjectExpenditure `gorm:"foreignKey:ExpenditureID" json:"expenditure"`
}