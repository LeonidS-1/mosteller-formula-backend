package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"web_backend/internal/app/config"
	"web_backend/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
	Config     *config.Config
}

func NewHandler(r *repository.Repository, cfg *config.Config) *Handler {
	return &Handler{
		Repository: r,
		Config:     cfg,
	}
}

// RegisterHandler godoc
// @title Prescription API
// @version 1.0
// @description API для управления рецептами, препаратами и расчетом доз.
// @contact.name API Support
// @contact.url http://localhost:8080
// @contact.email support@example.com
// @license.name MIT
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.Use(CORSMiddleware())

	api := router.Group("/api")

	public := api.Group("/")
	public.POST("/users/register", h.CreateUser)
	public.POST("/users/login", h.SignIn)
	public.GET("/drugs", h.GetDrugs)
	public.GET("/drugs/:id", h.GetDrug)

	optionalAuth := api.Group("/")
	optionalAuth.Use(h.WithOptionalAuthCheck())
	optionalAuth.GET("/prescriptions/cart", h.GetPrescriptionCart)

	authorized := api.Group("/")
	authorized.Use(h.AuthMiddleware(false))
	authorized.POST("/drugs", h.CreateDrug)
	authorized.GET("/prescriptions", h.GetAllPrescriptions)
	authorized.GET("/prescriptions/:id", h.GetPrescription)
	authorized.PUT("/prescriptions/:id", h.EditPrescription)
	authorized.PUT("/prescriptions/:id/form", h.FormPrescription)
	authorized.DELETE("/prescriptions/:id", h.DeletePrescription)
	authorized.POST("/prescription_drugs/add/:drug_id", h.AddDrugToPrescription)
	authorized.DELETE("/prescription_drugs/:drug_id/:prescription_id", h.DeleteDrugFromPrescription)
	authorized.PUT("/prescription_drugs/:drug_id/:prescription_id", h.EditDrugInPrescription)
	authorized.POST("/users/logout", h.SignOut)

	moderator := api.Group("/")
	moderator.Use(h.AuthMiddleware(true))
	moderator.PUT("/prescriptions/:id/finish", h.FinishPrescription)

	swaggerURL := ginSwagger.URL("/swagger/doc.json")
	router.Any("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerURL))
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.Static("/static", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	var errorMessage string
	switch {
	case errors.Is(err, repository.ErrNotFound):
		errorMessage = "Не найден"
	case errors.Is(err, repository.ErrAlreadyExists):
		errorMessage = "Уже существует"
	case errors.Is(err, repository.ErrNotAllowed):
		errorMessage = "Доступ запрещен"
	case errors.Is(err, repository.ErrNoDraft):
		errorMessage = "Черновик не найден"
	default:
		errorMessage = err.Error()
	}
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": errorMessage,
	})
}
