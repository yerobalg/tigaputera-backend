package controller

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"os"
	"tigaputera-backend/sdk/auth"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/sdk/number"
	"tigaputera-backend/src/model"
	"time"
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

	users := []model.User{}
	if err := tx.
		Where("role = ?", model.Inspector).
		Find(&users).
		Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
	}

	intervalMonths := []int{1, 3, 6, 12}

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
		projectCountData, err := r.getAllProjectCount(tx, starDateUnix, 0)
		if err != nil {
			tx.Rollback()
			r.ErrorResponse(c, errors.InternalServerError(err.Error()))
			return
		}

		totalExpenditureData, err := r.getAllExpenditureCount(tx, starDateUnix, 0)
		if err != nil {
			tx.Rollback()
			r.ErrorResponse(c, errors.InternalServerError(err.Error()))
			return
		}

		totalIncomeData, err := r.sumTotalIncome(tx, starDateUnix, 0)
		if err != nil {
			tx.Rollback()
			r.ErrorResponse(c, errors.InternalServerError(err.Error()))
			return
		}

		margin := totalIncomeData - totalExpenditureData.Total

		allInspectorStats := model.MqtInspectorStats{
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

		if err := tx.Create(&allInspectorStats).Error; err != nil {
			tx.Rollback()
			r.ErrorResponse(c, errors.InternalServerError(err.Error()))
			return
		}

		// insert each inspector stats
		for _, user := range users {
			projectCountData, err := r.getAllProjectCount(tx, starDateUnix, user.ID)
			if err != nil {
				tx.Rollback()
				r.ErrorResponse(c, errors.InternalServerError(err.Error()))
				return
			}

			totalExpenditureData, err := r.getAllExpenditureCount(tx, starDateUnix, user.ID)
			if err != nil {
				tx.Rollback()
				r.ErrorResponse(c, errors.InternalServerError(err.Error()))
				return
			}

			totalIncomeData, err := r.sumTotalIncome(tx, starDateUnix, user.ID)
			if err != nil {
				tx.Rollback()
				r.ErrorResponse(c, errors.InternalServerError(err.Error()))
				return
			}

			margin := totalIncomeData - totalExpenditureData.Total

			inspectorStats := model.MqtInspectorStats{
				StartTime:                starDateUnix,
				EndTime:                  endDateUnix,
				IntervalMonth:            int64(intervalMonth),
				InspectorID:              &user.ID,
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

			if err := tx.Create(&inspectorStats).Error; err != nil {
				tx.Rollback()
				r.ErrorResponse(c, errors.InternalServerError(err.Error()))
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	r.SuccessResponse(c, "Berhasil memperbarui statistik", nil, nil)
}

func (r *rest) getAllProjectCount(
	tx *gorm.DB,
	starDateUnix int64,
	userID int64,
) (model.ProjectData, error) {
	var totalProjectStats model.ProjectData

	totalDrainage, err := r.countProjectByType(tx, starDateUnix, "Drainase", userID)
	if err != nil {
		tx.Rollback()
		return totalProjectStats, err
	}

	totalAshpalt, err := r.countProjectByType(tx, starDateUnix, "Hotmix", userID)
	if err != nil {
		tx.Rollback()
		return totalProjectStats, err
	}

	totalConcrete, err := r.countProjectByType(tx, starDateUnix, "Beton", userID)
	if err != nil {
		tx.Rollback()
		return totalProjectStats, err
	}

	totalBuilding, err := r.countProjectByType(tx, starDateUnix, "Bangunan", userID)
	if err != nil {
		tx.Rollback()
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

func (r *rest) getAllExpenditureCount(
	tx *gorm.DB,
	startDateUnix int64,
	inspectorID int64,
) (model.ProjectData, error) {
	var totalExpenditureStats model.ProjectData

	totalDrainage, err := r.sumExpenditureByType(tx, startDateUnix, "Drainase", inspectorID)
	if err != nil {
		tx.Rollback()
		return totalExpenditureStats, err
	}

	totalAshpalt, err := r.sumExpenditureByType(tx, startDateUnix, "Hotmix", inspectorID)
	if err != nil {
		tx.Rollback()
		return totalExpenditureStats, err
	}

	totalConcrete, err := r.sumExpenditureByType(tx, startDateUnix, "Beton", inspectorID)
	if err != nil {
		tx.Rollback()
		return totalExpenditureStats, err
	}

	totalBuilding, err := r.sumExpenditureByType(tx, startDateUnix, "Bangunan", inspectorID)
	if err != nil {
		tx.Rollback()
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
) (int64, error) {
	whereQuery := "IL.deleted_at IS NULL AND IL.created_at >= ? AND IL.ledger_type = ?"
	whereQueryArgs := []interface{}{startDateUnix, model.Credit}
	if inspectorID != 0 {
		whereQuery += " AND IL.inspector_id = ?"
		whereQueryArgs = append(whereQueryArgs, inspectorID)
	}

	var total int64
	if err := tx.
		Table("inspector_ledgers IL").
		Select("COALESCE(SUM(IL.amount), 0) AS total").
		Joins("INNER JOIN expenditure_details ED ON ED.id = IL.ref_id").
		Joins("INNER JOIN projects P ON P.id = ED.project_id AND P.type = ?", projectType).
		Where(whereQuery, whereQueryArgs...).
		Scan(&total).
		Error; err != nil {
		return 0, err
	}

	return -total, nil
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
		Model(&model.InspectorLedger{}).
		Select("COALESCE(SUM(amount), 0) AS total").
		Where(whereQuery, whereQueryArgs...).
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
		userStatsParam.UserID = user.ID
	} else {
		userStatsParam.UserID = 0
	}

	lastMonth := time.Now().UTC().AddDate(0, -1, 0)
	userStatsParam.StartTime = time.Date(
		lastMonth.Year(),
		lastMonth.Month(),
		lastMonth.Day(),
		0,
		0,
		0,
		0,
		lastMonth.Location(),
	).Unix()

	var totalProject int64
	var totalExpenditure int64
	var totalIncome int64
	var totalMargin int64

	var userStats model.MqtInspectorStats
	err := r.db.WithContext(ctx).
		Where(&userStatsParam).
		First(&userStats).Error

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
		userStatsParam.UserID = user.ID
	}

	intervalMonth := int(userStatsParam.IntervalMonth)
	if intervalMonth == 0 {
		intervalMonth = 1
	}

	beginMonth := time.Now().UTC().AddDate(0, -intervalMonth, 0)
	userStatsParam.StartTime = time.Date(
		beginMonth.Year(),
		beginMonth.Month(),
		beginMonth.Day(),
		0,
		0,
		0,
		0,
		beginMonth.Location(),
	).Unix()

	var userStats model.MqtInspectorStats
	err := r.db.WithContext(ctx).
		Where(
			"inspector_id = ? AND start_time = ?",
			userStatsParam.UserID,
			userStatsParam.StartTime,
		).
		Take(&userStats).Error

	if r.isNoRecordFound(err) {
		r.ErrorResponse(c, errors.BadRequest("Statistik pengguna tidak ditemukan"))
		return
	} else if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	inspectorStatsDetailResponse := model.InspectorStatsDetailResponse{
		LastUpdated:       userStats.EndTime,
		InspectorID:       *userStats.InspectorID,
		InspectorUsername: userStats.InspectorUsername,
		IntervalMonth:     userStats.IntervalMonth,
		ProjectCount: model.TotalProjectStats{
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
		},
		Expenditure: model.TotalExpenditureStats{
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
		},
		Income: number.ConvertToRupiah(*userStats.TotalIncome),
		Margin: number.ConvertToRupiah(*userStats.Margin),
	}

	r.SuccessResponse(
		c,
		"Berhasil mendapatkan detail statistik pengguna",
		inspectorStatsDetailResponse, nil,
	)
}
