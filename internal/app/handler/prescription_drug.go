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

// AddDrugToPrescription godoc
// @Summary Добавить препарат в рецепт
// @Description Добавляет препарат в черновик рецепта пользователя.
// @Tags prescription_drugs
// @Produce json
// @Param drug_id path int true "ID препарата"
// @Success 200 {object} serializer.PrescriptionJSON "Обновленный рецепт"
// @Success 201 {object} serializer.PrescriptionJSON "Создан новый черновик рецепта"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 404 {object} map[string]string "Препарат не найден"
// @Failure 409 {object} map[string]string "Препарат уже в рецепте"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /prescription_drugs/add/{drug_id} [post]
func (h *Handler) AddDrugToPrescription(ctx *gin.Context) {
	drugIDStr := ctx.Param("drug_id")
	drugID, err := strconv.Atoi(drugIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	creatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
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

// DeleteDrugFromPrescription godoc
// @Summary Удалить препарат из рецепта
// @Description Удаляет препарат из черновика рецепта пользователя.
// @Tags prescription_drugs
// @Produce json
// @Param drug_id path int true "ID препарата"
// @Param prescription_id path int true "ID рецепта"
// @Success 200 {object} serializer.PrescriptionJSON "Обновленный рецепт"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Не найдено"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /prescription_drugs/{drug_id}/{prescription_id} [delete]
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
	creatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	rx, err := h.Repository.DeleteDrugFromPrescription(prescriptionID, drugID, creatorID)
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

// EditDrugInPrescription godoc
// @Summary Изменить параметры препарата в рецепте
// @Description Изменяет рост/вес в строке назначения и пересчитывает дозу.
// @Tags prescription_drugs
// @Accept json
// @Produce json
// @Param drug_id path int true "ID препарата"
// @Param prescription_id path int true "ID рецепта"
// @Param body body serializer.PrescriptionDrugJSON true "Новые параметры"
// @Success 200 {object} serializer.PrescriptionDrugJSON "Обновленные данные строки назначения"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Не найдено"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /prescription_drugs/{drug_id}/{prescription_id} [put]
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
	creatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	item, err := h.Repository.EditDrugInPrescription(prescriptionID, drugID, creatorID, j)
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
