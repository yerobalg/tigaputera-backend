package controller

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"tigaputera-backend/sdk/auth"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/sdk/number"
	"tigaputera-backend/src/model"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// @Summary Refresh Statistics
// @Description Refresh Statistics
// @Tags Statistics
// @Produce json
// @Param scheduler-key header string true "scheduler-key"
// @Success 200 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
// @Router /v1/user/statistics/refresh [PUT]
func (r *rest) RefreshStatistics(c *gin.Context) {
	schedulerKey := c.Request.Header.Get("scheduler-key")
	if schedulerKey != os.Getenv("SCHEDULER_KEY") {
		r.ErrorResponse(c, errors.Unauthorized("scheduler-key tidak valid"))
		return
	}

	ctx := c.Request.Context()
	tx := r.db.WithContext(ctx).Begin()

	// delete all inspector stats
	if err := tx.
		Unscoped().
		Where("1 = 1").
		Delete(&model.MqtInspectorStats{}).
		Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	// delete all project stats
	if err := tx.
		Unscoped().
		Where("1 = 1").
		Delete(&model.MqtProjectStats{}).
		Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	users := []model.User{}
	if err := tx.
		Where("role = ?", model.Inspector).
		Find(&users).
		Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
	}

	intervalMonths := []int{1, 3, 6, 12}

	inspectorStats, err := r.getInspectorStats(tx, users, intervalMonths)
	if err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	if err := tx.Create(&inspectorStats).Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	r.SuccessResponse(c, "Berhasil memperbarui statistik", nil, nil)
}

func (r *rest) getInspectorStats(
	tx *gorm.DB,
	users []model.User,
	intervalMonths []int,
) ([]model.MqtInspectorStats, error) {
	var inspectorStats []model.MqtInspectorStats

	for _, intervalMonth := range intervalMonths {
		startDate := time.Now().UTC().AddDate(0, -intervalMonth, 0)

		now := time.Now().UTC()
		starDateUnix := time.Date(
			startDate.Year(),
			startDate.Month(),
			startDate.Day(),
			0,
			0,
			0,
			0,
			startDate.Location(),
		).Unix()
		endDateUnix := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			0,
			0,
			0,
			0,
			now.Location(),
		).Unix()

		// insert all inspector stats
		projectCountData, err := r.getInspectorProjectCount(tx, starDateUnix, 0)
		if err != nil {
			return inspectorStats, err
		}

		totalExpenditureData, err := r.getInspectorExpenditureCount(tx, starDateUnix, 0)
		if err != nil {
			return inspectorStats, err
		}

		totalIncomeData, err := r.sumTotalIncome(tx, starDateUnix, 0)
		if err != nil {
			return inspectorStats, err
		}

		margin := totalIncomeData - totalExpenditureData.Total

		allInspectorStat := model.MqtInspectorStats{
			StartTime:                starDateUnix,
			EndTime:                  endDateUnix,
			IntervalMonth:            int64(intervalMonth),
			InspectorID:              new(int64), // 0
			InspectorUsername:        "All",
			TotalDrainageProject:     &projectCountData.Drainage,
			TotalAshpaltProject:      &projectCountData.Ashpalt,
			TotalConcreteProject:     &projectCountData.Concrete,
			TotalBuildingProject:     &projectCountData.Building,
			TotalProject:             &projectCountData.Total,
			TotalDrainageExpenditure: &totalExpenditureData.Drainage,
			TotalAshpaltExpenditure:  &totalExpenditureData.Ashpalt,
			TotalConcreteExpenditure: &totalExpenditureData.Concrete,
			TotalBuildingExpenditure: &totalExpenditureData.Building,
			TotalExpenditure:         &totalExpenditureData.Total,
			TotalIncome:              &totalIncomeData,
			Margin:                   &margin,
		}

		inspectorStats = append(inspectorStats, allInspectorStat)

		// insert each inspector stats
		for _, user := range users {
			projectCountData, err := r.getInspectorProjectCount(tx, starDateUnix, user.ID)
			if err != nil {
				return inspectorStats, err
			}

			totalExpenditureData, err := r.getInspectorExpenditureCount(tx, starDateUnix, user.ID)
			if err != nil {
				return inspectorStats, err
			}

			totalIncomeData, err := r.sumTotalIncome(tx, starDateUnix, user.ID)
			if err != nil {
				return inspectorStats, err
			}

			margin := totalIncomeData - totalExpenditureData.Total

			inspectorStat := model.MqtInspectorStats{
				StartTime:                starDateUnix,
				EndTime:                  endDateUnix,
				IntervalMonth:            int64(intervalMonth),
				InspectorID:              &[]int64{user.ID}[0],
				InspectorUsername:        user.Username,
				TotalDrainageProject:     &projectCountData.Drainage,
				TotalAshpaltProject:      &projectCountData.Ashpalt,
				TotalConcreteProject:     &projectCountData.Concrete,
				TotalBuildingProject:     &projectCountData.Building,
				TotalProject:             &projectCountData.Total,
				TotalDrainageExpenditure: &totalExpenditureData.Drainage,
				TotalAshpaltExpenditure:  &totalExpenditureData.Ashpalt,
				TotalConcreteExpenditure: &totalExpenditureData.Concrete,
				TotalBuildingExpenditure: &totalExpenditureData.Building,
				TotalExpenditure:         &totalExpenditureData.Total,
				TotalIncome:              &totalIncomeData,
				Margin:                   &margin,
			}

			inspectorStats = append(inspectorStats, inspectorStat)
		}
	}

	return inspectorStats, nil
}

