package controller

import (
	"github.com/gin-gonic/gin"
	"tigaputera-backend/sdk/auth"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/src/model"
)

// @Summary Get List Project Expenditure
// @Description Get list project expenditure
// @Tags Project Expenditure
// @Produce json
// @Security BearerAuth
// @Param project_id path  int true "project_id"
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

	r.SuccessResponse(
		c,
		"Berhasil mendapatkan list pengeluaran proyek",
		projectExpenditures,
		&param.PaginationParam,
	)
}

// @Summary Create Project Expenditure
// @Description Create new project expenditure
// @Tags Project Expenditure
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "project_id"
// @Param createProjectExpenditureBody body model.CreateProjectExpenditureBody true "body"
// @Success 200 {object} model.HTTPResponse{}
// @Failure 400 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project/{project_id}/expenditure [POST]
func (r *rest) CreateProjectExpenditure(c *gin.Context) {
	ctx := c.Request.Context()
	var param model.ProjectExpenditureParam
	var body model.CreateProjectExpenditureBody

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

	// Get latest sequence
	var latestProjectExpenditure model.ProjectExpenditure
	err := r.db.WithContext(ctx).
		Where("project_expenditures.project_id = ?", param.ProjectID).
		InnerJoins("Project", r.db.Where(&model.Project{InspectorID: user.ID})).
		Order("sequence desc").
		Take(&latestProjectExpenditure).Error

	if err != nil && r.isNoRecordFound(err) {
		r.ErrorResponse(c, errors.BadRequest("Proyek tidak ditemukan"))
		return
	} else if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
	}

	projectExpenditure := model.ProjectExpenditure{
		ProjectID:   param.ProjectID,
		Sequence:    latestProjectExpenditure.Sequence + 1,
		Name:        body.Name,
		IsFixedCost: body.IsFixedCost,
	}

	if err := r.db.WithContext(ctx).
		Model(&model.ProjectExpenditure{}).
		Create(&projectExpenditure).Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	r.SuccessResponse(c, "Berhasil membuat pengeluaran proyek", nil, nil)
}
