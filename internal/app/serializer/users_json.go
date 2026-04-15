package serializer

import "web_backend/internal/app/ds"

type SignInRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignUpRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignUpResponse struct {
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
}

func SignUpResponseFromUser(user ds.Users) SignUpResponse {
	return SignUpResponse{
		Login:       user.Login,
		IsModerator: user.IsModerator,
	}
}

func SignUpRequestToUser(j SignUpRequest) ds.Users {
	return ds.Users{
		Login:    j.Login,
		Password: j.Password,
	}
}
