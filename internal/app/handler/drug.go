package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"web_backend/internal/app/ds"
	"web_backend/internal/app/repository"
	"web_backend/internal/app/serializer"
)

// GetDrugs godoc
// @Summary Получить список препаратов
// @Description Возвращает список препаратов, доступных для назначения, с фильтром по названию.
// @Tags drugs
// @Produce json
// @Param Title query string false "Название препарата для поиска"
// @Success 200 {array} serializer.DrugJSON "Список препаратов"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /drugs [get]
func (h *Handler) GetDrugs(ctx *gin.Context) {
	var drugs []ds.Drug
	var err error
	searchQuery := ctx.Query("Title")
	if searchQuery == "" {
		drugs, err = h.Repository.GetDrugs()
	} else {
		drugs, err = h.Repository.GetDrugsByTitle(searchQuery)
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	resp := make([]serializer.DrugJSON, 0, len(drugs))
	for _, d := range drugs {
		resp = append(resp, serializer.DrugToJSON(d))
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetDrug godoc
// @Summary Получить препарат по ID
// @Description Возвращает карточку препарата по идентификатору.
// @Tags drugs
// @Produce json
// @Param id path int true "ID препарата"
// @Success 200 {object} serializer.DrugJSON "Данные препарата"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 404 {object} map[string]string "Препарат не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /drugs/{id} [get]
func (h *Handler) GetDrug(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	drug, err := h.Repository.GetDrug(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}
	ctx.JSON(http.StatusOK, serializer.DrugToJSON(*drug))
}

// CreateDrug godoc
// @Summary Создать препарат
// @Description Создает новый препарат, при необходимости загружает изображение и видео в MinIO.
// @Tags drugs
// @Accept json
// @Produce json
// @Param drug body serializer.DrugJSON true "Данные препарата"
// @Success 201 {object} serializer.DrugJSON "Созданный препарат"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /drugs [post]
func (h *Handler) CreateDrug(ctx *gin.Context) {
	contentType := ctx.GetHeader("Content-Type")
	var j serializer.DrugJSON
	if strings.HasPrefix(contentType, "application/json") {
		if err := ctx.BindJSON(&j); err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
	} else {
		title := ctx.PostForm("title")
		desc := ctx.PostForm("description")
		if title == "" || desc == "" {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("title and description are required"))
			return
		}
		adult := 0.0
		perM2 := 0.0
		maxDaily := 0.0
		if v := ctx.PostForm("adult_dose_mg"); v != "" {
			fmt.Sscanf(v, "%f", &adult)
		}
		if v := ctx.PostForm("dose_per_m2_mg"); v != "" {
			fmt.Sscanf(v, "%f", &perM2)
		}
		if v := ctx.PostForm("max_daily_mg"); v != "" {
			fmt.Sscanf(v, "%f", &maxDaily)
		}
		j = serializer.DrugJSON{
			Title:       title,
			Description: desc,
			AdultDoseMg: adult,
			DosePerM2Mg: perM2,
			MaxDailyMg:  maxDaily,
		}
	}

	drug, err := h.Repository.CreateDrug(j)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if imageFile, err := ctx.FormFile("image"); err == nil {
		d, err := h.Repository.AddDrugPhoto(ctx, int(drug.DrugID), imageFile)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
		drug = *d
	}
	if videoFile, err := ctx.FormFile("video"); err == nil {
		d, err := h.Repository.AddDrugVideo(ctx, int(drug.DrugID), videoFile)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
		drug = *d
	}

	ctx.Header("Location", fmt.Sprintf("/api/drugs/%d", drug.DrugID))
	ctx.JSON(http.StatusCreated, serializer.DrugToJSON(drug))
}
