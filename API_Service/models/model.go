package models

type MessageResponse struct {
	Message string `json:"message"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}
