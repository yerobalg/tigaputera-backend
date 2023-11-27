package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"tigaputera-backend/sdk/auth"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/src/model"
)

// @Summary Create Project Expenditure Detail
// @Description Create project expenditure detail
// @Tags Project Expenditure Detail
// @Produce json
// @Security BearerAuth
// @Param project_id path  int true "project_id"
// @Param expenditure_id path  int true "expenditure_id"
// @Param createExpenditureDetailBody body model.CreateExpenditureDetailBody true "body"
// @Success 201 {object} model.HTTPResponse{}
// @Failure 400 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 404 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project/{project_id}/expenditure/{expenditure_id}/detail [POST]
func (r *rest) CreateProjectExpenditureDetail(c *gin.Context) {
	ctx := c.Request.Context()
	var param model.ExpenditureDetailParam
	var body model.CreateExpenditureDetailBody

	if err := r.BindParam(c, &param); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	if err := r.BindBody(c, &body); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	if err := r.validator.ValidateStruct(body); err != nil {
		r.ErrorResponse(c, errors.BadRequest(err.Error()))
		return
	}

	user := auth.GetUser(ctx)
	var projectExpenditure model.ProjectExpenditure

	err := r.db.WithContext(ctx).
		InnerJoins("Project", r.db.Where(&model.Project{InspectorID: user.ID})).
		First(&projectExpenditure, param.ExpenditureID).Error
	if r.isNoRecordFound(err) {
		r.ErrorResponse(c, errors.NotFound("pengeluaran proyek tidak ditemukan"))
		return
	} else if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	// Get inspector's balance
	var inspectorLedger model.InspectorLedger
	inspectorLedgerParam := model.InspectorLedgerParam{
		InspectorID: user.ID,
	}
	err = r.db.WithContext(ctx).
		Where(&inspectorLedgerParam).
		Order("created_at desc").
		Take(&inspectorLedger).Error

	if r.isNoRecordFound(err) {
		r.ErrorResponse(c, errors.BadRequest("Saldo tidak mencukupi"))
		return
	} else if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
	}

	totalPrice := body.Price * body.Amount
	expenditureDetail := model.ExpenditureDetail{
		Name:          body.Name,
		Price:         body.Price,
		Amount:        body.Amount,
		TotalPrice:    totalPrice,
		ReceiptURL:    "", // TODO: Upload receipt
		ExpenditureID: projectExpenditure.ID,
		ProjectID:     projectExpenditure.ProjectID,
		InspectorID:   user.ID,
	}

	projectExpenditure.TotalPrice += expenditureDetail.TotalPrice
	projectExpenditure.UpdatedBy = &user.ID

	tx := r.db.WithContext(ctx).Begin()

	if err := tx.Create(&expenditureDetail).Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	var newLedger model.InspectorLedger

	if inspectorLedger.FinalBalance < totalPrice {
		tx.Rollback()
		r.ErrorResponse(c, errors.BadRequest("Saldo tidak mencukupi"))
		return
	} else {
		newLedger = model.InspectorLedger{
			InspectorID:    user.ID,
			LedgerType:     model.Credit,
			Ref:            fmt.Sprintf("%s Proyek %s", body.Name, projectExpenditure.Project.Name),
			RefID:          &expenditureDetail.ID,
			Amount:         totalPrice * -1,
			CurrentBalance: inspectorLedger.FinalBalance,
			FinalBalance:   inspectorLedger.FinalBalance - totalPrice,
		}
	}

	if err := tx.Create(&newLedger).Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	if err := tx.Save(&projectExpenditure).Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	if err := tx.Commit().Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	r.CreatedResponse(c, "Berhasil membuat detail pengeluaran proyek", nil)
}

