package repo

import (
	"context"
	"database/sql"
	"errors"
	"full-stack-assesment/internal/apierrors"
	"full-stack-assesment/internal/helpers"
	"full-stack-assesment/internal/scheme"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

var (
	ErrNotFound = errors.New("Project not found")
	ErrConflict = errors.New("ProjectName already exists")
)

type ProjectsRepository interface {
	Create(ctx context.Context, t scheme.Project) error
	UpdateStatus(ctx context.Context, id string, completed bool) error
	Get(ctx context.Context, id string) (scheme.Project, error)
	List(ctx context.Context) ([]scheme.Project, error)
	Delete(ctx context.Context, id string) error
}

type SQLiteProjectsRepo struct {
	db *sql.DB
}

func NewSQLiteProjectsRepo(db *sql.DB) *SQLiteProjectsRepo {
	return &SQLiteProjectsRepo{db: db}
}

func (r *SQLiteProjectsRepo) Create(ctx context.Context, project scheme.Project) error {
	const q = `
		INSERT INTO projects (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`
	if _, err := r.db.ExecContext(ctx, q, project.Id.String(), project.Name, project.CreatedAt, project.UpdatedAt); err != nil {
		return err
	}
	return nil
}

func (r *SQLiteProjectsRepo) List(ctx context.Context) ([]scheme.Project, error) {
	const q = `
		SELECT id, name, created_at, updated_at
		FROM projects
		ORDER BY updated_at DESC, name ASC
	`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	projects := make([]scheme.Project, 0, 16)
	for rows.Next() {
		var idStr, name, created, updated string
		if err := rows.Scan(&idStr, &name, &created, &updated); err != nil {
			return nil, err
		}
		u, parseErr := uuid.Parse(idStr)
		if parseErr != nil {
			return nil, err
		}
		projects = append(projects, scheme.Project{
			Id:        types.UUID(u),
			Name:      name,
			CreatedAt: helpers.ParseTimeOrNow(created),
			UpdatedAt: helpers.ParseTimeOrNow(updated),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *SQLiteProjectsRepo) EnsureProjectExists(ctx context.Context, projectID string) error {
	const q = `SELECT 1 FROM projects WHERE id = ?`
	var one int
	if err := r.db.QueryRowContext(ctx, q, projectID).Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return apierrors.ErrProjectNotFound
		}
		return err
	}
	return nil
}