func (r *rest) getProjectStats(
	tx *gorm.DB,
	projects []model.Project,
) ([]model.MqtProjectStats, error) {
	var projectStats []model.MqtProjectStats
	intervalMonths := []int{1} // TODO: change to 1, 3, 6, 12
	for _, intervalMonth := range intervalMonths {
		startDate := time.Now().UTC().AddDate(0, -intervalMonth, 0)

		now := time.Now().UTC()
		starDateUnix := time.Date(
			startDate.Year(),
			startDate.Month(),
			startDate.Day(),
			0,
			0,
			0,
			0,
			startDate.Location(),
		).Unix()
		endDateUnix := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			0,
			0,
			0,
			0,
			now.Location(),
		).Unix()

		for _, project := range projects {
			totalExpenditure, err := r.sumExpenditureByType(tx, starDateUnix, "", 0, project.ID)
			if err != nil {
				return projectStats, err
			}

			totalIncome, err := r.sumIncomeByType(tx, starDateUnix, "", 0, project.ID)
			if err != nil {
				return projectStats, err
			}

			margin := totalIncome - totalExpenditure
			projectStat := model.MqtProjectStats{
				StartTime:        starDateUnix,
				EndTime:          endDateUnix,
				IntervalMonth:    int64(intervalMonth),
				ProjectID:        &[]int64{project.ID}[0],
				TotalExpenditure: &totalExpenditure,
				TotalIncome:      &totalIncome,
				Margin:           &margin,
			}

			projectStats = append(projectStats, projectStat)
		}
	}

	return projectStats, nil
}

func (r *rest) getInspectorProjectCount(
	tx *gorm.DB,
	starDateUnix int64,
	userID int64,
) (model.ProjectData, error) {
	var totalProjectStats model.ProjectData

	totalDrainage, err := r.countProjectByType(tx, starDateUnix, "Drainase", userID)
	if err != nil {
		return totalProjectStats, err
	}

	totalAshpalt, err := r.countProjectByType(tx, starDateUnix, "Hotmix", userID)
	if err != nil {
		return totalProjectStats, err
	}

	totalConcrete, err := r.countProjectByType(tx, starDateUnix, "Beton", userID)
	if err != nil {
		return totalProjectStats, err
	}

	totalBuilding, err := r.countProjectByType(tx, starDateUnix, "Bangunan", userID)
	if err != nil {
		return totalProjectStats, err
	}

	totalProjectStats.Drainage = totalDrainage
	totalProjectStats.Ashpalt = totalAshpalt
	totalProjectStats.Concrete = totalConcrete
	totalProjectStats.Building = totalBuilding
	totalProjectStats.Total = totalDrainage + totalAshpalt + totalConcrete + totalBuilding

	return totalProjectStats, nil
}