// @Summary Get List Project Expenditure Detail
// @Description Get list project expenditure detail
// @Tags Project Expenditure Detail
// @Produce json
// @Security BearerAuth
// @Param project_id path  int true "project_id"
// @Param expenditure_id path  int true "expenditure_id"
// @Success 200 {object} model.HTTPResponse{data=model.ExpenditureDetailListResponse}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 404 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project/{project_id}/expenditure/{expenditure_id}/detail [GET]
func (r *rest) GetProjectExpenditureDetailList(c *gin.Context) {
	ctx := c.Request.Context()
	var param model.ExpenditureDetailParam

	if err := r.BindParam(c, &param); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	rows, err := r.db.WithContext(ctx).
		Model(&model.ExpenditureDetail{}).
		InnerJoins(
			"Expenditure",
			r.db.Where(&model.ProjectExpenditure{ProjectID: param.ProjectID}),
		).
		InnerJoins(
			"Project",
			r.db.Where(&model.Project{ID: param.ProjectID}),
		).
		InnerJoins("Project.Inspector").
		Rows()

	if r.isNoRecordFound(err) {
		r.ErrorResponse(
			c,
			errors.NotFound("pengeluaran proyek tidak ditemukan"),
		)
		return
	} else if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	defer rows.Close()
	var expenditureDetailResponse model.ExpenditureDetailListResponse
	for rows.Next() {
		var expenditureDetail model.ExpenditureDetail
		if err := r.db.ScanRows(rows, &expenditureDetail); err != nil {
			r.ErrorResponse(c, errors.InternalServerError(err.Error()))
			return
		}

		expenditureDetailResponse.ExpenditureName = expenditureDetail.Expenditure.Name
		expenditureDetailResponse.ProjectName = expenditureDetail.Project.Name
		expenditureDetailResponse.InspectorName = expenditureDetail.Project.Inspector.Name

		expenditureDetailList := model.ExpenditureDetailList{
			Name:       expenditureDetail.Name,
			Price:      expenditureDetail.Price,
			Amount:     expenditureDetail.Amount,
			TotalPrice: expenditureDetail.TotalPrice,
		}

		expenditureDetailResponse.Details = append(
			expenditureDetailResponse.Details,
			expenditureDetailList,
		)
	}

	r.SuccessResponse(c, "Berhasil mendapatkan detail pengeluaran proyek", expenditureDetailResponse, nil)
}

// @Summary Delete Project Expenditure Detail
// @Description Delete project expenditure detail
// @Tags Project Expenditure Detail
// @Produce json
// @Security BearerAuth
// @Param project_id path  int true "project_id"
// @Param expenditure_id path  int true "expenditure_id"
// @Param expenditure_detail_id path  int true "expenditure_detail_id"
// @Success 200 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 404 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project/{project_id}/expenditure/{expenditure_id}/detail/{expenditure_detail_id} [DELETE]
func (r *rest) DeleteProjectExpenditureDetail(c *gin.Context) {
	ctx := c.Request.Context()
	var param model.ExpenditureDetailParam

	if err := r.BindParam(c, &param); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	user := auth.GetUser(ctx)
	param.InspectorID = user.ID
	var expenditureDetail model.ExpenditureDetail

	err := r.db.WithContext(ctx).
		InnerJoins("Project").
		InnerJoins("Expenditure").
		InnerJoins("Inspector").
		Where(&param).
		First(&expenditureDetail).Error
	if r.isNoRecordFound(err) {
		r.ErrorResponse(c, errors.NotFound("detail pengeluaran proyek tidak ditemukan"))
		return
	} else if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	expenditureDetail.Expenditure.TotalPrice -= expenditureDetail.TotalPrice

	var latestLedger model.InspectorLedger

	if err := r.db.WithContext(ctx).
		Where("inspector_id = ?", user.ID).
		Order("created_at desc").
		Take(&latestLedger).
		Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	inspectorLedger := model.InspectorLedger{
		InspectorID:    user.ID,
		LedgerType:     model.Debit,
		Ref:            "Pembatalan pengeluaran proyek",
		RefID:          &expenditureDetail.ID,
		Amount:         expenditureDetail.TotalPrice,
		CurrentBalance: latestLedger.FinalBalance,
		FinalBalance:   latestLedger.FinalBalance + expenditureDetail.TotalPrice,
		IsCanceled:     &[]bool{true}[0],
	}

	tx := r.db.WithContext(ctx).Begin()

	if err := tx.Delete(&expenditureDetail).Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	if err := tx.Save(&expenditureDetail.Expenditure).Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	if err := tx.
		Model(&model.InspectorLedger{}).
		Create(&inspectorLedger).
		Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
	}

	if err := tx.Commit().Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	r.SuccessResponse(c, "Berhasil menghapus detail pengeluaran proyek", nil, nil)
}
