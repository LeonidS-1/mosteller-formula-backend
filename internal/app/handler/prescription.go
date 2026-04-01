package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"web_backend/internal/app/repository"
	"web_backend/internal/app/serializer"
)

func (h *Handler) GetPrescriptionCart(ctx *gin.Context) {
	creatorID := uint(h.Repository.GetCreatorID())
	count := h.Repository.GetPrescriptionDrugCount(creatorID)
	if count == 0 {
		rx, err := h.Repository.CheckCurrentDraft(creatorID)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"status":      "no_draft",
				"drugs_count": 0,
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"id":          rx.PrescriptionID,
			"drugs_count": 0,
		})
		return
	}
	prescriptionID := h.Repository.GetActivePrescriptionID(creatorID)
	ctx.JSON(http.StatusOK, gin.H{
		"id":          prescriptionID,
		"drugs_count": count,
	})
}

func (h *Handler) GetAllPrescriptions(ctx *gin.Context) {
	fromDate := ctx.Query("from-date")
	var from, to time.Time
	if fromDate != "" {
		t, err := time.Parse("2006-01-02", fromDate)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		from = t
	}
	toDate := ctx.Query("to-date")
	if toDate != "" {
		t, err := time.Parse("2006-01-02", toDate)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		to = t
	}
	status := ctx.Query("status")
	list, err := h.Repository.GetAllPrescriptions(from, to, status)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	resp := make([]serializer.PrescriptionJSON, 0, len(list))
	for _, rx := range list {
		creatorLogin, moderatorLogin, _ := h.Repository.GetModeratorAndCreatorLogin(rx)
		completedCount, _ := h.Repository.GetCompletedDoseLineCount(rx.PrescriptionID)
		resp = append(resp, serializer.PrescriptionToJSON(rx, creatorLogin, moderatorLogin, completedCount))
	}
	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) GetPrescription(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	rx, err := h.Repository.GetSinglePrescription(id)
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
	items, err := h.Repository.GetPrescriptionItems(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	creatorLogin, moderatorLogin, _ := h.Repository.GetModeratorAndCreatorLogin(rx)
	completedCount, _ := h.Repository.GetCompletedDoseLineCount(rx.PrescriptionID)
	itemsResp := make([]serializer.PrescriptionDrugDetailJSON, 0, len(items))
	for _, item := range items {
		itemsResp = append(itemsResp, serializer.PrescriptionDrugDetailToJSON(item))
	}
	ctx.JSON(http.StatusOK, gin.H{
		"prescription": serializer.PrescriptionToJSON(rx, creatorLogin, moderatorLogin, completedCount),
		"drugs":        itemsResp,
	})
}

func (h *Handler) EditPrescription(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	var j serializer.PrescriptionEditJSON
	if err := ctx.BindJSON(&j); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	rx, err := h.Repository.EditPrescription(id, j)
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

func (h *Handler) FormPrescription(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	rx, err := h.Repository.FormPrescription(id)
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

func (h *Handler) FinishPrescription(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	var statusJSON serializer.StatusJSON
	if err := ctx.BindJSON(&statusJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	rx, err := h.Repository.FinishPrescription(id, statusJSON.Status)
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

func (h *Handler) DeletePrescription(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	_, err = h.Repository.DeletePrescription(id)
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
	ctx.JSON(http.StatusOK, gin.H{"message": "Рецепт удалён"})
}
