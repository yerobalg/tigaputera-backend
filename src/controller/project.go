package controller

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"tigaputera-backend/sdk/auth"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/sdk/number"
	"tigaputera-backend/src/model"

	"context"
)

// @Summary Create Project
// @Description Create new project
// @Tags Project
// @Produce json
// @Security BearerAuth
// @Param createProjectBody body model.CreateProjectBody true "body"
// @Success 201 {object} model.HTTPResponse{}
// @Failure 400 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project [POST]
func (r *rest) CreateProject(c *gin.Context) {
	ctx := c.Request.Context()
	var body model.CreateProjectBody

	if err := r.BindBody(c, &body); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	if err := r.validator.ValidateStruct(body); err != nil {
		r.ErrorResponse(c, errors.BadRequest(err.Error()))
		return
	}

	if !model.IsProjectTypeCorrect(body.Type) {
		r.ErrorResponse(c, errors.BadRequest("Tipe proyek harus Drainase, Beton, Hotmix, atau Bangunan"))
		return
	}

	project := model.Project{
		Name:        body.Name,
		Description: body.Description,
		Type:        body.Type,
		DeptName:    body.DeptName,
		CompanyName: body.CompanyName,
		Status:      string(model.Running),
		Volume:      body.Volume,
		Length:      body.Length,
		Width:       body.Width,
		InspectorID: body.InspectorID,
		StartDate:   body.StartDate,
		FinalDate:   body.FinalDate,
	}

	tx := r.db.WithContext(ctx).Begin()

	err := tx.Create(&project).Error
	if r.isUniqueKeyViolation(err) {
		tx.Rollback()
		r.ErrorResponse(c, errors.BadRequest("Nama proyek sudah ada"))
		return
	} else if err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	initialProjectExpenditures := make([]model.ProjectExpenditure, len(model.InitialProjectExpenditures))
	copy(initialProjectExpenditures, model.InitialProjectExpenditures)
	for i := range initialProjectExpenditures {
		initialProjectExpenditures[i].ProjectID = project.ID
	}

	if err := tx.Create(&initialProjectExpenditures).Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	r.CreatedResponse(c, "Berhasil membuat proyek", nil)
}

// @Summary Get list project
// @Description Get list project
// @Tags Project
// @Produce json
// @Security BearerAuth
// @param limit query int false "limit"
// @param page query int false "page"
// @param keyword query string false "keyword"
// @Success 200 {object} model.HTTPResponse{data=[]model.ProjectListResponse}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project [GET]
func (r *rest) GetListProject(c *gin.Context) {
	ctx := c.Request.Context()
	var param model.ProjectParam

	if err := r.BindParam(c, &param); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	param.SetDefaultPagination()

	user := auth.GetUser(ctx)

	var rows *sql.Rows
	var err error

	if user.Role == string(model.Admin) {
		rows, err = r.db.WithContext(ctx).
			Where("projects.name ILIKE ?", "%"+param.Keyword+"%").
			Model(&model.Project{}).
			InnerJoins("Inspector").
			Limit(int(param.Limit)).
			Offset(int(param.Offset)).
			Order("projects.updated_at DESC").
			Rows()
	} else {
		rows, err = r.db.WithContext(ctx).
			Where("projects.name ILIKE ?", "%"+param.Keyword+"%").
			Where("projects.inspector_id = ?", user.ID).
			Model(&model.Project{}).
			InnerJoins("Inspector").
			Limit(int(param.Limit)).
			Offset(int(param.Offset)).
			Order("projects.updated_at DESC").
			Rows()
	}

	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	defer rows.Close()
	projectListResponses := []model.ProjectListResponse{}
	for rows.Next() {
		var project model.Project
		if err := r.db.ScanRows(rows, &project); err != nil {
			r.ErrorResponse(c, errors.InternalServerError(err.Error()))
			return
		}

		projectListResponse := model.ProjectListResponse{
			ID:            project.ID,
			Name:          project.Name,
			Type:          model.GetProjectTypeStyle(project.Type),
			Status:        model.GetProjectStatusStyle(project.Status),
			UpdatedAt:     project.UpdatedAt,
			InspectorName: project.Inspector.Name,
		}
		projectListResponses = append(projectListResponses, projectListResponse)
	}

	if user.Role == string(model.Admin) {
		err = r.db.WithContext(ctx).
			Where("name ILIKE ?", "%"+param.Keyword+"%").
			Model(&model.Project{}).
			Count(&param.TotalElement).Error
	} else {
		err = r.db.WithContext(ctx).
			Where("name ILIKE ?", "%"+param.Keyword+"%").
			Where("inspector_id = ?", user.ID).
			Model(&model.Project{}).
			Count(&param.TotalElement).Error
	}

	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	param.ProcessPagination(int64(len(projectListResponses)))

	r.SuccessResponse(c, "Berhasil mendapatkan list proyek", projectListResponses, &param.PaginationParam)
}