func (r *rest) countProjectByType(
	tx *gorm.DB,
	startDateUnix int64,
	projectType string,
	inspectorID int64,
) (int64, error) {
	whereQuery := "deleted_at IS NULL AND created_at >= ? AND type = ?"
	whereQueryArgs := []interface{}{startDateUnix, projectType}
	if inspectorID != 0 {
		whereQuery += " AND inspector_id = ?"
		whereQueryArgs = append(whereQueryArgs, inspectorID)
	}

	var total int64
	if err := tx.
		Model(&model.Project{}).
		Where(whereQuery, whereQueryArgs...).
		Count(&total).
		Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (r *rest) getInspectorExpenditureCount(
	tx *gorm.DB,
	startDateUnix int64,
	inspectorID int64,
) (model.ProjectData, error) {
	var totalExpenditureStats model.ProjectData

	totalDrainage, err := r.sumExpenditureByType(tx, startDateUnix, "Drainase", inspectorID, 0)
	if err != nil {
		return totalExpenditureStats, err
	}

	totalAshpalt, err := r.sumExpenditureByType(tx, startDateUnix, "Hotmix", inspectorID, 0)
	if err != nil {
		return totalExpenditureStats, err
	}

	totalConcrete, err := r.sumExpenditureByType(tx, startDateUnix, "Beton", inspectorID, 0)
	if err != nil {
		return totalExpenditureStats, err
	}

	totalBuilding, err := r.sumExpenditureByType(tx, startDateUnix, "Bangunan", inspectorID, 0)
	if err != nil {
		return totalExpenditureStats, err
	}

	totalExpenditureStats.Drainage = totalDrainage
	totalExpenditureStats.Ashpalt = totalAshpalt
	totalExpenditureStats.Concrete = totalConcrete
	totalExpenditureStats.Building = totalBuilding
	totalExpenditureStats.Total = totalDrainage + totalAshpalt + totalConcrete + totalBuilding

	return totalExpenditureStats, nil
}

func (r *rest) sumExpenditureByType(
	tx *gorm.DB,
	startDateUnix int64,
	projectType string,
	inspectorID int64,
	projectID int64,
) (int64, error) {
	whereQuery := "IL.deleted_at IS NULL AND IL.created_at >= ? AND IL.ledger_type = ?"
	whereQueryArgs := []interface{}{startDateUnix, model.Credit}
	if inspectorID != 0 {
		whereQuery += " AND IL.inspector_id = ?"
		whereQueryArgs = append(whereQueryArgs, inspectorID)
	}
	if projectID != 0 {
		whereQuery += " AND IL.project_id = ?"
		whereQueryArgs = append(whereQueryArgs, projectID)
	}

	joinQuery := "INNER JOIN projects P ON P.id = IL.project_id AND 1=?"
	joinQueryArgs := []interface{}{1}
	if projectType != "" {
		joinQuery += " AND P.type = ?"
		joinQueryArgs = append(joinQueryArgs, projectType)
	}

	var total int64
	if err := tx.
		Table("ledgers IL").
		Select("COALESCE(SUM(IL.total_price), 0) AS total").
		Joins(joinQuery, joinQueryArgs...).
		Where(whereQuery, whereQueryArgs...).
		Scan(&total).
		Error; err != nil {
		return 0, err
	}

	return -total, nil
}

func (r *rest) sumIncomeByType(
	tx *gorm.DB,
	startDateUnix int64,
	projectType string,
	inspectorID int64,
	projectID int64,
) (int64, error) {
	whereQuery := "IL.deleted_at IS NULL AND IL.created_at >= ? AND IL.ledger_type = ?"
	whereQueryArgs := []interface{}{startDateUnix, model.Debit}
	if inspectorID != 0 {
		whereQuery += " AND IL.inspector_id = ?"
		whereQueryArgs = append(whereQueryArgs, inspectorID)
	}
	if projectID != 0 {
		whereQuery += " AND IL.project_id = ?"
		whereQueryArgs = append(whereQueryArgs, projectID)
	}

	joinQuery := "INNER JOIN projects P ON P.id = IL.project_id AND 1=?"
	joinQueryArgs := []interface{}{1}
	if projectType != "" {
		joinQuery += " AND P.type = ?"
		joinQueryArgs = append(joinQueryArgs, projectType)
	}

	var total int64
	if err := tx.
		Table("ledgers IL").
		Select("COALESCE(SUM(IL.total_price), 0) AS total").
		Joins(joinQuery, joinQueryArgs...).
		Where(whereQuery, whereQueryArgs...).
		Scan(&total).
		Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (r *rest) sumTotalIncome(
	tx *gorm.DB,
	startDateUnix int64,
	inspectorID int64,
) (int64, error) {
	whereQuery := "deleted_at IS NULL AND created_at >= ? AND ledger_type = ?"
	whereQueryArgs := []interface{}{startDateUnix, model.Debit}
	if inspectorID != 0 {
		whereQuery += " AND inspector_id = ?"
		whereQueryArgs = append(whereQueryArgs, inspectorID)
	}

	var total int64
	if err := tx.
		Model(&model.Ledger{}).
		Select("COALESCE(SUM(total_price), 0) AS total").
		Where(whereQuery, whereQueryArgs...).
		Scan(&total).
		Error; err != nil {
		return 0, err
	}

	return total, nil
}

// @Summary Get User Stats
// @Description Get user statistics
// @Tags Statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.HTTPResponse{data=model.InspectorStatsResponse{}}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/user/statistics [GET]
func (r *rest) GetUserStats(c *gin.Context) {
	ctx := c.Request.Context()
	user := auth.GetUser(ctx)

	var userStatsParam model.InspectorStatsParam
	if user.Role == string(model.Inspector) {
		userStatsParam.InspectorID = user.ID
	} else {
		userStatsParam.InspectorID = 0
	}

	var totalProject int64
	var totalExpenditure int64
	var totalIncome int64
	var totalMargin int64

	var userStats model.MqtInspectorStats
	err := r.db.WithContext(ctx).
		Where(
			"inspector_id = ? AND interval_month = 1",
			userStatsParam.InspectorID,
		).
		Take(&userStats).Error

	if r.isNoRecordFound(err) {
		totalProject = 0
		totalExpenditure = 0
		totalIncome = 0
		totalMargin = 0
	} else if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	} else {
		totalProject = *userStats.TotalProject
		totalExpenditure = *userStats.TotalExpenditure
		totalIncome = *userStats.TotalIncome
		totalMargin = *userStats.Margin
	}

	userStatsResponse := model.InspectorStatsResponse{
		TotalProject:     totalProject,
		TotalExpenditure: number.ConvertToRupiah(totalExpenditure),
		TotalIncome:      number.ConvertToRupiah(totalIncome),
		Margin:           number.ConvertToRupiah(totalMargin),
	}

	r.SuccessResponse(c, "Berhasil mendapatkan statistik pengguna", userStatsResponse, nil)
}

// @Summary Get User Stats Detail
// @Description Get user statistics detail
// @Tags Statistics
// @Produce json
// @Security BearerAuth
// @Param interval_month query int false "interval_month"
// @Param user_id query integer false "user_id"
// @Success 200 {object} model.HTTPResponse{data=model.InspectorStatsDetailResponse}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/user/statistics/detail [GET]
func (r *rest) GetUserStatsDetail(c *gin.Context) {
	ctx := c.Request.Context()

	var userStatsParam model.InspectorStatsParam
	if err := r.BindParam(c, &userStatsParam); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	user := auth.GetUser(ctx)
	if user.Role == string(model.Inspector) {
		userStatsParam.InspectorID = user.ID
	}

	intervalMonth := int(userStatsParam.IntervalMonth)
	if intervalMonth == 0 {
		intervalMonth = 1
	}

	var userStats model.MqtInspectorStats
	err := r.db.WithContext(ctx).
		Where(
			"inspector_id = ? AND interval_month = ?",
			userStatsParam.InspectorID,
			intervalMonth,
		).
		Take(&userStats).Error

	var inspectorStatsDetailResponse model.InspectorStatsDetailResponse

	if r.isNoRecordFound(err) {
		emptyUserStats := model.MqtInspectorStats{
			StartTime:                userStatsParam.StartTime,
			EndTime:                  time.Time{}.Unix(),
			IntervalMonth:            userStatsParam.IntervalMonth,
			InspectorID:              &userStatsParam.InspectorID,
			InspectorUsername:        user.Username,
			TotalDrainageProject:     new(int64), // 0
			TotalAshpaltProject:      new(int64), // 0
			TotalConcreteProject:     new(int64), // 0
			TotalBuildingProject:     new(int64), // 0
			TotalProject:             new(int64), // 0
			TotalDrainageExpenditure: new(int64), // 0
			TotalAshpaltExpenditure:  new(int64), // 0
			TotalConcreteExpenditure: new(int64), // 0
			TotalBuildingExpenditure: new(int64), // 0
			TotalExpenditure:         new(int64), // 0
			TotalIncome:              new(int64), // 0
			Margin:                   new(int64), // 0
		}

		inspectorStatsDetailResponse = r.getUserStatsDetailResponse(emptyUserStats)
	} else if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	} else {
		inspectorStatsDetailResponse = r.getUserStatsDetailResponse(userStats)
	}

	r.SuccessResponse(
		c,
		"Berhasil mendapatkan detail statistik pengguna",
		inspectorStatsDetailResponse, nil,
	)
}

