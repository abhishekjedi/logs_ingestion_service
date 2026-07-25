package impl

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	"error-logging/dto"
	"error-logging/services"
)

type projectService struct {
	projects repository.ProjectRepository
}

func NewProjectService(projects repository.ProjectRepository) services.ProjectService {
	return &projectService{projects: projects}
}

func (s *projectService) CreateProject(ctx context.Context, orgID, ownerID uint64, req dto.CreateProjectRequest) (*dbdto.Project, error) {
	project := &dbdto.Project{
		OrgID:   &orgID,
		OwnerID: &ownerID,
		Name:    req.Name,
	}
	if err := s.projects.Create(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *projectService) GetProject(ctx context.Context, id uint64) (*dbdto.Project, error) {
	return s.projects.GetByID(ctx, id)
}

func (s *projectService) ListProjects(ctx context.Context, orgID uint64) ([]dbdto.Project, error) {
	return s.projects.ListByOrg(ctx, orgID)
}
