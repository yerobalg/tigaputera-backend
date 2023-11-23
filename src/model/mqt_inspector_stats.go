package model

type MqtUserStats struct {
	StartTime                int64  `gorm:"column:start_time"`
	EndTime                  int64  `gorm:"column:end_time"`
	IntervalMonth            int64  `gorm:"column:interval_month"`
	InspectorID              *int64 `gorm:"column:inspector_id"`
	InspectorUsername        string `gorm:"column:inspector_username"`
	TotalDrainageProject     *int64 `gorm:"column:total_drainage_project"`
	TotalAshpaltProject      *int64 `gorm:"column:total_ashpalt_project"`
	TotalConcreteProject     *int64 `gorm:"column:total_concrete_project"`
	TotalBuildingProject     *int64 `gorm:"column:total_building_project"`
	TotalDrainageExpenditure *int64 `gorm:"column:total_drainage_expenditure"`
	TotalAshpaltExpenditure  *int64 `gorm:"column:total_ashpalt_expenditure"`
	TotalConcreteExpenditure *int64 `gorm:"column:total_concrete_expenditure"`
	TotalBuildingExpenditure *int64 `gorm:"column:total_building_expenditure"`
	TotalDrainageIncome      *int64 `gorm:"column:total_drainage_income"`
	TotalAshpaltIncome       *int64 `gorm:"column:total_ashpalt_income"`
	TotalConcreteIncome      *int64 `gorm:"column:total_concrete_income"`
	TotalBuildingIncome      *int64 `gorm:"column:total_building_income"`
}

type UserStatsParam struct {
	StartTime   int64 `form:"startTime"`
	UserID      int64 
}

type UserStatsDetailResponse struct {
	LastUpdated       int64                 `json:"lastUpdated"`
	InspectorID       int64                 `json:"inspectorID"`
	InspectorUsername string                `json:"inspectorUsername"`
	IntervalMonth     int64                 `json:"intervalMonth"`
	Project           TotalProjectStats     `json:"project"`
	Expenditure       TotalExpenditureStats `json:"expenditure"`
	Income            TotalIncomeStats      `json:"income"`
	Margin            MarginStats           `json:"margin"`
}

type UserStatsResponse struct {
	TotalProject     int64  `json:"totalProject"`
	TotalExpenditure string `json:"totalExpenditure"`
	TotalIncome      string `json:"totalIncome"`
	Margin           string `json:"margin"`
}

type TotalProjectStats struct {
	TotalProject int64 `json:"totalProject"`
	Drainage     Stats `json:"drainage"`
	Ashpalt      Stats `json:"ashpalt"`
	Concrete     Stats `json:"concrete"`
	Building     Stats `json:"building"`
}

type TotalExpenditureStats struct {
	TotalExpenditure string      `json:"totalExpenditure"`
	Drainage         StatsString `json:"drainage"`
	Ashpalt          StatsString `json:"ashpalt"`
	Concrete         StatsString `json:"concrete"`
	Building         StatsString `json:"building"`
}

type TotalIncomeStats struct {
	TotalIncome string      `json:"totalIncome"`
	Drainage    StatsString `json:"drainage"`
	Ashpalt     StatsString `json:"ashpalt"`
	Concrete    StatsString `json:"concrete"`
	Building    StatsString `json:"building"`
}

type MarginStats struct {
	TotalMargin string      `json:"totalMargin"`
	Drainage    StatsString `json:"drainage"`
	Ashpalt     StatsString `json:"ashpalt"`
	Concrete    StatsString `json:"concrete"`
	Building    StatsString `json:"building"`
}

type Stats struct {
	Name       string  `json:"name"`
	Total      int64   `json:"total"`
	Percentage float64 `json:"percentage"`
}

type StatsString struct {
	Name       string  `json:"name"`
	Total      string  `json:"total"`
	Percentage float64 `json:"percentage"`
}

func GetTotalProject(userStats MqtUserStats) int64 {
	return *userStats.TotalDrainageProject + 
		*userStats.TotalAshpaltProject + 
		*userStats.TotalConcreteProject + 
		*userStats.TotalBuildingProject
}

func GetTotalExpenditure(userStats MqtUserStats) int64 {
	return *userStats.TotalDrainageExpenditure + 
		*userStats.TotalAshpaltExpenditure + 
		*userStats.TotalConcreteExpenditure + 
		*userStats.TotalBuildingExpenditure
}

func GetTotalIncome(userStats MqtUserStats) int64 {
	return *userStats.TotalDrainageIncome + 
		*userStats.TotalAshpaltIncome + 
		*userStats.TotalConcreteIncome + 
		*userStats.TotalBuildingIncome
}
