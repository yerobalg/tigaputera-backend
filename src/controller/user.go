package controller

import (
	"tigaputera-backend/src/model"
	"tigaputera-backend/sdk/error"
	"github.com/gin-gonic/gin"
)

// @Summary Login
// @Description Login for user
// @Tags User
// @Produce json
// @Param loginBody body model.UserLoginBody true "User login body"
// @Success 200 {object} model.HTTPResponse{data=model.UserLoginResponse}
// @Failure 400 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/auth/login [POST]
func (r *rest) Login (ctx *gin.Context) {
	var loginBody model.UserLoginBody

	if err := r.BindBody(ctx, &loginBody); err != nil {
		r.ErrorResponse(ctx, err)
		return
	}

	if err := r.validator.ValidateStruct(loginBody); err != nil {
		r.ErrorResponse(ctx, err)
		return
	}

	var user model.User

	if err := r.db.WithContext(ctx.Request.Context()).
		Where("username = ?", loginBody.Username).
		First(&user).Error; 
	err != nil {
		r.ErrorResponse(ctx, errors.BadRequest("User not found"))
		return
	}

	if !r.password.Compare(user.Password, loginBody.Password) {
		r.ErrorResponse(ctx, errors.BadRequest("Wrong password"))
		return
	}

	token, err := r.jwt.GenerateToken(user)
	if err != nil {
		r.ErrorResponse(ctx, errors.InternalServerError("Failed to generate token"))
		return
	}

	userResponse := model.UserLoginResponse{
		User: user,
		Token: token,
	}

	r.SuccessResponse(ctx, "Login successfull", userResponse, nil)
}
