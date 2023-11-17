package model

type ProjectExpenditure struct {
	ID        int64  `gorm:"primaryKey;index:idx_project_expenditure,unique" json:"id"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
	DeletedAt int64  `gorm:"index" json:"-"`
	CreatedBy *int64 `json:"createdBy"`
	UpdatedBy *int64 `json:"updatedBy"`
	DeletedBy *int64 `json:"deletedBy"`

	Sequence   int64   `gorm:"not null;index:idx_project_expenditure,unique" json:"sequence"`
	Name       string  `gorm:"not null;type:varchar(255)" json:"name"`
	TotalPrice int64   `gorm:"default:0" json:"totalPrice"`
	ProjectID  int64   `json:"projectId"`
	Project    Project `gorm:"foreignKey:ProjectID" json:"project"`
}