// @Summary Get list project name
// @Description Get list project name
// @Tags Project
// @Produce json
// @Security BearerAuth
// @param limit query int false "limit"
// @param page query int false "page"
// @param keyword query string false "keyword"
// @Success 200 {object} model.HTTPResponse{data=[]model.ProjectNameResponse}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project/name [GET]
func (r *rest) GetListProjectName(c *gin.Context) {
	ctx := c.Request.Context()

	user := auth.GetUser(ctx)

	queryStmt := "1 = ?"
	queryArgs := []interface{}{1}
	if user.Role == string(model.Inspector) {
		queryStmt += " AND inspector_id = ?"
		queryArgs = append(queryArgs, user.ID)
	}

	project := []model.ProjectNameResponse{}
	if err := r.db.WithContext(ctx).
		Model(&model.Project{}).
		Where(queryStmt, queryArgs...).
		Select("id, name").
		Find(&project).Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	r.SuccessResponse(c, "Berhasil mendapatkan list nama proyek", project, nil)
}

// @Summary Get project
// @Description Get a project
// @Tags Project
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "project_id"
// @Success 200 {object} model.HTTPResponse{data=model.Project}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project/{project_id} [GET]
func (r *rest) GetProject(c *gin.Context) {
	ctx := c.Request.Context()
	var param model.ProjectParam

	if err := r.BindParam(c, &param); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	var project model.Project
	if err := r.db.WithContext(ctx).
		Where(&param).
		First(&project).
		Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	r.SuccessResponse(c, "Berhasil mendapatkan proyek", project, nil)
}

// @Summary Get project detail
// @Description Get a project with its budget
// @Tags Project
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "project_id"
// @Success 200 {object} model.HTTPResponse{data=model.ProjectDetailResponse}
// @Failure 400 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project/{project_id}/detail [GET]
func (r *rest) GetProjectDetail(c *gin.Context) {
	ctx := c.Request.Context()
	var param model.ProjectParam

	if err := r.BindParam(c, &param); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	var project model.Project
	if err := r.db.WithContext(ctx).
		Where(&param).
		InnerJoins("Inspector").
		First(&project).
		Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	ppnPrice := -int64(float64(*project.Budget) * project.PPN)
	pphPrice := -int64(float64(*project.Budget) * project.PPH)
	totalBudget := *project.Budget + ppnPrice + pphPrice

	projectBudget := model.ProjectBudget{
		Budgets: []model.Budget{
			{
				Name:  "Pagu Pekerjaan",
				Price: number.ConvertToRupiah(*project.Budget),
			},
			{
				Name:  "PPN",
				Price: number.ConvertToRupiah(ppnPrice),
			},
			{
				Name:  "PPh",
				Price: number.ConvertToRupiah(pphPrice),
			},
		},
		PPNPercentage: project.PPN * 100,
		PPHPercentage: project.PPH * 100,
		Total:         number.ConvertToRupiah(totalBudget),
	}

	projectExpenditure, totalExpenditure, err := r.getProjectExpenditure(
		ctx,
		param.ID,
	)
	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	projectStats := r.GetProjectDetailStats(ctx, project, totalBudget)
	margin := totalBudget - totalExpenditure
	projectDetailResponse := r.getProjectDetailRes(
		project,
		projectBudget,
		projectExpenditure,
		projectStats,
		margin,
	)

	r.SuccessResponse(c, "Berhasil mendapatkan proyek", projectDetailResponse, nil)
}

func (r *rest) getProjectExpenditure(
	ctx context.Context,
	projectID int64,
) (model.ProjectExpenditureResponse, int64, error) {
	var res model.ProjectExpenditureResponse
	var totalExpenditure int64

	expenditureParam := model.ProjectExpenditureParam{
		ProjectID: projectID,
	}

	rows, err := r.db.WithContext(ctx).
		Model(&model.ProjectExpenditure{}).
		Where(&expenditureParam).
		Order("sequence").
		Rows()
	if err != nil {
		return res, totalExpenditure, err
	}

	defer rows.Close()
	expenditures := []model.ProjectExpenditureList{}

	for rows.Next() {
		var expenditure model.ProjectExpenditure
		if err := r.db.ScanRows(rows, &expenditure); err != nil {
			return res, totalExpenditure, err
		}

		expenditures = append(expenditures, model.ProjectExpenditureList{
			ID:          expenditure.ID,
			Sequence:    expenditure.Sequence,
			Name:        expenditure.Name,
			TotalPrice:  number.ConvertToRupiah(*expenditure.TotalPrice),
			IsFixedCost: *expenditure.IsFixedCost,
		})

		totalExpenditure += *expenditure.TotalPrice
	}

	res.Expenditures = expenditures
	res.SumTotal = number.ConvertToRupiah(totalExpenditure)

	return res, totalExpenditure, nil
}

