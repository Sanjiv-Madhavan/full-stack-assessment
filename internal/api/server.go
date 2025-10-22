package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"

	"full-stack-assesment/internal/helpers"
	"full-stack-assesment/internal/scheme"
)

var _ ServerInterface = (*Server)(nil)

type Server struct {
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}

func (s *Server) GetHealth(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, "OK")
}

func (s *Server) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body scheme.NewProject
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		helpers.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if l := len(name); l > 128 {
		helpers.WriteError(w, http.StatusBadRequest, "name too long (max 128)")
		return
	}

	id := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	const q = `
		INSERT INTO projects (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`
	if _, err := s.db.ExecContext(ctx, q, id.String(), name, now, now); err != nil {
		if errStr := strings.ToLower(err.Error()); strings.Contains(errStr, "unique") && strings.Contains(errStr, "projects.name") {
			helpers.WriteError(w, http.StatusConflict, "project name already exists")
			return
		}
		helpers.WriteError(w, http.StatusInternalServerError, "failed to create project")
		return
	}

	created := scheme.Project{
		Id:        types.UUID(id),
		Name:      name,
		CreatedAt: parseTimeOrNow(now),
		UpdatedAt: parseTimeOrNow(now),
	}
	helpers.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) ListProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	const q = `
		SELECT id, name, created_at, updated_at
		FROM projects
		ORDER BY updated_at DESC, name ASC
	`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "failed to query projects")
		return
	}
	defer func() { _ = rows.Close() }()

	projects := make([]scheme.Project, 0, 16)
	for rows.Next() {
		var idStr, name, created, updated string
		if err := rows.Scan(&idStr, &name, &created, &updated); err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "failed to read project row")
			return
		}
		u, parseErr := uuid.Parse(idStr)
		if parseErr != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "invalid project id in database")
			return
		}
		projects = append(projects, scheme.Project{
			Id:        types.UUID(u),
			Name:      name,
			CreatedAt: parseTimeOrNow(created),
			UpdatedAt: parseTimeOrNow(updated),
		})
	}
	if err := rows.Err(); err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "rows error")
		return
	}

	helpers.WriteJSON(w, http.StatusOK, projects)
}

func parseTimeOrNow(s string) time.Time {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	return time.Now().UTC()
}

// ensureProjectExists returns 404 if project is missing.
func (s *Server) ensureProjectExists(ctx context.Context, projectID string) error {
	const q = `SELECT 1 FROM projects WHERE id = ?`
	var one int
	if err := s.db.QueryRowContext(ctx, q, projectID).Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return &helpers.NotFoundErr{}
		}
		return err
	}
	return nil
}
