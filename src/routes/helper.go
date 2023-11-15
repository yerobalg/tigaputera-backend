package routes

import (
	"fmt"
	"time"

	"tigaputera-backend/sdk/appcontext"
	"tigaputera-backend/sdk/error"
	"tigaputera-backend/src/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func (r *rest) BindParam(ctx *gin.Context, param interface{}) error {
	if err := ctx.ShouldBindUri(param); err != nil {
		return err
	}

	return ctx.ShouldBindWith(param, binding.Query)
}

func (r *rest) BindBody(ctx *gin.Context, body interface{}) error {
	return ctx.ShouldBindWith(body, binding.Default(ctx.Request.Method, ctx.ContentType()))
}

func getRequestMetadata(ctx *gin.Context) model.Meta {
	meta := model.Meta{
		RequestID: appcontext.GetRequestId(ctx.Request.Context()),
		Time:      time.Now().Format(time.RFC3339),
	}

	requestStartTime := appcontext.GetRequestStartTime(ctx.Request.Context())
	if !requestStartTime.IsZero() {
		elapsedTimeMs := time.Since(requestStartTime).Milliseconds()
		meta.TimeElapsed = fmt.Sprintf("%dms", elapsedTimeMs)
	}

	return meta
}

func SuccessResponse(ctx *gin.Context, message string, data interface{}, pg *model.PaginationParam) {
	ctx.JSON(200, model.HTTPResponse{
		Meta:       getRequestMetadata(ctx),
		Message:    message,
		IsSuccess:  true,
		Data:       data,
		Pagination: pg,
	})
}

func ErrorResponse(ctx *gin.Context, err error) {
	ctx.JSON(int(errors.GetCode(err)), model.HTTPResponse{
		Meta:      getRequestMetadata(ctx),
		Message:   errors.GetType(err),
		IsSuccess: false,
		Data:      errors.GetMessage(err),
	})
}