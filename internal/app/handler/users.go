package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"web_backend/internal/app/repository"
	"web_backend/internal/app/serializer"
)

func (h *Handler) CreateUser(ctx *gin.Context) {
	var j serializer.UserJSON
	if err := ctx.BindJSON(&j); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	if j.Login == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("field 'login' is required"))
		return
	}
	if j.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("field 'password' is required"))
		return
	}
	user, err := h.Repository.CreateUser(j)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			h.errorHandler(ctx, http.StatusConflict, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}
	ctx.Header("Location", fmt.Sprintf("/api/users/%d", user.UserID))
	ctx.JSON(http.StatusCreated, serializer.UserToJSON(user))
}

func (h *Handler) SignIn(ctx *gin.Context) {
	var j serializer.UserJSON
	if err := ctx.BindJSON(&j); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	if j.Login == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("field 'login' is required"))
		return
	}
	if j.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("field 'password' is required"))
		return
	}
	user, err := h.Repository.SignIn(j)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) || err.Error() == "неверный логин или пароль" {
			h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("invalid login or password"))
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}
	ctx.JSON(http.StatusOK, serializer.UserToJSON(user))
}

func (h *Handler) SignOut(ctx *gin.Context) {
	h.Repository.SignOut()
	ctx.JSON(http.StatusOK, gin.H{
		"status": "signed_out",
	})
}
