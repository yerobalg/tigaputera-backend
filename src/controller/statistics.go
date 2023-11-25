package controller

import (
	"gorm.io/gorm"
	"os"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/src/model"
	"time"

	"github.com/gin-gonic/gin"
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
		r.ErrorResponse(c, errors.Unauthorized("Scheduler key is not valid"))
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
			InspectorID:              new(int64),  // 0
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
