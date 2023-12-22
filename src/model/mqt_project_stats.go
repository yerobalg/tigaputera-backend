package model

type MqtProjectStats struct {
	StartTime        int64  `gorm:"column:start_time"`
	EndTime          int64  `gorm:"column:end_time"`
	IntervalMonth    int64  `gorm:"column:interval_month"`
	ProjectID        *int64 `gorm:"column:project_id"`
	TotalExpenditure *int64 `gorm:"column:total_expenditure"`
	TotalIncome      *int64 `gorm:"column:total_income"`
	Margin           *int64 `gorm:"column:balance"`
}
