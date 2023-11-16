package controller

import (
	"github.com/gin-gonic/gin"
	"tigaputera-backend/sdk/auth"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/src/model"
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

	user := auth.GetUser(ctx)

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
		InspectorID: user.ID,
	}

	if err := r.db.WithContext(ctx).Create(&project).Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	r.CreatedResponse(c, "Berhasil membuat proyek", nil)
}
