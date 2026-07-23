package impl

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	"error-logging/dto"
	"error-logging/pkg/config"
	"error-logging/services"
)

// apiKeyScheme prefixes every minted key so raw keys are recognisable on sight.
const apiKeyScheme = "elk_live_"

type serviceService struct {
	svcs     repository.ServiceRepository
	projects repository.ProjectRepository
	appCfg   config.AppConfig
}

func NewServiceService(
	svcs repository.ServiceRepository,
	projects repository.ProjectRepository,
	appCfg config.AppConfig,
) services.ServiceService {
	return &serviceService{svcs: svcs, projects: projects, appCfg: appCfg}
}

func (s *serviceService) CreateService(ctx context.Context, projectID uint64, req dto.CreateServiceRequest) (*dto.CreateServiceResponse, error) {
	// Ensure the parent project exists before minting a service under it.
	if _, err := s.projects.GetByID(ctx, projectID); err != nil {
		return nil, fmt.Errorf("project %d not found: %w", projectID, err)
	}

	publicID, err := randomHex(12)
	if err != nil {
		return nil, fmt.Errorf("generate public id: %w", err)
	}

	// Mint the API key: only its hash is persisted; the raw key is returned once.
	secret, err := randomHex(24)
	if err != nil {
		return nil, fmt.Errorf("generate api key: %w", err)
	}
	rawKey := apiKeyScheme + secret

	service := &dbdto.Service{
		ProjectID:    projectID,
		Name:         req.Name,
		PublicID:     publicID,
		APIKeyHash:   hashAPIKey(rawKey),
		APIKeyPrefix: apiKeyScheme + secret[:4],
	}
	if err := s.svcs.Create(ctx, service); err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}

	return &dto.CreateServiceResponse{
		Service:   service,
		APIKey:    rawKey,
		IngestURL: fmt.Sprintf("%s/api/v1/logs/%s", s.appCfg.BaseURL, publicID),
	}, nil
}

func (s *serviceService) ListServices(ctx context.Context, projectID uint64) ([]dbdto.Service, error) {
	return s.svcs.ListByProject(ctx, projectID)
}

// randomHex returns a random hex string encoding nBytes of entropy.
func randomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashAPIKey returns the sha256 hex digest of a raw key, used for constant-schema
// lookup at ingest time and to persist keys without storing the secret itself.
func hashAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:])
}
