package dto

type CreateOrgRequest struct {
	Name string `json:"name" binding:"required"`
}
