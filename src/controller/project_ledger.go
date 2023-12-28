package controller

import (
	"github.com/gin-gonic/gin"
	"tigaputera-backend/sdk/auth"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/sdk/file"
	"tigaputera-backend/sdk/number"
	"tigaputera-backend/src/model"
	"gorm.io/gorm"

	"context"
	"database/sql"
	"fmt"
	"time"
)

// @Summary Create Project Income
// @Description Create project income
// @Tags Project Ledger
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "project_id"
// @Param amount formData int64 true "amount"
// @Param ref formData string true "ref"
// @Param receiptImage formData file true "receiptImage"
// @Accept multipart/form-data
// @Success 201 {object} model.HTTPResponse{}
// @Failure 400 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project/{project_id}/income [POST]
func (r *rest) CreateIncomeTransaction(c *gin.Context) {
	ctx := c.Request.Context()
	var reqBody model.CreateProjectIncomeBody

	if err := r.BindBody(c, &reqBody); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	var param model.LedgerParam
	if err := r.BindParam(c, &param); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	if err := r.validator.ValidateStruct(reqBody); err != nil {
		r.ErrorResponse(c, errors.BadRequest(err.Error()))
		return
	}

	recieptImage, err := r.getReceiptImage(c)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	user := auth.GetUser(ctx)
	projectId := param.ProjectID
	latestLedger, err := r.getLatestLedger(ctx, user.ID, projectId)
	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	now := time.Now().Unix()
	recieptImage.SetFileName(fmt.Sprintf(
		"%s_%d_%d", // username_projectId_timestamp
		user.Username,
		projectId,
		now,
	))

	recieptURL, err := r.storage.Upload(ctx, recieptImage, "incomes")
	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	ledgerDesc := "Pemasukan"
	prevInspectorBalance := latestLedger.FinalInspectorBalance
	prevProjectBalance := latestLedger.FinalProjectBalance
	finalInspectorBalance := *prevInspectorBalance + reqBody.Amount
	finalProjectBalance := *prevProjectBalance + reqBody.Amount
	newLedger := model.Ledger{
		InspectorID:             user.ID,
		ProjectID:               projectId,
		LedgerType:              model.Debit,
		Ref:                     reqBody.Ref,
		Amount:                  1,
		Price:                   reqBody.Amount,
		TotalPrice:              reqBody.Amount,
		Description:             &ledgerDesc,
		CurrentInspectorBalance: prevInspectorBalance,
		FinalInspectorBalance:   &finalInspectorBalance,
		CurrentProjectBalance:   prevProjectBalance,
		FinalProjectBalance:     &finalProjectBalance,
		ReceiptURL:              recieptURL,
	}

	latestLedger.Project.UpdatedBy = &user.ID
	*latestLedger.Project.Income += reqBody.Amount

	if err := r.insertIncome(ctx, newLedger, latestLedger); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	r.CreatedResponse(c, "Berhasil menambahkan pemasukan pengawas", nil)
}

func (r *rest) getReceiptImage(c *gin.Context) (*file.File, error) {
	var receiptImage *file.File
	var err error

	receiptImage, err = file.Init(c, "receiptImage")
	if err != nil {
		return nil, errors.BadRequest("Gambar bukti tidak ditemukan")
	}

	if !receiptImage.IsImage() {
		return nil, errors.BadRequest("Gambar bukti harus berupa png, jpg, atau jpeg")
	}

	return receiptImage, nil
}

func (r *rest) getLatestLedger(
	ctx context.Context,
	inspectorID int64,
	projectId int64,
) (model.Ledger, error) {
	latestLedger := model.Ledger{
		CurrentInspectorBalance: new(int64),
		FinalInspectorBalance:   new(int64),
		CurrentProjectBalance:   new(int64),
		FinalProjectBalance:     new(int64),
		Project: model.Project{
			ID:     projectId,
			Income: new(int64),
		},
	}
	err := r.db.WithContext(ctx).
		Where("inspector_id = ?", inspectorID).
		Order("created_at desc").
		Take(&latestLedger).Error

	if r.isNoRecordFound(err) {
		return latestLedger, nil
	} else if err != nil {
		return latestLedger, err
	}

	var projectLedger model.Ledger
	err = r.db.WithContext(ctx).
		InnerJoins("Project").
		Where(model.Ledger{
			ProjectID:   projectId,
			InspectorID: inspectorID,
		}).
		Order("created_at desc").
		Take(&projectLedger).Error
	if r.isNoRecordFound(err) {
		latestLedger.CurrentProjectBalance = new(int64)
		latestLedger.FinalProjectBalance = new(int64)
		return latestLedger, nil
	} else if err != nil {
		return latestLedger, err
	}

	latestLedger.CurrentProjectBalance = projectLedger.CurrentProjectBalance
	latestLedger.FinalProjectBalance = projectLedger.FinalProjectBalance

	return latestLedger, nil
}

