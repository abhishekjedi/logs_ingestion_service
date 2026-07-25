package dto

// InviteMemberRequest adds a member to an org. Role is "admin" or "member"
// (defaults to member); owner cannot be assigned via invite.
type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role"`
}