func (r *rest) getUserStatsDetailResponse(
	userStats model.MqtInspectorStats,
) model.InspectorStatsDetailResponse {
	projectCount := model.TotalProjectStats{
		TotalProject: *userStats.TotalProject,
		Drainage: model.Stats{
			Name:  "Drainase",
			Total: *userStats.TotalDrainageProject,
			Percentage: number.GetPercentage(
				*userStats.TotalDrainageProject,
				*userStats.TotalProject,
			),
		},
		Ashpalt: model.Stats{
			Name:  "Hotmix",
			Total: *userStats.TotalAshpaltProject,
			Percentage: number.GetPercentage(
				*userStats.TotalAshpaltProject,
				*userStats.TotalProject,
			),
		},
		Concrete: model.Stats{
			Name:  "Beton",
			Total: *userStats.TotalConcreteProject,
			Percentage: number.GetPercentage(
				*userStats.TotalConcreteProject,
				*userStats.TotalProject,
			),
		},
		Building: model.Stats{
			Name:  "Bangunan",
			Total: *userStats.TotalBuildingProject,
			Percentage: number.GetPercentage(
				*userStats.TotalBuildingProject,
				*userStats.TotalProject,
			),
		},
	}

	expenditure := model.TotalExpenditureStats{
		TotalExpenditure: number.ConvertToRupiah(*userStats.TotalExpenditure),
		Drainage: model.StatsString{
			Name:  "Drainase",
			Total: number.ConvertToRupiah(*userStats.TotalDrainageExpenditure),
			Percentage: number.GetPercentage(
				*userStats.TotalDrainageExpenditure,
				*userStats.TotalExpenditure,
			),
		},
		Ashpalt: model.StatsString{
			Name:  "Hotmix",
			Total: number.ConvertToRupiah(*userStats.TotalAshpaltExpenditure),
			Percentage: number.GetPercentage(
				*userStats.TotalAshpaltExpenditure,
				*userStats.TotalExpenditure,
			),
		},
		Concrete: model.StatsString{
			Name:  "Beton",
			Total: number.ConvertToRupiah(*userStats.TotalConcreteExpenditure),
			Percentage: number.GetPercentage(
				*userStats.TotalConcreteExpenditure,
				*userStats.TotalExpenditure,
			),
		},
		Building: model.StatsString{
			Name:  "Bangunan",
			Total: number.ConvertToRupiah(*userStats.TotalBuildingExpenditure),
			Percentage: number.GetPercentage(
				*userStats.TotalBuildingExpenditure,
				*userStats.TotalExpenditure,
			),
		},
	}

	return model.InspectorStatsDetailResponse{
		LastUpdated:       userStats.EndTime,
		InspectorID:       *userStats.InspectorID,
		InspectorUsername: userStats.InspectorUsername,
		IntervalMonth:     userStats.IntervalMonth,
		ProjectCount:      projectCount,
		Expenditure:       expenditure,
		Income:            number.ConvertToRupiah(*userStats.TotalIncome),
		Margin:            number.ConvertToRupiah(*userStats.Margin),
	}
}