func (r *rest) GetProjectDetailStats(
	ctx context.Context,
	project model.Project,
	totalBudget int64,
) model.ProjectStatistics {
	projectStats := model.ProjectStatistics{
		TotalIncome:          number.ConvertToRupiah(*project.Income),
		PercentageFromBudget: number.GetPercentage(*project.Income, totalBudget),
	}

	return projectStats
}

func (r *rest) getProjectDetailRes(
	project model.Project,
	projectBudget model.ProjectBudget,
	projectExpenditure model.ProjectExpenditureResponse,
	projectStats model.ProjectStatistics,
	margin int64,
) model.ProjectDetailResponse {
	projectDetailResponse := model.ProjectDetailResponse{
		ID:                 project.ID,
		Name:               project.Name,
		Description:        project.Description,
		Type:               model.GetProjectTypeStyle(project.Type),
		Status:             model.GetProjectStatusStyle(project.Status),
		DeptName:           project.DeptName,
		CompanyName:        project.CompanyName,
		Volume:             project.Volume,
		Length:             project.Length,
		Width:              project.Width,
		InspectorName:      project.Inspector.Name,
		StartDate:          project.StartDate,
		FinalDate:          project.FinalDate,
		ProjectBudget:      projectBudget,
		ProjectExpenditure: projectExpenditure,
		ProjectStatistics:  projectStats,
		Margin:             number.ConvertToRupiah(margin),
	}

	return projectDetailResponse
}

// @Summary Update Project Budget
// @Description Update project budget
// @Tags Project
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "project_id"
// @Param updateProjectBudgetBody body model.UpdateProjectBudgetBody true "body"
// @Success 200 {object} model.HTTPResponse{}
// @Failure 400 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 404 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project/{project_id}/budget [PATCH]
func (r *rest) UpdateProjectBudget(c *gin.Context) {
	ctx := c.Request.Context()
	var param model.ProjectParam
	var body model.UpdateProjectBudgetBody

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

	updatedProject := model.Project{
		Budget: &body.Budget,
		PPN:    body.PPN,
		PPH:    body.PPH,
	}

	res := r.db.WithContext(ctx).
		Model(&model.Project{}).
		Where(&param).
		Updates(&updatedProject)
	if res.RowsAffected == 0 {
		r.ErrorResponse(c, errors.NotFound("Proyek tidak ditemukan"))
		return
	} else if res.Error != nil {
		r.ErrorResponse(c, errors.InternalServerError(res.Error.Error()))
		return
	}

	r.SuccessResponse(c, "Berhasil mengubah anggaran proyek", nil, nil)
}

// @Summary Update Project Status
// @Description Update project Status
// @Tags Project
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "project_id"
// @Param updateProjectStatusBody body model.UpdateProjectStatusBody true "body"
// @Success 200 {object} model.HTTPResponse{}
// @Failure 400 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 404 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/project/{project_id}/status [PATCH]
func (r *rest) UpdateProjectStatus(c *gin.Context) {
	ctx := c.Request.Context()
	var param model.ProjectParam
	var body model.UpdateProjectStatusBody

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

	if !model.IsProjectStatusCorrect(body.Status) {
		r.ErrorResponse(
			c,
			errors.BadRequest("Status proyek harus Sedang Berjalan, Selesai, Ditunda, atau Dibatalkan"),
		)
		return
	}

	updatedProject := model.Project{
		Status: body.Status,
	}

	res := r.db.WithContext(ctx).
		Model(&model.Project{}).
		Where(&param).
		Updates(&updatedProject)
	if res.RowsAffected == 0 {
		r.ErrorResponse(c, errors.NotFound("Proyek tidak ditemukan"))
		return
	} else if res.Error != nil {
		r.ErrorResponse(c, errors.InternalServerError(res.Error.Error()))
		return
	}

	r.SuccessResponse(c, "Berhasil mengubah status proyek", nil, nil)
}
