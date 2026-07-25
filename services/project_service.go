// Package services holds business logic that orchestrates repositories and other
// clients. Controllers call services; services call repositories.
package services

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/dto"
)

type ProjectService interface {
	CreateProject(ctx context.Context, orgID, ownerID uint64, req dto.CreateProjectRequest) (*dbdto.Project, error)
	GetProject(ctx context.Context, id uint64) (*dbdto.Project, error)
	ListProjects(ctx context.Context, orgID uint64) ([]dbdto.Project, error)
}
