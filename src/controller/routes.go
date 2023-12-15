package controller

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	swagger "tigaputera-backend/docs"
	"tigaputera-backend/sdk/jwt"
	"tigaputera-backend/sdk/log"
	"tigaputera-backend/sdk/password"
	"tigaputera-backend/sdk/storage"
	"tigaputera-backend/sdk/validator"
	"tigaputera-backend/src/database"
	"tigaputera-backend/src/model"

	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var once = sync.Once{}

type rest struct {
	http      *gin.Engine
	db        *database.DB
	log       log.LogInterface
	jwt       jwt.Interface
	password  password.Interface
	validator validator.Interface
	storage   storage.Interface
}

func Init(
	log log.LogInterface,
	db *database.DB,
	jwt jwt.Interface,
	password password.Interface,
	validator validator.Interface,
	storage storage.Interface,
) *rest {
	r := &rest{}

	// Initialize server with graceful shutdown
	once.Do(func() {
		gin.SetMode(gin.ReleaseMode)

		r.http = gin.New()
		r.log = log
		r.db = db
		r.jwt = jwt
		r.password = password
		r.validator = validator
		r.storage = storage

		r.RegisterMiddlewareAndRoutes()
	})

	return r
}

func (r *rest) RegisterMiddlewareAndRoutes() {
	// Global middleware
	r.http.Use(r.CorsMiddleware())
	r.http.Use(gin.Recovery())
	r.http.Use(r.SetTimeout)
	r.http.Use(r.AddFieldsToContext)

	r.setupSwagger()

	// Global routes
	r.http.GET("/ping", r.Ping)
	r.http.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Auth routes
	r.http.POST("/v1/auth/login", r.Login)

	// Protected Routes
	r.http.PUT("/v1/user/statistics/refresh", r.RefreshStatistics)
	v1 := r.http.Group("v1", r.Authorization())

	// User routes
	v1.Group("user")
	{
		v1.GET("user/profile", r.GetUserProfile)
		v1.PATCH("user/reset-password", r.ResetPassword)
		v1.POST(
			"user/inspector",
			r.AuthorizeRole(model.Admin),
			r.CreateInspector,
		)
		v1.GET(
			"user/inspector",
			r.AuthorizeRole(model.Admin),
			r.GetListInspector,
		)
		v1.DELETE(
			"user/inspector/:user_id",
			r.AuthorizeRole(model.Admin),
			r.DeactiveInspector,
		)
		v1.POST(
			"user/inspector/income",
			r.AuthorizeRole(model.Inspector),
			r.CreateInspectorIncome,
		)
		v1.GET(
			"user/inspector/ledger",
			r.GetInspectorLedger,
		)
		v1.GET("user/statistics", r.GetUserStats)
		v1.GET("user/statistics/detail", r.GetUserStatsDetail)
	}

	// Project routes
	v1.Group("project")
	{
		v1.POST("project", r.AuthorizeRole(model.Admin), r.CreateProject)
		v1.GET("project", r.GetListProject)
		v1.GET("project/:project_id", r.GetProject)
		v1.GET("project/:project_id/detail", r.GetProjectDetail)
		v1.PATCH(
			"project/:project_id/budget",
			r.AuthorizeRole(model.Admin),
			r.UpdateProjectBudget,
		)
		v1.PATCH(
			"project/:project_id/status",
			r.AuthorizeRole(model.Admin),
			r.UpdateProjectStatus,
		)
		v1.POST(
			"project/:project_id/expenditure",
			r.AuthorizeRole(model.Admin),
			r.CreateProjectExpenditure,
		)
		v1.POST(
			"project/:project_id/expenditure/:expenditure_id/detail",
			r.CreateProjectExpenditureDetail,
		)
		v1.GET(
			"project/:project_id/expenditure/:expenditure_id/detail",
			r.GetProjectExpenditureDetailList,
		)
		v1.DELETE(
			"project/:project_id/expenditure/:expenditure_id/detail/:expenditure_detail_id",
			r.DeleteProjectExpenditureDetail,
		)
	}
}

func (r *rest) setupSwagger() {
	swagger.SwaggerInfo.Host = os.Getenv("APP_HOST")
	swagger.SwaggerInfo.Schemes = []string{"https", "http"}
}

func (r *rest) Run() {
	/*
		Create context that listens for the interrupt signal from the OS.
		This will allow us to gracefully shutdown the server.
	*/
	c := context.Background()
	ctx, stop := signal.NotifyContext(c, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	port := ":8080"
	if os.Getenv("APP_PORT") != "" {
		port = ":" + os.Getenv("APP_PORT")
	}
	server := &http.Server{
		Addr:              port,
		Handler:           r.http,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// Run the server in a goroutine so that it doesn't block the graceful shutdown handling below

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			r.log.Error(ctx, err.Error())
		}
	}()

	r.log.Info(context.Background(), "Server is running on port "+os.Getenv("APP_PORT"))

	// Block until we receive our signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	r.log.Info(context.Background(), "Shutting down server...")

	// Create a deadline to wait for.
	quitCtx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()
	if err := server.Shutdown(quitCtx); err != nil {
		r.log.Fatal(quitCtx, fmt.Sprintf("Server Shutdown error: %s", err.Error()))
	}

	r.log.Info(context.Background(), "Server gracefully stopped")
}

// @Summary Health Check
// @Description Check if the server is running
// @Tags Server
// @Produce json
// @Success 200 string example="PONG!!"
// @Router /ping [GET]
func (r *rest) Ping(c *gin.Context) {
	r.SuccessResponse(c, "PONG!!", nil, nil)
}
