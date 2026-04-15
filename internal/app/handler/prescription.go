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

// GetPrescriptionCart godoc
// @Summary Получить корзину рецепта
// @Description Возвращает информацию о текущем черновике рецепта пользователя.
// @Tags prescriptions
// @Produce json
// @Success 200 {object} map[string]interface{} "Данные корзины"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /prescriptions/cart [get]
func (h *Handler) GetPrescriptionCart(ctx *gin.Context) {
	creatorID, err := getUserID(ctx)
	if err != nil || creatorID == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"has_draft":   false,
			"drugs_count": 0,
		})
		return
	}

	count := h.Repository.GetPrescriptionDrugCount(creatorID)
	if count == 0 {
		rx, err := h.Repository.CheckCurrentDraft(creatorID)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"has_draft":   false,
				"drugs_count": 0,
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"id":          rx.PrescriptionID,
			"has_draft":   true,
			"drugs_count": 0,
		})
		return
	}
	prescriptionID := h.Repository.GetActivePrescriptionID(creatorID)
	ctx.JSON(http.StatusOK, gin.H{
		"id":          prescriptionID,
		"has_draft":   true,
		"drugs_count": count,
	})
}

// GetAllPrescriptions godoc
// @Summary Получить список рецептов
// @Description Возвращает список рецептов с фильтрацией по дате и статусу.
// @Tags prescriptions
// @Produce json
// @Param from-date query string false "Начальная дата (YYYY-MM-DD)"
// @Param to-date query string false "Конечная дата (YYYY-MM-DD)"
// @Param status query string false "Статус рецепта"
// @Success 200 {array} serializer.PrescriptionJSON "Список рецептов"
// @Failure 400 {object} map[string]string "Неверный формат даты"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /prescriptions [get]
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

	userID, _ := getUserID(ctx)
	creatorID := uint(0)
	if isModVal, ok := ctx.Get("is_moderator"); ok {
		if isMod, _ := isModVal.(bool); !isMod {
			creatorID = userID
		}
	} else {
		creatorID = userID
	}

	list, err := h.Repository.GetAllPrescriptions(from, to, status, creatorID)
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

// GetPrescription godoc
// @Summary Получить рецепт по ID
// @Description Возвращает рецепт и список назначенных препаратов.
// @Tags prescriptions
// @Produce json
// @Param id path int true "ID рецепта"
// @Success 200 {object} map[string]interface{} "Рецепт и препараты"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Рецепт не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /prescriptions/{id} [get]
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

	userID, _ := getUserID(ctx)
	if isModVal, ok := ctx.Get("is_moderator"); ok {
		if isMod, _ := isModVal.(bool); !isMod && rx.CreatorID != userID {
			h.errorHandler(ctx, http.StatusForbidden, repository.ErrNotAllowed)
			return
		}
	} else if rx.CreatorID != userID {
		h.errorHandler(ctx, http.StatusForbidden, repository.ErrNotAllowed)
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

// EditPrescription godoc
// @Summary Изменить рецепт
// @Description Обновляет поля черновика рецепта: ФИО врача и примечания.
// @Tags prescriptions
// @Accept json
// @Produce json
// @Param id path int true "ID рецепта"
// @Param body body serializer.PrescriptionEditJSON true "Новые данные рецепта"
// @Success 200 {object} serializer.PrescriptionJSON "Обновленный рецепт"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Рецепт не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /prescriptions/{id} [put]
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
	creatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	rx, err := h.Repository.EditPrescription(id, creatorID, j)
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

// FormPrescription godoc
// @Summary Сформировать рецепт
// @Description Переводит рецепт из черновика в статус formed и пересчитывает дозы.
// @Tags prescriptions
// @Produce json
// @Param id path int true "ID рецепта"
// @Success 200 {object} serializer.PrescriptionJSON "Сформированный рецепт"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Рецепт не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /prescriptions/{id}/form [put]
func (h *Handler) FormPrescription(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	creatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	rx, err := h.Repository.FormPrescription(id, creatorID)
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

// FinishPrescription godoc
// @Summary Завершить или отклонить рецепт
// @Description Меняет статус сформированного рецепта на completed или rejected (только модератор).
// @Tags prescriptions
// @Accept json
// @Produce json
// @Param id path int true "ID рецепта"
// @Param status body serializer.StatusJSON true "Новый статус"
// @Success 200 {object} serializer.PrescriptionJSON "Обновленный рецепт"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Рецепт не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /prescriptions/{id}/finish [put]
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
	moderatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	rx, err := h.Repository.FinishPrescription(id, statusJSON.Status, moderatorID)
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

// DeletePrescription godoc
// @Summary Удалить рецепт
// @Description Выполняет логическое удаление черновика рецепта.
// @Tags prescriptions
// @Produce json
// @Param id path int true "ID рецепта"
// @Success 200 {object} map[string]string "Рецепт удалён"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Рецепт не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /prescriptions/{id} [delete]
func (h *Handler) DeletePrescription(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	creatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	_, err = h.Repository.DeletePrescription(id, creatorID)
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
