package model

type MqtProjectStats struct {
	StartTime                int64  `gorm:"column:start_time"`
	EndTime                  int64  `gorm:"column:end_time"`
	IntervalMonth            int64  `gorm:"column:interval_month"`
	ProjectID                *int64 `gorm:"column:project_id"`
	TotalDrainageExpenditure *int64 `gorm:"column:total_drainage_expenditure"`
	TotalAshpaltExpenditure  *int64 `gorm:"column:total_ashpalt_expenditure"`
	TotalConcreteExpenditure *int64 `gorm:"column:total_concrete_expenditure"`
	TotalBuildingExpenditure *int64 `gorm:"column:total_building_expenditure"`
	TotalExpenditure         *int64 `gorm:"column:total_expenditure"`
	TotalDrainageIncome      *int64 `gorm:"column:total_drainage_income"`
	TotalAshpaltIncome       *int64 `gorm:"column:total_ashpalt_income"`
	TotalConcreteIncome      *int64 `gorm:"column:total_concrete_income"`
	TotalBuildingIncome      *int64 `gorm:"column:total_building_income"`
	TotalIncome              *int64 `gorm:"column:total_income"`
	Balance                  *int64 `gorm:"column:balance"`
}
