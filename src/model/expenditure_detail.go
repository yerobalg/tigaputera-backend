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
	ProjectID     int64              `json:"projectId"`
	InspectorID   int64              `json:"inspectorId"`
	Expenditure   ProjectExpenditure `gorm:"foreignKey:ExpenditureID" json:"expenditure"`
	Project       Project            `gorm:"foreignKey:ProjectID" json:"project"`
	Inspector     User               `gorm:"foreignKey:InspectorID" json:"inspector"`
}

type CreateExpenditureDetailBody struct {
	Name   string `json:"name" validate:"required"`
	Price  int64  `json:"price" validate:"required"`
	Amount int64  `json:"amount" validate:"required"`
}

type ExpenditureDetailParam struct {
	ProjectID           int64 `uri:"project_id" param:"project_id"`
	ExpenditureID       int64 `uri:"expenditure_id" param:"expenditure_id"`
	ExpenditureDetailID int64 `uri:"expenditure_detail_id" param:"expenditure_detail_id"`
	PaginationParam
}

type ExpenditureDetailList struct {
	Name       string `json:"name"`
	Price      int64  `json:"price"`
	Amount     int64  `json:"amount"`
	TotalPrice int64  `json:"totalPrice"`
}

type ExpenditureDetailListResponse struct {
	ExpenditureName string                  `json:"expenditureName"`
	ProjectName     string                  `json:"projectName"`
	InspectorName   string                  `json:"inspectorName"`
	Details         []ExpenditureDetailList `json:"details"`
	SumTotal        int64                   `json:"sumTotal"`
}
