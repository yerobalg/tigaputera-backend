package controller

import (
	"context"
	"tigaputera-backend/sdk/auth"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/sdk/number"
	"tigaputera-backend/src/model"
	"github.com/gin-gonic/gin"

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
		Where("inspector_id = ?", user.ID).
		Order("created_at desc").
		Take(&latestLedger).Error

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
// @Success 200 {object} model.HTTPResponse{data=model.InspectorLedgerResponse}
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

	var inspectorLedgerResponse model.InspectorLedgerResponse
	var err error

	if param.InspectorID == 0 {
		inspectorLedgerResponse, err = r.getAllInspectorLedger(
			ctx,
			&param.PaginationParam,
			startTime,
		)
	} else {
		inspectorLedgerResponse, err = r.getSingleInspectorLedger(
			ctx,
			&param.PaginationParam,
			startTime,
			param.InspectorID,
		)
	}

	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	r.SuccessResponse(
		c, 
		"Berhasil mendapatkan buku kas pengawas", 
		inspectorLedgerResponse, 
		&param.PaginationParam,
	)
}

func (r *rest) getAllInspectorLedger(
	ctx context.Context,
	param *model.PaginationParam,
	startTime int64,
) (model.InspectorLedgerResponse, error) {
	var inspectorLedgerResponse model.InspectorLedgerResponse

	param.SetDefaultPagination()
	rows, err := r.db.WithContext(ctx).
		Model(&model.InspectorLedger{}).
		InnerJoins("Inspector").
		Where("inspector_ledgers.created_at >= ?", startTime).
		Limit(int(param.Limit)).
		Offset(int(param.Offset)).
		Order("inspector_ledgers.created_at desc").
		Rows()
	if err != nil {
		return inspectorLedgerResponse, err
	}

	transactions := []model.InspectorLedgerTransaction{}

	defer rows.Close()
	for rows.Next() {
		var ledger model.InspectorLedger
		if err := r.db.WithContext(ctx).ScanRows(rows, &ledger); err != nil {
			return inspectorLedgerResponse, err
		}

		transaction := model.InspectorLedgerTransaction{
			Timestamp: ledger.CreatedAt,
			Type:      string(ledger.LedgerType),
			RefName:   ledger.Inspector.Name,
			Amount:    number.ConvertToRupiah(ledger.Amount),
		}

		transactions = append(transactions, transaction)
	}

	if err := r.db.WithContext(ctx).
		Model(&model.InspectorLedger{}).
		Where("created_at >= ?", startTime).
		Count(&param.TotalElement).Error; err != nil {
		return inspectorLedgerResponse, err
	}

	param.ProcessPagination(int64(len(transactions)))

	var finalBalance int64

	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			SUM(IL.final_balance) AS final_balance
		FROM inspector_ledgers IL
		INNER JOIN (
			SELECT
				inspector_id,
				MAX(created_at) AS created_at
			FROM inspector_ledgers
			GROUP BY 1
		) AS IL2 ON IL.inspector_id = IL2.inspector_id 
			AND IL.created_at = IL2.created_at
			AND IL.deleted_at IS NULL
	`).Scan(&finalBalance).Error; err != nil {
		return inspectorLedgerResponse, err
	}

	inspectorLedgerResponse = model.InspectorLedgerResponse{
		Account: model.InspectorLedgerAccount{
			InspectorName:  "Semua Pengawas",
			CurrentBalance: number.ConvertToRupiah(finalBalance),
		},
		Transactions: transactions,
	}

	return inspectorLedgerResponse, nil
}

func (r *rest) getSingleInspectorLedger(
	ctx context.Context,
	param *model.PaginationParam,
	startTime int64,
	inspectorID int64,
) (model.InspectorLedgerResponse, error) {
	var inspectorLedgerResponse model.InspectorLedgerResponse

	param.SetDefaultPagination()
	rows, err := r.db.WithContext(ctx).
		Model(&model.InspectorLedger{}).
		Where("created_at >= ? AND inspector_id = ?", startTime, inspectorID).
		Limit(int(param.Limit)).
		Offset(int(param.Offset)).
		Order("created_at desc").
		Rows()
	if err != nil {
		return inspectorLedgerResponse, err
	}

	transactions := []model.InspectorLedgerTransaction{}

	defer rows.Close()
	for rows.Next() {
		var ledger model.InspectorLedger
		if err := r.db.WithContext(ctx).ScanRows(rows, &ledger); err != nil {
			return inspectorLedgerResponse, err
		}

		transaction := model.InspectorLedgerTransaction{
			Timestamp: ledger.CreatedAt,
			Type:      string(ledger.LedgerType),
			RefName:   ledger.Ref,
			Amount:    number.ConvertToRupiah(ledger.Amount),
		}

		transactions = append(transactions, transaction)
	}

	if err := r.db.WithContext(ctx).
		Model(&model.InspectorLedger{}).
		Where("created_at >= ? AND inspector_id = ?", startTime, inspectorID).
		Count(&param.TotalElement).Error; err != nil {
		return inspectorLedgerResponse, err
	}

	param.ProcessPagination(int64(len(transactions)))

	var latestLedger model.InspectorLedger

	if err := r.db.WithContext(ctx).
		InnerJoins("Inspector").
		Where("inspector_id = ?", inspectorID).
		Order("created_at desc").
		Take(&latestLedger).Error; err != nil {
		return inspectorLedgerResponse, err
	}

	inspectorLedgerResponse = model.InspectorLedgerResponse{
		Account: model.InspectorLedgerAccount{
			InspectorName:  latestLedger.Inspector.Name,
			CurrentBalance: number.ConvertToRupiah(latestLedger.FinalBalance),
		},
		Transactions: transactions,
	}

	return inspectorLedgerResponse, nil
}