func (r *rest) insertIncome(
	ctx context.Context,
	incomeTrans model.Ledger,
	latestLedger model.Ledger,
) error {
	tx := r.db.WithContext(ctx).Begin()

	if err := tx.Create(&incomeTrans).Error; err != nil {
		tx.Rollback()
		return err
	}

	updateProject := map[string]interface{}{
		"income":     gorm.Expr("income + ?", incomeTrans.TotalPrice),
		"updated_by": incomeTrans.InspectorID,
	}

	if err := tx.Model(&model.Project{}).
		Where("id = ?", incomeTrans.ProjectID).
		Updates(updateProject).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// @Summary Create Project Expenditure Transaction
// @Description Create project expenditure detail
// @Tags Project Ledger
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
// @Router /v1/project/{project_id}/expenditure/{expenditure_id}/transaction [POST]
func (r *rest) CreateExpenditureTransaction(c *gin.Context) {
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

	projectID := projectExpenditure.ProjectID
	inspectorLedger, err := r.getLatestLedger(ctx, user.ID, projectID)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	totalPrice := body.Price * body.Amount
	prevProjectBalance := *inspectorLedger.FinalProjectBalance
	if prevProjectBalance < totalPrice {
		r.ErrorResponse(c, errors.BadRequest("Saldo anda tidak mencukupi"))
		return
	}

	recieptURL, err := r.getReceiptURL(
		ctx,
		recieptImage,
		user.Username,
		projectExpenditure,
	)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	prevInspectorBalance := *inspectorLedger.FinalInspectorBalance
	finalInspectorBalance := prevInspectorBalance - totalPrice
	finalProjectBalance := prevProjectBalance - totalPrice
	expenditureTransaction := model.Ledger{
		InspectorID:             user.ID,
		ProjectID:               projectID,
		LedgerType:              model.Credit,
		RefID:                   &projectExpenditure.ID,
		Ref:                     projectExpenditure.Name,
		Description:             &body.Name,
		Amount:                  body.Amount,
		Price:                   -body.Price,
		TotalPrice:              -totalPrice,
		CurrentInspectorBalance: &prevInspectorBalance,
		FinalInspectorBalance:   &finalInspectorBalance,
		CurrentProjectBalance:   &prevProjectBalance,
		FinalProjectBalance:     &finalProjectBalance,
		ReceiptURL:              recieptURL,
	}

	projectExpenditure.UpdatedBy = &user.ID

	if err := r.insertExpenditure(
		ctx,
		expenditureTransaction,
		projectExpenditure,
	); err != nil {
		r.ErrorResponse(c, err)
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

func (r *rest) getReceiptURL(
	ctx context.Context,
	receiptImage *file.File,
	username string,
	projectExpenditure model.ProjectExpenditure,
) (string, error) {
	now := time.Now().Unix()
	imageName := fmt.Sprintf(
		"%s_%d_%d_%d", // username_projectId_expenditureId_timestamp
		username,
		projectExpenditure.Project.ID,
		projectExpenditure.ID,
		now,
	)

	receiptImage.SetFileName(imageName)

	receipt, err := r.storage.Upload(
		ctx,
		receiptImage,
		"expenditures",
	)
	if err != nil {
		return receipt, err
	}

	return receipt, nil
}

func (r *rest) insertExpenditure(
	ctx context.Context,
	expenditureTrans model.Ledger,
	projectExpenditure model.ProjectExpenditure,
) error {
	expenditurePrice := *projectExpenditure.TotalPrice - expenditureTrans.TotalPrice
	expenditureUpdate := model.ProjectExpenditure{
		TotalPrice: &expenditurePrice,
		UpdatedBy:  &expenditureTrans.InspectorID,
	}

	tx := r.db.WithContext(ctx).Begin()

	if err := tx.Create(&expenditureTrans).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&model.ProjectExpenditure{}).
		Where("id = ?", projectExpenditure.ID).
		Updates(expenditureUpdate).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&model.Project{}).
		Where("id = ?", projectExpenditure.ProjectID).
		Update("updated_by", expenditureTrans.InspectorID).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// @Summary Get List Project Expenditure Detail
// @Description Get list project expenditure detail
// @Tags Project Ledger
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "project_id"
// @Param expenditure_id path int true "expenditure_id"
// @Success 200 {object} model.HTTPResponse{data=model.ExpenditureDetailListResponse}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 404 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project/{project_id}/expenditure/{expenditure_id}/transaction [GET]
func (r *rest) GetExpenditureTransactionList(c *gin.Context) {
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
		Model(&model.Ledger{}).
		Where(
			`inspector_id = ? AND 
			project_id = ? AND
			ref_id = ? AND
			ledger_type = ? AND
			is_canceled = ?`,
			inspector.ID,
			projectExpenditure.ProjectID,
			projectExpenditure.ID,
			model.Credit,
			false,
		).
		Order("created_at desc").
		Rows()

	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}
	defer rows.Close()

	expenditureTrans := []model.ExpenditureDetailList{}
	var sumTotal int64
	for rows.Next() {
		var expenditure model.Ledger
		if err := r.db.ScanRows(rows, &expenditure); err != nil {
			r.ErrorResponse(c, errors.InternalServerError(err.Error()))
			return
		}

		expenditure.TotalPrice = -expenditure.TotalPrice
		expenditure.Price = -expenditure.Price
		expenditureDetail := r.getExpenditureDetailList(expenditure)
		expenditureTrans = append(expenditureTrans, expenditureDetail)
		sumTotal += expenditure.TotalPrice
	}

	expenditureDetailListResponse := model.ExpenditureDetailListResponse{
		ExpenditureName: projectExpenditure.Name,
		ProjectName:     projectExpenditure.Project.Name,
		InspectorName:   inspector.Name,
		Details:         expenditureTrans,
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
	expenditureTrans model.Ledger,
) model.ExpenditureDetailList {
	return model.ExpenditureDetailList{
		ID:         expenditureTrans.ID,
		Name:       *expenditureTrans.Description,
		Price:      number.ConvertToRupiah(expenditureTrans.Price),
		Amount:     expenditureTrans.Amount,
		TotalPrice: number.ConvertToRupiah(expenditureTrans.TotalPrice),
		ReceiptURL: expenditureTrans.ReceiptURL,
	}
}

// @Summary Delete Project Expenditure Detail
// @Description Delete project expenditure detail
// @Tags Project Ledger
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "project_id"
// @Param expenditure_id path int true "expenditure_id"
// @Param transaction_id path int true "transaction_id"
// @Success 200 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 404 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project/{project_id}/expenditure/{expenditure_id}/transaction/{transaction_id} [DELETE]
func (r *rest) DeleteExpenditureTransaction(c *gin.Context) {
	ctx := c.Request.Context()
	var param model.ExpenditureDetailParam

	if err := r.BindParam(c, &param); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	user := auth.GetUser(ctx)
	var expenditureDetail model.Ledger
	err := r.db.WithContext(ctx).
		Where(
			`inspector_id = ? AND project_id = ?`,
			user.ID,
			param.ProjectID,
		).
		First(&expenditureDetail, param.ID).Error
	if r.isNoRecordFound(err) {
		r.ErrorResponse(c, errors.NotFound("transaksi pengeluaran proyek tidak ditemukan"))
		return
	} else if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	} else if *expenditureDetail.IsCanceled || string(expenditureDetail.LedgerType) == string(model.Debit) {
		r.ErrorResponse(c, errors.NotFound("transaksi pengeluaran proyek tidak ditemukan"))
		return
	}

	projectExpenditure, err := r.getProjectExpenditureByID(ctx, param)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	expenditureDetail.Price = -expenditureDetail.Price
	expenditureDetail.TotalPrice = -expenditureDetail.TotalPrice
	*projectExpenditure.TotalPrice += expenditureDetail.TotalPrice
	projectExpenditureUpd := model.ProjectExpenditure{
		TotalPrice: projectExpenditure.TotalPrice,
		UpdatedBy:  &user.ID,
	}

	latestLedger, err := r.getLatestLedger(ctx, user.ID, param.ProjectID)
	if err != nil {
		r.ErrorResponse(c, errors.NotFound("pengeluaran proyek tidak ditemukan"))
		return
	}

	deleteName := "Pembatalan " + *expenditureDetail.Description
	prevInspectorBalance := *latestLedger.FinalInspectorBalance
	finalInspectorBalance := prevInspectorBalance + expenditureDetail.TotalPrice
	prevProjectBalance := *latestLedger.FinalProjectBalance
	finalProjectBalance := prevProjectBalance + expenditureDetail.TotalPrice
	canceledLedger := model.Ledger{
		InspectorID:             user.ID,
		ProjectID:               param.ProjectID,
		LedgerType:              model.Debit,
		RefID:                   &projectExpenditure.ID,
		Ref:                     projectExpenditure.Name,
		Description:             &deleteName,
		Amount:                  expenditureDetail.Amount,
		Price:                   expenditureDetail.Price,
		TotalPrice:              expenditureDetail.TotalPrice,
		CurrentInspectorBalance: &prevInspectorBalance,
		FinalInspectorBalance:   &finalInspectorBalance,
		CurrentProjectBalance:   &prevProjectBalance,
		FinalProjectBalance:     &finalProjectBalance,
	}

	tx := r.db.WithContext(ctx).Begin()

	if err := tx.Model(&model.ProjectExpenditure{}).
		Where("id = ?", projectExpenditure.ID).
		Updates(projectExpenditureUpd).Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	if err := tx.Model(&model.Project{}).
		Where("id = ?", param.ProjectID).
		Update("updated_by", user.ID).Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	if err := tx.Model(&model.Ledger{}).
		Where("id = ?", param.ID).
		Update("is_canceled", true).Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	if err := tx.Create(&canceledLedger).Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	r.SuccessResponse(c, "Berhasil menghapus detail pengeluaran proyek", nil, nil)
}

// @Summary Get List Project Transaction
// @Description Get list project transaction
// @Tags Project Ledger
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "project_id"
// @Param page query int false "page"
// @Param limit query int false "limit"
// @Param interval_month query int false "interval_month"
// @Success 200 {object} model.HTTPResponse{data=model.ProjectLedgerResponse}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 404 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project/{project_id}/ledger [GET]
func (r *rest) GetProjectLedger(c *gin.Context) {
	ctx := c.Request.Context()
	var param model.LedgerParam

	if err := r.BindParam(c, &param); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	if param.IntervalMonth <= 0 {
		param.IntervalMonth = 1
	}

	project, err := r.getProjectByID(ctx, param.ProjectID)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	param.PaginationParam.SetDefaultPagination()

	rows, err := r.getTransactionRows(ctx, &param, project)
	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	defer rows.Close()
	var sumTotal int64
	transactions := []model.InspectorLedgerTransaction{}
	for rows.Next() {
		var ledger model.Ledger
		if err := r.db.ScanRows(rows, &ledger); err != nil {
			r.ErrorResponse(c, errors.InternalServerError(err.Error()))
			return
		}

		ledgerDetail := r.getTransaction(ledger)
		transactions = append(transactions, ledgerDetail)
		sumTotal += ledger.TotalPrice
	}

	if err := r.countTransaction(
		ctx,
		project,
		&param.TotalElement,
		param.IntervalMonth,
	); err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}
	param.ProcessPagination(int64(len(transactions)))

	inspectorBalance, err := r.getLatestProjectBalance(ctx, project)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	r.SuccessResponse(
		c,
		"Berhasil mendapatkan buku kas proyek",
		r.getProjectLedgerRes(project, inspectorBalance, transactions),
		&param.PaginationParam,
	)
}

