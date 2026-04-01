package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"web_backend/internal/app/repository"
	"web_backend/internal/app/serializer"
)

func (h *Handler) AddDrugToPrescription(ctx *gin.Context) {
	drugIDStr := ctx.Param("drug_id")
	drugID, err := strconv.Atoi(drugIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	creatorID := uint(h.Repository.GetCreatorID())
	rx, created, err := h.Repository.GetPrescriptionDraft(creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	err = h.Repository.AddDrugToPrescription(uint(drugID), creatorID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrAlreadyExists) {
			h.errorHandler(ctx, http.StatusConflict, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}
	creatorLogin, moderatorLogin, _ := h.Repository.GetModeratorAndCreatorLogin(rx)
	completedCount, _ := h.Repository.GetCompletedDoseLineCount(rx.PrescriptionID)
	status := http.StatusOK
	if created {
		ctx.Header("Location", fmt.Sprintf("/api/prescriptions/%d", rx.PrescriptionID))
		status = http.StatusCreated
	}
	ctx.JSON(status, serializer.PrescriptionToJSON(rx, creatorLogin, moderatorLogin, completedCount))
}

func (h *Handler) DeleteDrugFromPrescription(ctx *gin.Context) {
	drugID, err := strconv.Atoi(ctx.Param("drug_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	prescriptionID, err := strconv.Atoi(ctx.Param("prescription_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	rx, err := h.Repository.DeleteDrugFromPrescription(prescriptionID, drugID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}
	creatorLogin, moderatorLogin, _ := h.Repository.GetModeratorAndCreatorLogin(rx)
	completedCount, _ := h.Repository.GetCompletedDoseLineCount(rx.PrescriptionID)
	ctx.JSON(http.StatusOK, serializer.PrescriptionToJSON(rx, creatorLogin, moderatorLogin, completedCount))
}

func (h *Handler) EditDrugInPrescription(ctx *gin.Context) {
	drugID, err := strconv.Atoi(ctx.Param("drug_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	prescriptionID, err := strconv.Atoi(ctx.Param("prescription_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	var j serializer.PrescriptionDrugJSON
	if err := ctx.BindJSON(&j); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	item, err := h.Repository.EditDrugInPrescription(prescriptionID, drugID, j)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}
	ctx.JSON(http.StatusOK, serializer.PrescriptionDrugToJSON(item))
}
