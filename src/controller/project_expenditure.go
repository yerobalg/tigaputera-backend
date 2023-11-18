package controller

import (
	"github.com/gin-gonic/gin"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/src/model"
)

// @Summary Get List Project Expenditure
// @Description Get list project expenditure
// @Tags Project
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "project_id"
// @Param page query int false "page"
// @Param limit query int false "limit"
// @Success 200 {object} model.HTTPResponse{data=[]model.ProjectExpenditureListResponse}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project/{project_id}/expenditure [GET]
func (r *rest) GetProjectExpenditureList(c *gin.Context) {
	ctx := c.Request.Context()
	var param model.ProjectExpenditureParam

	if err := r.BindParam(c, &param); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	param.PaginationParam.SetDefaultPagination()
	projectExpenditures := []model.ProjectExpenditureListResponse{}

	if err := r.db.WithContext(ctx).
		Model(&model.ProjectExpenditure{}).
		Where(&param).
		Limit(int(param.Limit)).
		Offset(int(param.Offset)).
		Order("sequence").
		Find(&projectExpenditures).Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
	}

	if err := r.db.WithContext(ctx).
		Model(&model.ProjectExpenditure{}).
		Where(&param).
		Count(&param.TotalElement).
		Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	param.ProcessPagination(int64(len(projectExpenditures)))

	r.SuccessResponse(c, "Berhasil mendapatkan list pengeluaran proyek", projectExpenditures, &param.PaginationParam)
}
