package controller

import (
	"context"
	"time"

	"tigaputera-backend/sdk/appcontext"
	"tigaputera-backend/sdk/auth"
	"tigaputera-backend/sdk/error"
	"tigaputera-backend/src/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// timeout middleware wraps the request context with a timeout
func (r *rest) SetTimeout(ctx *gin.Context) {
	// wrap the request context with a timeout, this will cause the request to fail if it takes more than defined timeout
	c, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Minute) // TODO: change this hardcoded timeout to config later

	// cancel to clear resources after finished
	defer cancel()
	c = appcontext.SetRequestStartTime(c, time.Now())

	// replace request with context wrapped request
	ctx.Request = ctx.Request.WithContext(c)
	ctx.Next()
}

func (r *rest) AddFieldsToContext(ctx *gin.Context) {
	requestID := uuid.New().String()

	c := ctx.Request.Context()
	c = appcontext.SetRequestId(c, requestID)
	c = appcontext.SetUserAgent(c, ctx.Request.Header.Get(appcontext.HeaderUserAgent))
	c = appcontext.SetDeviceType(c, ctx.Request.Header.Get(appcontext.HeaderDeviceType))
	ctx.Request = ctx.Request.WithContext(c)

	ctx.Next()
}

func (r *rest) CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (r *rest) Authorization() gin.HandlerFunc {
	return r.checkToken
}

func (r *rest) checkToken(ctx *gin.Context) {
	header := ctx.Request.Header.Get("Authorization")
	if header == "" {
		r.ErrorResponse(ctx, errors.Unauthorized("Harap login terlebih dahulu"))
		ctx.Abort()
		return
	}

	header = header[len("Bearer "):]
	tokenClaims, err := r.jwt.DecodeToken(header)
	if err != nil {
		r.ErrorResponse(ctx, errors.Unauthorized("Token tidak valid"))
		ctx.Abort()
		return
	}

	c := ctx.Request.Context()
	c = auth.SetUser(c, tokenClaims["data"].(map[string]interface{}))
	ctx.Request = ctx.Request.WithContext(c)

	ctx.Next()
}

func (r *rest) AuthorizeRole(roleName model.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userRole := auth.GetUser(ctx.Request.Context()).Role
		if userRole != string(roleName) {
			r.ErrorResponse(ctx, errors.Unauthorized("Anda tidak memiliki akses untuk melakukan aksi ini"))
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
