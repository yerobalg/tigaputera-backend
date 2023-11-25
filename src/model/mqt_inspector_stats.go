package model

type MqtInspectorStats struct {
	StartTime                int64  `gorm:"column:start_time"`
	EndTime                  int64  `gorm:"column:end_time"`
	IntervalMonth            int64  `gorm:"column:interval_month"`
	InspectorID              *int64 `gorm:"column:inspector_id"`
	InspectorUsername        string `gorm:"column:inspector_username"`
	TotalDrainageProject     *int64 `gorm:"column:total_drainage_project"`
	TotalAshpaltProject      *int64 `gorm:"column:total_ashpalt_project"`
	TotalConcreteProject     *int64 `gorm:"column:total_concrete_project"`
	TotalBuildingProject     *int64 `gorm:"column:total_building_project"`
	TotalProject             *int64 `gorm:"column:total_project"`
	TotalDrainageExpenditure *int64 `gorm:"column:total_drainage_expenditure"`
	TotalAshpaltExpenditure  *int64 `gorm:"column:total_ashpalt_expenditure"`
	TotalConcreteExpenditure *int64 `gorm:"column:total_concrete_expenditure"`
	TotalBuildingExpenditure *int64 `gorm:"column:total_building_expenditure"`
	TotalExpenditure         *int64 `gorm:"column:total_expenditure"`
	TotalIncome              *int64 `gorm:"column:total_income"`
	Margin                   *int64 `gorm:"column:margin"`
}

type InspectorStatsParam struct {
	IntervalMonth int64 `form:"interval_month"`
	UserID        int64 `form:"user_id"`
	StartTime     int64
}

type InspectorStatsDetailResponse struct {
	LastUpdated       int64                 `json:"lastUpdated"`
	InspectorID       int64                 `json:"inspectorID"`
	InspectorUsername string                `json:"inspectorUsername"`
	IntervalMonth     int64                 `json:"intervalMonth"`
	ProjectCount      TotalProjectStats     `json:"projectCount"`
	Expenditure       TotalExpenditureStats `json:"expenditure"`
	Income            string                `json:"income"`
	Margin            string                `json:"margin"`
}

type InspectorStatsResponse struct {
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

type ProjectData struct {
	Drainage int64 `json:"drainage"`
	Ashpalt  int64 `json:"ashpalt"`
	Concrete int64 `json:"concrete"`
	Building int64 `json:"building"`
	Total    int64 `json:"total"`
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
