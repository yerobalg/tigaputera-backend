package controller

import (
	"github.com/gin-gonic/gin"
	"tigaputera-backend/sdk/auth"
	"tigaputera-backend/sdk/error"
	"tigaputera-backend/src/model"
)

// @Summary Login
// @Description Login for user
// @Tags User
// @Produce json
// @Param loginBody body model.UserLoginBody true "body"
// @Success 200 {object} model.HTTPResponse{data=model.UserLoginResponse}
// @Failure 400 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/auth/login [POST]
func (r *rest) Login(c *gin.Context) {
	ctx := c.Request.Context()
	var loginBody model.UserLoginBody

	if err := r.BindBody(c, &loginBody); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	if err := r.validator.ValidateStruct(loginBody); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	userParam := model.UserParam{Username: loginBody.Username}
	var user model.User

	if err := r.db.WithContext(ctx).
		Where(&userParam).
		First(&user).Error; err != nil {
		r.ErrorResponse(c, errors.BadRequest("User not found"))
		return
	}

	if !r.password.Compare(user.Password, loginBody.Password) {
		r.ErrorResponse(c, errors.BadRequest("Wrong password"))
		return
	}

	token, err := r.jwt.GenerateToken(user)
	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError("Failed to generate token"))
		return
	}

	userResponse := model.UserLoginResponse{
		User:  user,
		Token: token,
	}

	r.SuccessResponse(c, "Login successfull", userResponse, nil)
}

// @Summary Get user profile
// @Description Get user profile
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.HTTPResponse{data=model.User}
// @Failure 401 {object} model.HTTPResponse{}
// @Router /v1/user/profile [GET]
func (r *rest) GetUserProfile(c *gin.Context) {
	ctx := c.Request.Context()
	user := auth.GetUser(ctx)

	userResponse := model.User{
		ID:       user.ID,
		Username: user.Username,
		Name:     user.Name,
		Role:     model.Role(user.Role),
	}

	r.SuccessResponse(c, "Get user profile success", userResponse, nil)
}

// @Summary Reset password
// @Description Reset password
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param resetPasswordBody body model.ResetPasswordBody true "body"
// @Success 200 {object} model.HTTPResponse{}
// @Failure 400 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/user/reset-password [PATCH]
func (r *rest) ResetPassword(c *gin.Context) {
	ctx := c.Request.Context()
	userInfo := auth.GetUser(ctx)
	userParam := model.UserParam{Username: userInfo.Username}
	var user model.User

	if err := r.db.WithContext(ctx).
		Where(&userParam).
		First(&user).Error; err != nil {
		r.ErrorResponse(c, errors.BadRequest("User not found"))
		return
	}

	if !user.IsFirstLogin {
		r.ErrorResponse(c, errors.BadRequest("You have already reset your password"))
		return
	}

	var resetPasswordBody model.ResetPasswordBody

	if err := r.BindBody(c, &resetPasswordBody); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	if err := r.validator.ValidateStruct(resetPasswordBody); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	newPassword, err := r.password.Hash(resetPasswordBody.NewPassword)
	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError("Failed to hash password"))
		return
	}

	updatedUser := model.User{
		Password:     newPassword,
		IsFirstLogin: false,
	}

	if err := r.db.WithContext(ctx).
		Model(model.User{}).
		Where(&userParam).
		Updates(&updatedUser).Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError("Failed to reset password"))
		return
	}

	r.SuccessResponse(c, "Reset password success", nil, nil)
}
