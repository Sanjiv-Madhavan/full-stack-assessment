package service

import (
	"context"
	"full-stack-assesment/internal/apierrors"
	repo "full-stack-assesment/internal/repo/projects"
	"full-stack-assesment/internal/scheme"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

type ProjectsService struct {
	repo repo.SQLiteProjectsRepo
}

func NewService(repo repo.SQLiteProjectsRepo) *ProjectsService {
	return &ProjectsService{repo: repo}
}

func (s *ProjectsService) CreateProject(ctx context.Context, newProject scheme.NewProject) (*scheme.Project, error) {
	name := strings.TrimSpace(newProject.Name)
	if name == "" {
		return nil, apierrors.ErrProjectNameRequired
	}
	if l := len(name); l > 128 {
		return nil, apierrors.ErrProjectNameTooLong
	}

	id := uuid.New()
	now := time.Now().UTC()

	created := scheme.Project{
		Id:        types.UUID(id),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.repo.Create(ctx, created)
	if err != nil {
		if errStr := strings.ToLower(err.Error()); strings.Contains(errStr, "unique") && strings.Contains(errStr, "projects.name") {
			return nil, apierrors.ErrProjectNameExists
		}
		return nil, err
	}

	return &created, nil
}

func (s *ProjectsService) ListProject(ctx context.Context) ([]scheme.Project, error) {
	projects, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (s *ProjectsService) EnsureProjectExists(ctx context.Context, projectID string) error {
	if err := s.repo.EnsureProjectExists(ctx, projectID); err != nil {
		return err
	}
	return nil
}
