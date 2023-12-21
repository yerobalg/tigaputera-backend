package controller

import (
	"tigaputera-backend/sdk/auth"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/sdk/number"
	"tigaputera-backend/src/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"context"
	"fmt"
	"time"
)

// @Summary Create Project Expenditure Detail
// @Description Create project expenditure detail
// @Tags Project Expenditure Detail
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "project_id"
// @Param expenditure_id path int true "expenditure_id"
// @Param name formData string true "name"
// @Param price formData int64 true "price"
// @Param amount formData int64 true "amount"
// @Param receiptImage formData file true "receiptImage"
// @Accept multipart/form-data
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

	recieptImage, err := r.getReceiptImage(c)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	user := auth.GetUser(ctx)
	projectExpenditure, err := r.getInspectorProjectExpenditureByID(ctx, param, user.ID)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	inspectorLedger, err := r.getLatestLedger(ctx, user.ID)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	tx := r.db.WithContext(ctx).Begin()

	recieptImage.SetFileName(r.getExpenditureReceiptName(
		user.Username,
		projectExpenditure,
	))

	recieptURL, err := r.storage.Upload(
		ctx,
		recieptImage,
		"expenditures",
	)
	if err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	totalPrice := body.Price * body.Amount
	expenditureDetail := r.getExpenditureDetailInput(
		body,
		recieptURL,
		projectExpenditure,
		user.ID,
	)

	projectExpenditure.TotalPrice += expenditureDetail.TotalPrice
	projectExpenditure.UpdatedBy = &user.ID

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
	}

	ledgerRef := fmt.Sprintf(
		"%s Proyek %s",
		body.Name,
		projectExpenditure.Project.Name,
	)
	newLedger = model.InspectorLedger{
		InspectorID:    user.ID,
		LedgerType:     model.Credit,
		Ref:            ledgerRef,
		RefID:          &expenditureDetail.ID,
		Amount:         totalPrice * -1,
		CurrentBalance: inspectorLedger.FinalBalance,
		FinalBalance:   inspectorLedger.FinalBalance - totalPrice,
		ReceiptURL:     recieptURL,
	}

	if err := r.insertNewLedger(ctx, tx, newLedger, projectExpenditure); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	r.CreatedResponse(c, "Berhasil membuat detail pengeluaran proyek", nil)
}

func (r *rest) getInspectorProjectExpenditureByID(
	ctx context.Context,
	param model.ExpenditureDetailParam,
	inspectorID int64,
) (model.ProjectExpenditure, error) {
	var projectExpenditure model.ProjectExpenditure

	err := r.db.WithContext(ctx).
		InnerJoins("Project", r.db.Where(&model.Project{InspectorID: inspectorID})).
		First(&projectExpenditure, param.ExpenditureID).Error
	if r.isNoRecordFound(err) {
		return projectExpenditure, errors.NotFound("pengeluaran proyek tidak ditemukan")
	} else if err != nil {
		return projectExpenditure, errors.InternalServerError(err.Error())
	}

	return projectExpenditure, nil
}

func (r *rest) getLatestLedger(
	ctx context.Context,
	inspectorID int64,
) (model.InspectorLedger, error) {
	var latestLedger model.InspectorLedger
	err := r.db.WithContext(ctx).
		Where("inspector_id = ?", inspectorID).
		Order("created_at desc").
		Take(&latestLedger).Error
	if r.isNoRecordFound(err) {
		return latestLedger, errors.BadRequest("Saldo tidak mencukupi")
	} else if err != nil {
		return latestLedger, errors.InternalServerError(err.Error())
	}

	return latestLedger, nil
}

func (r *rest) getExpenditureReceiptName(
	username string,
	projectExpenditure model.ProjectExpenditure,
) string {
	now := time.Now().Unix()
	return fmt.Sprintf(
		"%s_%d_%d_%d", // username_projectId_expenditureId_timestamp
		username,
		projectExpenditure.Project.ID,
		projectExpenditure.ID,
		now,
	)
}

