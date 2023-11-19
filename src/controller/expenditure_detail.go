package controller

import (
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
// @Param body body CreateExpenditureDetailBody true "body"
// @Success 200 {object} model.HTTPResponse{}
// @Failure 400 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
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
	if err != nil && r.isNoRecordFound(err) {
		r.ErrorResponse(c, errors.BadRequest("pengeluaran proyek tidak ditemukan"))
		return
	} else if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	expenditureDetail := model.ExpenditureDetail{
		Name:          body.Name,
		Price:         body.Price,
		Amount:        body.Amount,
		TotalPrice:    body.Price * body.Amount,
		ReceiptURL:    "", // TODO: Upload receipt
		ExpenditureID: projectExpenditure.ID,
	}

	projectExpenditure.TotalPrice += expenditureDetail.TotalPrice
	projectExpenditure.UpdatedBy = &user.ID

	tx := r.db.WithContext(ctx).Begin()

	if err := tx.Create(&expenditureDetail).Error; err != nil {
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

	r.SuccessResponse(c, "Berhasil membuat detail pengeluaran proyek", nil, nil)
}