// @Summary Create Ledger Report
// @Description Create Ledger Report
// @Tags Statistics
// @Produce json
// @Param scheduler-key header string true "scheduler-key"
// @Success 200 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
// @Router /v1/user/statistics/ledger-report [POST]
func (r *rest) CreateLedgerReport(c *gin.Context) {
	ctx := c.Request.Context()

	if c.Request.Header.Get("scheduler-key") != os.Getenv("SCHEDULER_KEY") {
		r.ErrorResponse(c, errors.Unauthorized("scheduler-key tidak valid"))
		return
	}

	reportName := []string{
		"1_month_ledger.xlsx",
		"3_month_ledger.xlsx",
		"6_month_ledger.xlsx",
		"12_month_ledger.xlsx",
	}

	for _, name := range reportName {
		// skip if file not exist
		_ = r.storage.Delete(ctx, name, "report")
	}

	projects := []model.Project{}
	if err := r.db.WithContext(ctx).InnerJoins("Inspector").
		Find(&projects).
		Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	intervalMonths := []int{1, 3, 6, 12}

	for _, intervalMonth := range intervalMonths {
		startDate := time.Now().UTC().AddDate(0, -intervalMonth, 0)

		starDateUnix := time.Date(
			startDate.Year(),
			startDate.Month(),
			startDate.Day(),
			0,
			0,
			0,
			0,
			startDate.Location(),
		).Unix()

		now := time.Now().UTC()
		currentDateUnix := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			0,
			0,
			0,
			0,
			now.Location(),
		).Unix()

		f := excelize.NewFile()

		for _, project := range projects {
			ledgers := []model.Ledger{}
			if err := r.db.WithContext(ctx).
				Where(
					"project_id = ? AND created_at >= ?",
					project.ID,
					starDateUnix,
				).
				Order("created_at").
				Find(&ledgers).
				Error; err != nil {
				r.ErrorResponse(c, errors.InternalServerError(err.Error()))
				return
			}

			if err := r.createNewSheetsExcel(f, project, ledgers, intervalMonth, currentDateUnix); err != nil {
				r.ErrorResponse(c, errors.InternalServerError(err.Error()))
				return
			}
		}

		f.DeleteSheet("Sheet1")

		//convert excelize to multipart.File
		excelBytes, err := f.WriteToBuffer()
		if err != nil {
			r.ErrorResponse(c, errors.InternalServerError(err.Error()))
			return
		}

		excelReader := bytes.NewReader(excelBytes.Bytes())

		_, err = r.storage.UploadFromBytes(
			ctx,
			excelReader,
			fmt.Sprintf("%d_month_ledger.xlsx", intervalMonth),
			"ledger_report",
		);

		if err != nil {
			r.ErrorResponse(c, errors.InternalServerError(err.Error()))
			return
		}
	}

	r.SuccessResponse(c, "Berhasil membuat laporan buku kas", nil, nil)
}
func (r *rest) createNewSheetsExcel(
	f *excelize.File,
	project model.Project,
	projectLedgers []model.Ledger,
	intervalMonth int,
	currentTime int64,
) error {
	// Create a new sheet.
	_, err := f.NewSheet(project.Name)
	if err != nil {
		return err
	}

	for i := 1; i <= 9; i++ {
		col1 := fmt.Sprintf("A%d", i)
		col2 := fmt.Sprintf("B%d", i)
		col3 := fmt.Sprintf("C%d", i)
		col4 := fmt.Sprintf("D%d", i)
		err := f.MergeCell(project.Name, col1, col2)
		if err != nil {
			return err
		}
		err = f.MergeCell(project.Name, col3, col4)
		if err != nil {
			return err
		}
	}

	f.SetCellValue(project.Name, "A1", "Nama Proyek")
	f.SetCellValue(project.Name, "A2", "Nama Pekerjaan")
	f.SetCellValue(project.Name, "A3", "Nama Pengawas")
	f.SetCellValue(project.Name, "A4", "Kategori Proyek")
	f.SetCellValue(project.Name, "A5", "Status Proyek")
	f.SetCellValue(project.Name, "A6", "Nama Dinas")
	f.SetCellValue(project.Name, "A7", "Tanggal Mulai Proyek")
	f.SetCellValue(project.Name, "A8", "Tanggal Selesai Proyek")
	f.SetCellValue(project.Name, "A9", "Terakhir Diperbarui")

	f.SetCellValue(project.Name, "C1", project.Name)
	f.SetCellValue(project.Name, "C2", project.Description)
	f.SetCellValue(project.Name, "C3", project.Inspector.Username)
	f.SetCellValue(project.Name, "C4", project.Type)
	f.SetCellValue(project.Name, "C5", project.Status)
	f.SetCellValue(project.Name, "C6", project.DeptName)
	f.SetCellValue(project.Name, "C7", r.convertDateToString(project.StartDate))
	f.SetCellValue(project.Name, "C8", r.convertDateToString(project.FinalDate))
	f.SetCellValue(project.Name, "C9", r.convertDateToString(currentTime))

	projectInfostyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
	})
	if err != nil {
		return err
	}

	err = f.SetCellStyle(project.Name, "A1", "D9", projectInfostyle)
	if err != nil {
		return err
	}

	ledgerHeaderStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Border: []excelize.Border{
			{
				Type:  "left",
				Color: "#000000",
			},
			{
				Type:  "top",
				Color: "#000000",
			},
			{
				Type:  "right",
				Color: "#000000",
			},
			{
				Type:  "bottom",
				Color: "#000000",
			},
		},
	})
	if err != nil {
		return err
	}

	f.MergeCell(project.Name, "A11", "F11")
	f.SetCellValue(project.Name, "A11", fmt.Sprintf("BUKU KAS %d BULAN TERAKHIR", intervalMonth))
	f.SetCellStyle(project.Name, "A11", "F11", ledgerHeaderStyle)

	f.SetCellValue(project.Name, "A12", "No")
	f.SetCellValue(project.Name, "B12", "Tanggal")
	f.SetCellValue(project.Name, "C12", "Keterangan")
	f.SetCellValue(project.Name, "D12", "Debit")
	f.SetCellValue(project.Name, "E12", "Kredit")
	f.SetCellValue(project.Name, "F12", "Saldo")

	err = f.SetCellStyle(project.Name, "A12", "F12", ledgerHeaderStyle)
	if err != nil {
		return err
	}

	row := 13
	for i, ledger := range projectLedgers {
		row = i + 13
		var description string
		if ledger.LedgerType == model.Debit {
			description = "Pemasukan dari direktur"
		} else {
			description = *ledger.Description + " - " + ledger.Ref
		}
		f.SetCellValue(project.Name, fmt.Sprintf("A%d", row), i+1)                                          // No
		f.SetCellValue(project.Name, fmt.Sprintf("B%d", row), r.convertLocalDateToString(ledger.CreatedAt)) // Tanggal
		f.SetCellValue(project.Name, fmt.Sprintf("C%d", row), description)                                  // Keterangan
		if ledger.LedgerType == model.Debit {
			f.SetCellValue(project.Name, fmt.Sprintf("D%d", row), number.ConvertToRupiah(ledger.TotalPrice))
		} else {
			f.SetCellValue(project.Name, fmt.Sprintf("E%d", row), number.ConvertToRupiah(-ledger.TotalPrice))
		}
		f.SetCellValue(project.Name, fmt.Sprintf("F%d", row), number.ConvertToRupiah(*ledger.FinalProjectBalance))
	}

	ledgerBodyStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 12,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Border: []excelize.Border{
			{
				Type:  "left",
				Color: "#000000",
			},
			{
				Type:  "top",
				Color: "#000000",
			},
			{
				Type:  "right",
				Color: "#000000",
			},
			{
				Type:  "bottom",
				Color: "#000000",
			},
		},
	})
	if err != nil {
		return err
	}

	err = f.SetCellStyle(project.Name, "A12", fmt.Sprintf("F%d", row), ledgerBodyStyle)
	if err != nil {
		return err
	}

	return nil
}

func (r *rest) convertDateToString(date int64) string {
	return time.Unix(date, 0).Format("02-01-2006 15:04:05")
}

func (r *rest) convertLocalDateToString(date int64) string {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	jakartaDate := time.Unix(date, 0).In(loc)
	return jakartaDate.Format("02-01-2006 15:04:05")
}
