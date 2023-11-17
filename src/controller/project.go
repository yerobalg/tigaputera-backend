package controller

import (
	"database/sql"
	"tigaputera-backend/sdk/auth"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/src/model"

	"github.com/gin-gonic/gin"
)

// @Summary Create Project
// @Description Create new project
// @Tags Project
// @Produce json
// @Security BearerAuth
// @Param createProjectBody body model.CreateProjectBody true "body"
// @Success 201 {object} model.HTTPResponse{}
// @Failure 400 {object} model.HTTPResponse{}
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
		r.ErrorResponse(c, err)
		return
	}

	if !model.ValidateProjectType(body.Type) {
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
	}

	if err := r.db.WithContext(ctx).Create(&project).Error; err != nil {
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
			Type:          project.Type,
			Status:        project.Status,
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
