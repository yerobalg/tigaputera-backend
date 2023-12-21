package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"tigaputera-backend/sdk/auth"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/sdk/number"
	"tigaputera-backend/src/model"

	"database/sql"
	"time"
)

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

	var param model.LedgerParam
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

	startTime := r.getStartTime(intervalMonth)

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
	rows, err := r.getAllInspectorLedgerRows(ctx, param, startTime)
	if err != nil {
		return inspectorLedgerResponse, err
	}

	transactions := []model.InspectorLedgerTransaction{}

	defer rows.Close()
	for rows.Next() {
		var ledger model.Ledger
		if err := r.db.WithContext(ctx).ScanRows(rows, &ledger); err != nil {
			return inspectorLedgerResponse, err
		}

		transactions = append(transactions, r.getTransaction(ledger))
	}

	err = r.countAllInspectorLedger(ctx, startTime, &param.TotalElement)
	if err != nil {
		return inspectorLedgerResponse, err
	}

	param.ProcessPagination(int64(len(transactions)))

	finalBalance, err := r.getAllInspectorBalance(ctx)
	if err != nil {
		return inspectorLedgerResponse, err
	}

	inspectorLedgerResponse = r.getLedgerResponseAllInspector(finalBalance, transactions)

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
	rows, err := r.getSingleInspectorLedgerRows(ctx, param, startTime, inspectorID)
	if err != nil {
		return inspectorLedgerResponse, err
	}

	transactions := []model.InspectorLedgerTransaction{}

	defer rows.Close()
	for rows.Next() {
		var ledger model.Ledger
		if err := r.db.WithContext(ctx).ScanRows(rows, &ledger); err != nil {
			return inspectorLedgerResponse, err
		}

		transactions = append(transactions, r.getTransaction(ledger))
	}

	if err := r.countSingleInspectorLedger(ctx, startTime, inspectorID, &param.TotalElement); err != nil {
		return inspectorLedgerResponse, err
	}

	param.ProcessPagination(int64(len(transactions)))

	latestLedger, err := r.getInspectorLatestLedger(ctx, inspectorID)
	if err != nil {
		return inspectorLedgerResponse, err
	}

	inspectorLedgerResponse = r.getLedgerResponse(latestLedger, transactions)

	return inspectorLedgerResponse, nil
}

func (r *rest) getStartTime(intervalMonth int) int64 {
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

	return startTime
}

func (r *rest) getTransaction(ledger model.Ledger) model.InspectorLedgerTransaction {
	return model.InspectorLedgerTransaction{
		Timestamp:     ledger.CreatedAt,
		Type:          string(ledger.LedgerType),
		RefName:       ledger.Ref,
		Amount:        number.ConvertToRupiah(ledger.Amount),
		InspectorName: ledger.Inspector.Name,
		RecieptURL:    ledger.ReceiptURL,
	}
}

func (r *rest) getAllInspectorLedgerRows(
	ctx context.Context,
	param *model.PaginationParam,
	startTime int64,
) (*sql.Rows, error) {
	rows, err := r.db.WithContext(ctx).
		Model(&model.Ledger{}).
		InnerJoins("Inspector").
		Where("inspector_ledgers.created_at >= ?", startTime).
		Limit(int(param.Limit)).
		Offset(int(param.Offset)).
		Order("inspector_ledgers.created_at desc").
		Rows()
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *rest) countAllInspectorLedger(
	ctx context.Context,
	startTime int64,
	count *int64,
) error {
	err := r.db.WithContext(ctx).
		Model(&model.Ledger{}).
		Where("created_at >= ?", startTime).
		Count(count).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *rest) getSingleInspectorLedgerRows(
	ctx context.Context,
	param *model.PaginationParam,
	startTime int64,
	inspectorID int64,
) (*sql.Rows, error) {
	rows, err := r.db.WithContext(ctx).
		Model(&model.Ledger{}).
		InnerJoins("Inspector").
		Where("inspector_ledgers.created_at >= ? AND inspector_id = ?", startTime, inspectorID).
		Limit(int(param.Limit)).
		Offset(int(param.Offset)).
		Order("inspector_ledgers.created_at desc").
		Rows()
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *rest) countSingleInspectorLedger(
	ctx context.Context,
	startTime int64,
	inspectorID int64,
	count *int64,
) error {
	err := r.db.WithContext(ctx).
		Model(&model.Ledger{}).
		Where("created_at >= ? AND inspector_id = ?", startTime, inspectorID).
		Count(count).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *rest) getLedgerResponseAllInspector(
	finalBalance int64,
	transactions []model.InspectorLedgerTransaction,
) model.InspectorLedgerResponse {
	return model.InspectorLedgerResponse{
		Account: model.InspectorLedgerAccount{
			InspectorName:  "Semua Pengawas",
			CurrentBalance: number.ConvertToRupiah(finalBalance),
		},
		Transactions: transactions,
	}
}

func (r *rest) getLedgerResponse(
	latestLedger model.Ledger,
	transactions []model.InspectorLedgerTransaction,
) model.InspectorLedgerResponse {
	currentBalance := number.ConvertToRupiah(*latestLedger.FinalInspectorBalance)
	return model.InspectorLedgerResponse{
		Account: model.InspectorLedgerAccount{
			InspectorName:  latestLedger.Inspector.Name,
			CurrentBalance: currentBalance,
		},
		Transactions: transactions,
	}
}

func (r *rest) getAllInspectorBalance(
	ctx context.Context,
) (int64, error) {
	var finalBalance int64

	err := r.db.WithContext(ctx).
		Model(&model.MqtInspectorStats{}).
		Select("margin AS final_balance").
		Where("interval_month = 1 AND inspector_id = 0").
		Limit(1).
		Scan(&finalBalance).Error
	if err != nil {
		return finalBalance, err
	}

	return finalBalance, nil
}

func (r *rest) getInspectorLatestLedger(
	ctx context.Context,
	inspectorID int64,
) (model.Ledger, error) {
	var latestLedger model.Ledger

	err := r.db.WithContext(ctx).
		Where("inspector_id = ?", inspectorID).
		Order("created_at desc").
		Take(&latestLedger).Error
	if err != nil {
		return latestLedger, err
	}

	return latestLedger, nil
}
