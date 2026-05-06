package serializer

import "web_backend/internal/app/ds"

type SignInRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignInResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
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

func SignInResponseFromUser(token string, user ds.Users) SignInResponse {
	return SignInResponse{
		Token: token,
		Role:  roleFromModeratorFlag(user.IsModerator),
	}
}

func roleFromModeratorFlag(isModerator bool) string {
	if isModerator {
		return "moderator"
	}
	return "user"
}

func SignUpRequestToUser(j SignUpRequest) ds.Users {
	return ds.Users{
		Login:    j.Login,
		Password: j.Password,
	}
}
