package dto

type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role"`
}