func (r *rest) getProjectByID(
	ctx context.Context,
	projectID int64,
) (model.Project, error) {
	var project model.Project

	err := r.db.WithContext(ctx).
		InnerJoins("Inspector").
		First(&project, projectID).Error
	if r.isNoRecordFound(err) {
		return project, errors.NotFound("proyek tidak ditemukan")
	} else if err != nil {
		return project, errors.InternalServerError(err.Error())
	}

	return project, nil
}

func (r *rest) getTransactionRows(
	ctx context.Context,
	param *model.LedgerParam,
	project model.Project,
) (*sql.Rows, error) {
	startTime := r.getStartTime(int(param.IntervalMonth))
	rows, err := r.db.WithContext(ctx).
		Model(&model.Ledger{}).
		Where(
			`project_id = ? AND inspector_id = ? AND created_at >= ?`,
			project.ID,
			project.InspectorID,
			startTime,
		).
		Limit(int(param.Limit)).
		Offset(int(param.Offset)).
		Order("created_at desc").
		Rows()
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *rest) countTransaction(
	ctx context.Context,
	project model.Project,
	count *int64,
	intervalMonth int64,
) error {
	startTime := r.getStartTime(int(intervalMonth))

	err := r.db.WithContext(ctx).
		Model(&model.Ledger{}).
		Where(
			`project_id = ? AND inspector_id = ? AND created_at >= ?`,
			project.ID,
			project.InspectorID,
			startTime,
		).
		Count(count).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *rest) getLatestProjectBalance(
	ctx context.Context,
	project model.Project,
) (int64, error) {
	var ledger model.Ledger
	err := r.db.WithContext(ctx).
		Where(
			`project_id = ? AND inspector_id = ?`,
			project.ID,
			project.InspectorID,
		).
		Order("created_at desc").
		First(&ledger).Error
	if r.isNoRecordFound(err) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return *ledger.FinalProjectBalance, nil
}

func (r *rest) getProjectLedgerRes(
	project model.Project,
	balance int64,
	transactions []model.InspectorLedgerTransaction,
) model.ProjectLedgerResponse {
	return model.ProjectLedgerResponse{
		Account: model.ProjectLedgerAccount{
			ProjectID:      project.ID,
			ProjectName:    project.Name,
			InspectorName:  project.Inspector.Name,
			CurrentBalance: number.ConvertToRupiah(balance),
		},
		Transactions: transactions,
	}
}
