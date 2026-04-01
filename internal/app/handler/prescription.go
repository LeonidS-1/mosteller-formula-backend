package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetPrescription(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	creatorID := uint(1)
	isDraft, err := h.Repository.IsDraftPrescription(id, creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if !isDraft {
		ctx.Redirect(http.StatusSeeOther, "/")
		return
	}

	items, prescription, err := h.Repository.GetPrescription(id, creatorID)
	if err != nil {
		logrus.Error(err)
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "prescription.html", gin.H{
		"items":           items,
		"prescription":    prescription,
		"prescription_id": id,
		"minioUrl":        h.Config.MinioURL,
	})
}

func (h *Handler) AddToPrescription(ctx *gin.Context) {
	drugIDStr := ctx.PostForm("drug_id")
	drugID, err := strconv.Atoi(drugIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	creatorID := uint(1)

	err = h.Repository.AddDrug(uint(drugID), creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, ctx.Request.Referer())
}

func (h *Handler) DeletePrescription(ctx *gin.Context) {
	prescriptionIDStr := ctx.PostForm("prescription_id")
	prescriptionID, err := strconv.Atoi(prescriptionIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.DeletePrescription(uint(prescriptionID))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/")
}
