package dto

type CreateServiceRequest struct {
	Name string `json:"name" binding:"required"`
}
