package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api")

	drugs := api.Group("/drugs")
	{
		drugs.GET("", h.GetDrugs)
		drugs.GET("/:id", h.GetDrug)
		drugs.POST("", h.CreateDrug)
	}

	prescriptions := api.Group("/prescriptions")
	{
		prescriptions.GET("/cart", h.GetPrescriptionCart)
		prescriptions.GET("", h.GetAllPrescriptions)
		prescriptions.GET("/:id", h.GetPrescription)
		prescriptions.PUT("/:id", h.EditPrescription)
		prescriptions.PUT("/:id/form", h.FormPrescription)
		prescriptions.PUT("/:id/finish", h.FinishPrescription)
		prescriptions.DELETE("/:id", h.DeletePrescription)
	}

	pd := api.Group("/prescription_drugs")
	{
		pd.POST("/add/:drug_id", h.AddDrugToPrescription)
		pd.DELETE("/:drug_id/:prescription_id", h.DeleteDrugFromPrescription)
		pd.PUT("/:drug_id/:prescription_id", h.EditDrugInPrescription)
	}

	users := api.Group("/users")
	{
		users.POST("/register", h.CreateUser)
		users.POST("/login", h.SignIn)
		users.POST("/logout", h.SignOut)
	}
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
