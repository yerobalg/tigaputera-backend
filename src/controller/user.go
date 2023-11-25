package controller

import (
	"github.com/gin-gonic/gin"
	"tigaputera-backend/sdk/auth"
	errors "tigaputera-backend/sdk/error"
	"tigaputera-backend/sdk/number"
	"tigaputera-backend/src/model"
	"time"
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
		r.ErrorResponse(c, errors.BadRequest(err.Error()))
		return
	}

	userParam := model.UserParam{Username: loginBody.Username}
	var user model.User

	if err := r.db.WithContext(ctx).
		Where(&userParam).
		First(&user).Error; err != nil {
		r.ErrorResponse(c, errors.BadRequest("Pengguna tidak ditemukan"))
		return
	}

	if !r.password.Compare(user.Password, loginBody.Password) {
		r.ErrorResponse(c, errors.BadRequest("Password anda salah"))
		return
	}

	token, err := r.jwt.GenerateToken(user)
	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	userResponse := model.UserLoginResponse{
		User:  user,
		Token: token,
	}

	r.SuccessResponse(c, "Login berhasil", userResponse, nil)
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

	r.SuccessResponse(c, "Berhasil menampilkan profil", userResponse, nil)
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
		r.ErrorResponse(c, errors.BadRequest("Pengguna tidak ditemukan"))
		return
	}

	if !*user.IsFirstLogin {
		r.ErrorResponse(c, errors.BadRequest("Anda sudah pernah mengganti password"))
		return
	}

	var resetPasswordBody model.ResetPasswordBody

	if err := r.BindBody(c, &resetPasswordBody); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	if err := r.validator.ValidateStruct(resetPasswordBody); err != nil {
		r.ErrorResponse(c, errors.BadRequest(err.Error()))
		return
	}

	newPassword, err := r.password.Hash(resetPasswordBody.NewPassword)
	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	updatedUser := model.User{
		Password:     newPassword,
		IsFirstLogin: new(bool), // false
	}

	if err := r.db.WithContext(ctx).
		Model(model.User{}).
		Where(&userParam).
		Updates(&updatedUser).Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	r.SuccessResponse(c, "Reset password berhasil!", nil, nil)
}

// @Summary Create inspector
// @Description Create new inspector
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param createInspectorBody body model.CreateInspectorBody true "body"
// @Success 201 {object} model.HTTPResponse{}
// @Failure 400 {object} model.HTTPResponse{}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/user/inspector [POST]
func (r *rest) CreateInspector(c *gin.Context) {
	ctx := c.Request.Context()
	var createInspectorBody model.CreateInspectorBody

	if err := r.BindBody(c, &createInspectorBody); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	if err := r.validator.ValidateStruct(createInspectorBody); err != nil {
		r.ErrorResponse(c, errors.BadRequest(err.Error()))
		return
	}

	hashedPassword, err := r.password.Hash(createInspectorBody.Password)
	if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	}

	userInfo := auth.GetUser(ctx)
	newUser := model.User{
		Username:  createInspectorBody.Username,
		Name:      createInspectorBody.Name,
		Password:  hashedPassword,
		Role:      model.Inspector,
		CreatedBy: &userInfo.ID,
		UpdatedBy: &userInfo.ID,
	}

	insertError := r.db.WithContext(ctx).Create(&newUser).Error
	if insertError != nil && r.isUniqueKeyViolation(insertError) {
		r.ErrorResponse(c, errors.BadRequest("Username sudah digunakan"))
		return
	} else if insertError != nil {
		r.ErrorResponse(c, errors.InternalServerError(insertError.Error()))
		return
	}

	r.CreatedResponse(c, "Berhasil membuat pengawas", nil)
}

// @Summary Get list inspector
// @Description Get list inspector
// @Tags User
// @Produce json
// @Security BearerAuth
// @param limit query int false "limit"
// @param page query int false "page"
// @Success 200 {object} model.HTTPResponse{data=[]model.User}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/user/inspector [GET]
func (r *rest) GetListInspector(c *gin.Context) {
	ctx := c.Request.Context()
	var userParam model.UserParam
	if err := r.BindParam(c, &userParam); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	userParam.Role = string(model.Inspector)
	userParam.PaginationParam.SetDefaultPagination()

	var users []model.User

	if err := r.db.WithContext(ctx).
		Where(&userParam).
		Offset(int(userParam.Offset)).
		Limit(int(userParam.Limit)).
		Find(&users).Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
	}

	if err := r.db.WithContext(ctx).
		Model(model.User{}).
		Where(&userParam).
		Count(&userParam.TotalElement).Error; err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
	}

	userParam.PaginationParam.ProcessPagination(int64(len(users)))

	r.SuccessResponse(c, "Berhasil mendapatkan list pengawas", users, &userParam.PaginationParam)
}

// @Summary Create Inspector Income
// @Description Create inspector income
// @Tags Inspector Income
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
		Order("created_at desc").
		First(&latestLedger, user.ID).Error

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

// @Summary Get User Stats
// @Description Get user statistics
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.HTTPResponse{data=model.InspectorStatsResponse{}}
// @Failure 401 {object} model.HTTPResponse{}
// @Failure 500 {object} model.HTTPResponse{}
// @Router /v1/user/stats [GET]
func (r *rest) GetUserStats(c *gin.Context) {
	ctx := c.Request.Context()
	user := auth.GetUser(ctx)

	var userStatsParam model.InspectorStatsParam
	if user.Role == string(model.Inspector) {
		userStatsParam.UserID = user.ID
	} else {
		userStatsParam.UserID = 0
	}

	lastMonth := time.Now().UTC().AddDate(0, -1, 0)
	userStatsParam.StartTime = time.Date(
		lastMonth.Year(),
		lastMonth.Month(),
		lastMonth.Day(),
		0,
		0,
		0,
		0,
		lastMonth.Location(),
	).Unix()

	var totalProject int64
	var totalExpenditure int64
	var totalIncome int64
	var totalMargin int64

	var userStats model.MqtInspectorStats
	err := r.db.WithContext(ctx).
		Where(&userStatsParam).
		First(&userStats).Error

	if r.isNoRecordFound(err) {
		totalProject = 0
		totalExpenditure = 0
		totalIncome = 0
		totalMargin = 0
	} else if err != nil {
		r.ErrorResponse(c, errors.InternalServerError(err.Error()))
		return
	} else {
		totalProject = model.GetTotalProject(userStats)
		totalExpenditure = model.GetTotalExpenditure(userStats)
		totalIncome = *userStats.TotalIncome
		totalMargin = totalIncome - totalExpenditure
	}

	userStatsResponse := model.InspectorStatsResponse{
		TotalProject:     totalProject,
		TotalExpenditure: number.ConvertToRupiah(totalExpenditure),
		TotalIncome:      number.ConvertToRupiah(totalIncome),
		Margin:           number.ConvertToRupiah(totalMargin),
	}

	r.SuccessResponse(c, "Berhasil mendapatkan statistik pengguna", userStatsResponse, nil)
}