func (r *rest) getExpenditureDetailInput(
	body model.CreateExpenditureDetailBody,
	recieptURL string,
	projectExpenditure model.ProjectExpenditure,
	userID int64,
) model.ExpenditureDetail {
	return model.ExpenditureDetail{
		Name:          body.Name,
		Price:         body.Price,
		Amount:        body.Amount,
		TotalPrice:    body.Price * body.Amount,
		ReceiptURL:    recieptURL,
		ProjectID:     projectExpenditure.ProjectID,
		ExpenditureID: projectExpenditure.ID,
		InspectorID:   userID,
	}
}

func (r *rest) insertNewLedger(
	ctx context.Context,
	tx *gorm.DB,
	newLedger model.InspectorLedger,
	projectExpenditure model.ProjectExpenditure,
) error {
	if err := tx.Create(&newLedger).Error; err != nil {
		tx.Rollback()
		return errors.InternalServerError(err.Error())
	}

	if err := tx.Save(&projectExpenditure).Error; err != nil {
		tx.Rollback()
		return errors.InternalServerError(err.Error())
	}

	return nil
}



// @Summary Get List Project Expenditure Detail
// @Description Get list project expenditure detail
// @Tags Project Expenditure Detail
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "project_id"
// @Param expenditure_id path int true "expenditure_id"
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

	projectExpenditure, err := r.getProjectExpenditureByID(ctx, param)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	var inspector model.User
	if err := r.db.WithContext(ctx).
		First(&inspector, projectExpenditure.Project.InspectorID).
		Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
	}

	rows, err := r.db.WithContext(ctx).
		Model(&model.ExpenditureDetail{}).
		Where(
			"expenditure_id = ? AND project_id = ?",
			param.ExpenditureID, param.ProjectID,
		).
		Rows()

	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}
	defer rows.Close()

	expenditureDetails := []model.ExpenditureDetailList{}
	var sumTotal int64
	for rows.Next() {
		var expenditureDetail model.ExpenditureDetail
		if err := r.db.ScanRows(rows, &expenditureDetail); err != nil {
			r.ErrorResponse(c, errors.InternalServerError(err.Error()))
			return
		}

		expenditureDetailList := r.getExpenditureDetailList(expenditureDetail)
		expenditureDetails = append(expenditureDetails, expenditureDetailList)
		sumTotal += expenditureDetail.TotalPrice
	}

	expenditureDetailListResponse := model.ExpenditureDetailListResponse{
		ExpenditureName: projectExpenditure.Name,
		ProjectName:     projectExpenditure.Project.Name,
		InspectorName:   inspector.Name,
		Details:         expenditureDetails,
		SumTotal:        number.ConvertToRupiah(sumTotal),
	}

	r.SuccessResponse(c, "Berhasil mendapatkan detail pengeluaran proyek", expenditureDetailListResponse, nil)
}

func (r *rest) getProjectExpenditureByID(
	ctx context.Context,
	param model.ExpenditureDetailParam,
) (model.ProjectExpenditure, error) {
	var projectExpenditure model.ProjectExpenditure

	err := r.db.WithContext(ctx).
		InnerJoins("Project", r.db.Where(&model.Project{ID: param.ProjectID})).
		First(&projectExpenditure, param.ExpenditureID).Error
	if r.isNoRecordFound(err) {
		return projectExpenditure, errors.NotFound("pengeluaran proyek tidak ditemukan")
	} else if err != nil {
		return projectExpenditure, errors.InternalServerError(err.Error())
	}

	return projectExpenditure, nil
}

func (r *rest) getExpenditureDetailList(
	expenditureDetail model.ExpenditureDetail,
) model.ExpenditureDetailList {
	return model.ExpenditureDetailList{
		ID:         expenditureDetail.ID,
		Name:       expenditureDetail.Name,
		Price:      number.ConvertToRupiah(expenditureDetail.Price),
		Amount:     expenditureDetail.Amount,
		TotalPrice: number.ConvertToRupiah(expenditureDetail.TotalPrice),
		ReceiptURL: expenditureDetail.ReceiptURL,
	}
}

// @Summary Delete Project Expenditure Detail
// @Description Delete project expenditure detail
// @Tags Project Expenditure Detail
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "project_id"
// @Param expenditure_id path int true "expenditure_id"
// @Param expenditure_detail_id path int true "expenditure_detail_id"
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
