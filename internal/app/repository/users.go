package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"web_backend/internal/app/ds"
	"web_backend/internal/app/serializer"
)

func (r *Repository) GetUserByID(id int) (ds.Users, error) {
	var user ds.Users
	if id <= 0 {
		return ds.Users{}, fmt.Errorf("неверный id: должен быть > 0")
	}
	err := r.db.Where("user_id = ?", id).First(&user).Error
	if err != nil {
		return ds.Users{}, err
	}
	return user, nil
}

func (r *Repository) GetUserByLogin(login string) (ds.Users, error) {
	var user ds.Users
	if login == "" {
		return ds.Users{}, errors.New("логин не может быть пустым")
	}
	err := r.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.Users{}, fmt.Errorf("%w: пользователь с логином %s не найден", ErrNotFound, login)
		}
		return ds.Users{}, err
	}
	return user, nil
}

func (r *Repository) CreateUser(j serializer.UserJSON) (ds.Users, error) {
	user := serializer.UserFromJSON(j)
	if user.Login == "" {
		return ds.Users{}, errors.New("логин обязателен для заполнения")
	}
	if user.Password == "" {
		return ds.Users{}, errors.New("пароль обязателен для заполнения")
	}
	_, err := r.GetUserByLogin(user.Login)
	if err == nil {
		return ds.Users{}, fmt.Errorf("%w: пользователь с логином %s уже существует", ErrAlreadyExists, user.Login)
	}
	if !errors.Is(err, ErrNotFound) {
		return ds.Users{}, err
	}
	if err := r.db.Create(&user).Error; err != nil {
		return ds.Users{}, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}
	return user, nil
}

func (r *Repository) SignIn(j serializer.UserJSON) (ds.Users, error) {
	if j.Login == "" {
		return ds.Users{}, errors.New("логин обязателен для заполнения")
	}
	if j.Password == "" {
		return ds.Users{}, errors.New("пароль обязателен для заполнения")
	}
	user, err := r.GetUserByLogin(j.Login)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ds.Users{}, errors.New("неверный логин или пароль")
		}
		return ds.Users{}, err
	}
	if user.Password != j.Password {
		return ds.Users{}, errors.New("неверный логин или пароль")
	}
	r.SetUserID(int(user.UserID))
	return user, nil
}
