package dto

type CreateProjectRequest struct {
	Name    string  `json:"name" binding:"required"`
	OwnerID *uint64 `json:"owner_id"`
}
