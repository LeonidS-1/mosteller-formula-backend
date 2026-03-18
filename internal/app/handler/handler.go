package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"web_backend/internal/app/repository"
)

// Handler содержит зависимости HTTP-обработчиков.
type Handler struct {
	Repository *repository.Repository
}

// NewHandler создаёт новый Handler с переданным репозиторием.
func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// GetDrugs — обработчик главной страницы: список препаратов + иконка заявки.
func (h *Handler) GetDrugs(ctx *gin.Context) {
	var drugs []repository.Drug
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		drugs, err = h.Repository.GetDrugs()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		drugs, err = h.Repository.GetDrugsByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	prescriptions, err := h.Repository.GetPrescriptions()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"drugs":         drugs,
		"query":         searchQuery,
		"prescriptions": prescriptions,
	})
}

// GetDrug — обработчик страницы детальной информации о препарате.
func (h *Handler) GetDrug(ctx *gin.Context) {
	idStr := ctx.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	drug, err := h.Repository.GetDrug(id)
	if err != nil {
		logrus.Error(err)
	}

	prescriptionDrug, err := h.Repository.GetPrescriptionForDrug(id)
	hasDoseInfo := err == nil && prescriptionDrug != nil

	ctx.HTML(http.StatusOK, "drug.html", gin.H{
		"drug":             drug,
		"prescriptionDrug": prescriptionDrug,
		"hasDoseInfo":      hasDoseInfo,
	})
}

// GetPrescription — обработчик страницы заявки (расчёт детских доз).
func (h *Handler) GetPrescription(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	prescription, err := h.Repository.GetPrescription(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "prescription.html", gin.H{
		"prescription": prescription,
	})
}
