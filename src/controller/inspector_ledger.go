package controller

import (
	"github.com/gin-gonic/gin"
	"tigaputera-backend/sdk/auth"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/sdk/number"
	"tigaputera-backend/src/model"

	"time"
)

// @Summary Create Inspector Income
// @Description Create inspector income
// @Tags Inspector Ledger
// @Produce json
// @Security BearerAuth
// @Param createInspectorIncomeBody body model.CreateInspectorIncomeBody true "body"
// @Success 201 {object} model.HTTPResponse{}
// @Failure 400 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/user/inspector/income [POST]
func (r *rest) CreateInspectorIncome(c *gin.Context) {
	ctx := c.Request.Context()
	var createInspectorIncomeBody model.CreateInspectorIncomeBody

	if err := r.BindBody(c, &createInspectorIncomeBody); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	if err := r.validator.ValidateStruct(createInspectorIncomeBody); err != nil {
		r.ErrorResponse(c, errors.BadRequest(err.Error()))
		return
	}

	var latestLedger model.InspectorLedger
	user := auth.GetUser(ctx)
	var previousBalance int64
	err := r.db.WithContext(ctx).
		Order("created_at desc").
		First(&latestLedger, user.ID).Error

	if r.isNoRecordFound(err) {
		previousBalance = 0
	} else if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	} else {
		previousBalance = latestLedger.FinalBalance
	}

	newLedger := model.InspectorLedger{
		InspectorID:    user.ID,
		LedgerType:     model.Debit,
		Ref:            createInspectorIncomeBody.Ref,
		Amount:         createInspectorIncomeBody.Amount,
		CurrentBalance: previousBalance,
		FinalBalance:   previousBalance + createInspectorIncomeBody.Amount,
	}

	if err := r.db.WithContext(ctx).Create(&newLedger).Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	r.CreatedResponse(c, "Berhasil membuat pemasukan pengawas", nil)
}

// @Summary Get Inspector Ledger
// @Description Get inspector ledger
// @Tags Inspector Ledger
// @Produce json
// @Security BearerAuth
// @Param page query int false "page"
// @Param limit query int false "limit"
// @Param interval_month query int false "interval_month"
// @Param inspector_id query int false "inspector_id"
// @Success 200 {object} model.HTTPResponse{model.InspectorLedgerResponse{}}
// @Failure 400 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/user/inspector/ledger [GET]
func (r *rest) GetInspectorLedger(c *gin.Context) {
	ctx := c.Request.Context()

	var param model.InspectorLedgerParam
	if err := r.BindParam(c, &param); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	user := auth.GetUser(ctx)
	if user.Role == string(model.Inspector) {
		param.InspectorID = user.ID
	}

	intervalMonth := int(param.IntervalMonth)
	if intervalMonth == 0 {
		intervalMonth = 1
	}

	beginMonth := time.Now().UTC().AddDate(0, -intervalMonth, 0)
	startTime := time.Date(
		beginMonth.Year(),
		beginMonth.Month(),
		beginMonth.Day(),
		0,
		0,
		0,
		0,
		beginMonth.Location(),
	).Unix()

	var userAccount model.User

	if param.InspectorID != 0 {
		if err := r.db.WithContext(ctx).
			Where("id = ?", param.InspectorID).
			First(&userAccount).Error; err != nil {
			r.ErrorResponse(c, errors.InternalServerError(err.Error()))
			return
		}
	} else {
		userAccount.Name = "Semua Pengawas"
	}

	param.PaginationParam.SetDefaultPagination()

	rows, err := r.db.WithContext(ctx).
		Model(&model.InspectorLedger{}).
		Where(&param).
		Where("created_at >= ?", startTime).
		Order("created_at desc").
		Limit(int(param.Limit)).
		Offset(int(param.Offset)).
		Rows()
	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	transactions := []model.InspectorLedgerTransaction{}

	defer rows.Close()
	for rows.Next() {
		var ledger model.InspectorLedger
		if err := r.db.WithContext(ctx).ScanRows(rows, &ledger); err != nil {
			r.ErrorResponse(c, errors.InternalServerError(err.Error()))
			return
		}

		transaction := model.InspectorLedgerTransaction{
			Timestamp:     ledger.CreatedAt,
			InspectorName: userAccount.Name,
			Type:          string(ledger.LedgerType),
			RefName:       ledger.Ref,
			Amount:        number.ConvertToRupiah(ledger.Amount),
		}

		transactions = append(transactions, transaction)
	}

	if err := r.db.WithContext(ctx).
		Model(&model.InspectorLedger{}).
		Where(&param).
		Where("created_at >= ?", startTime).
		Count(&param.TotalElement).Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	param.PaginationParam.ProcessPagination(int64(len(transactions)))

	account := model.InspectorLedgerAccount{
		InspectorID:   userAccount.ID,
		InspectorName: userAccount.Name,
	}

	var latestLedger model.InspectorLedger
	err = r.db.WithContext(ctx).
		Order("created_at desc").
		Take(&latestLedger, param.InspectorID).
		Error
	if r.isNoRecordFound(err) {
		account.CurrentBalance = number.ConvertToRupiah(0)
	} else if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	} else {
		account.CurrentBalance = number.ConvertToRupiah(latestLedger.FinalBalance)
	}

	inspectorLedgerResponse := model.InspectorLedgerResponse{
		Account:      account,
		Transactions: transactions,
	}

	r.SuccessResponse(
		c,
		"Berhasil mendapatkan buku kas pengawas",
		inspectorLedgerResponse,
		&param.PaginationParam,
	)
}
