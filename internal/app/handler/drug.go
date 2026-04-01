package handler

import (
	"net/http"
	"strconv"
	"web_backend/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetDrugs(ctx *gin.Context) {
	var drugs []ds.Drug
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		drugs, err = h.Repository.GetDrugs()
	} else {
		drugs, err = h.Repository.GetDrugsByTitle(searchQuery)
	}
	if err != nil {
		logrus.Error(err)
	}

	creatorID := uint(1)
	prescriptionCount := h.Repository.GetPrescriptionDrugCount(creatorID)
	activePrescriptionID := h.Repository.GetActivePrescriptionID(creatorID)

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"drugs":              drugs,
		"query":              searchQuery,
		"prescription_count": prescriptionCount,
		"prescription_id":    activePrescriptionID,
		"minioUrl":           h.Config.MinioURL,
	})
}

func (h *Handler) GetDrug(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	drug, err := h.Repository.GetDrug(id)
	if err != nil {
		logrus.Error(err)
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	ctx.HTML(http.StatusOK, "drug.html", gin.H{
		"drug":     drug,
		"minioUrl": h.Config.MinioURL,
	})
}
