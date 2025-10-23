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
	openapi_types "github.com/oapi-codegen/runtime/types"

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

// func (s *Server) ListTasks(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, params ListTasksParams) {
// 	ctx := r.Context()

// 	projectUUID, ok := helpers.ParseUUIDParam(w, r, "projectId")
// 	if !ok {
// 		return
// 	}
// 	projectID := projectUUID.String()

// 	// Ensure project exists (helpful error early)
// 	if err := s.ensureProjectExists(ctx, projectID); err != nil {
// 		if _, nf := err.(*helpers.NotFoundErr); nf {
// 			helpers.WriteError(w, http.StatusNotFound, "project not found")
// 			return
// 		}
// 		helpers.WriteError(w, http.StatusInternalServerError, "failed to verify project")
// 		return
// 	}

// 	qp := r.URL.Query()
// 	limit, offset := helpers.ParseLimitOffset(qp)

// 	var (
// 		args    []any
// 		clauses = []string{"project_id = ?"}
// 	)
// 	args = append(args, projectID)

// 	if statusStr := qp.Get("status"); statusStr != "" {
// 		if norm, ok := helpers.NormalizeStatus(statusStr); ok {
// 			clauses = append(clauses, "status = ?")
// 			args = append(args, norm)
// 		} else {
// 			helpers.WriteError(w, http.StatusBadRequest, "invalid status; use TODO|IN_PROGRESS|DONE")
// 			return
// 		}
// 	}
// 	if q := strings.TrimSpace(qp.Get("q")); q != "" {
// 		clauses = append(clauses, "title LIKE ?")
// 		args = append(args, "%"+q+"%")
// 	}

// 	// ORDER + LIMIT/OFFSET last
// 	args = append(args, limit, offset)

// 	stmt := `
// 		SELECT id, project_id, title, description, status, created_at, updated_at
// 		FROM tasks
// 		WHERE ` + strings.Join(clauses, " AND ") + `
// 		ORDER BY updated_at DESC
// 		LIMIT ? OFFSET ?;
// 	`

// 	rows, err := s.db.QueryContext(ctx, stmt, args...)
// 	if err != nil {
// 		helpers.WriteError(w, http.StatusInternalServerError, "failed to query tasks")
// 		return
// 	}
// 	defer rows.Close()

// 	out := make([]scheme.Task, 0, limit)
// 	for rows.Next() {
// 		var idStr, projStr, title, desc, status, created, updated string
// 		if err := rows.Scan(&idStr, &projStr, &title, &desc, &status, &created, &updated); err != nil {
// 			helpers.WriteError(w, http.StatusInternalServerError, "failed to read task row")
// 			return
// 		}
// 		out = append(out, scheme.Task{
// 			Id:        helpers.MustUUID(idStr),
// 			ProjectId: helpers.MustUUID(projStr),
// 			Title:     title,
// 			Description: func() *string {
// 				if desc == "" {
// 					return nil
// 				}
// 				return &desc
// 			}(),
// 			Status:    scheme.TaskStatus(status),
// 			CreatedAt: parseTimeOrNow(created),
// 			UpdatedAt: parseTimeOrNow(updated),
// 		})
// 	}
// 	if err := rows.Err(); err != nil {
// 		helpers.WriteError(w, http.StatusInternalServerError, "rows error")
// 		return
// 	}
// 	helpers.WriteJSON(w, http.StatusOK, out)
// }

func (s *Server) ListTasks(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, params scheme.ListTasksParams) {
	ctx := r.Context()

	// 404 if project missing (early)
	if err := s.ensureProjectExists(ctx, projectId.String()); err != nil {
		if _, nf := err.(*helpers.NotFoundErr); nf {
			helpers.WriteError(w, http.StatusNotFound, "project not found")
			return
		}
		helpers.WriteError(w, http.StatusInternalServerError, "failed to verify project")
		return
	}

	// Normalize filters
	var (
		where  = []string{"project_id = ?"}
		args   = []any{projectId.String()}
		limit  = 50
		offset = 0
	)

	if params.Status != nil {
		if norm, ok := helpers.NormalizeStatus(string(*params.Status)); ok {
			where = append(where, "status = ?")
			args = append(args, norm)
		} else {
			helpers.WriteError(w, http.StatusBadRequest, "invalid status; use TODO|IN_PROGRESS|DONE")
			return
		}
	}
	if params.Q != nil {
		q := strings.TrimSpace(*params.Q)
		if q != "" {
			where = append(where, "title LIKE ?")
			args = append(args, "%"+q+"%")
		}
	}
	if params.Limit != nil {
		limit = helpers.ClampInt(*params.Limit, 1, 200, 50)
	}
	if params.Offset != nil && *params.Offset >= 0 {
		offset = *params.Offset
	}

	args = append(args, limit, offset)

	stmt := `
		SELECT id, project_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?;
	`

	rows, err := s.db.QueryContext(ctx, stmt, args...)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "failed to query tasks")
		return
	}
	defer rows.Close()

	out := make([]scheme.Task, 0, limit)
	for rows.Next() {
		var idStr, projStr, title, desc, status, created, updated string
		if err := rows.Scan(&idStr, &projStr, &title, &desc, &status, &created, &updated); err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "failed to read task row")
			return
		}
		var descPtr *string
		if strings.TrimSpace(desc) != "" {
			cp := desc
			descPtr = &cp
		}
		out = append(out, scheme.Task{
			Id:          helpers.MustUUID(idStr),
			ProjectId:   helpers.MustUUID(projStr),
			Title:       title,
			Description: descPtr,
			Status:      scheme.TaskStatus(status),
			CreatedAt:   helpers.ParseTimeOrNow(created),
			UpdatedAt:   helpers.ParseTimeOrNow(updated),
		})
	}
	helpers.WriteJSON(w, http.StatusOK, out)
}

func (s *Server) CreateTask(w http.ResponseWriter, r *http.Request, projectUUID types.UUID) {
	ctx := r.Context()

	// projectUUID, ok := helpers.ParseUUIDParam(w, r, "projectId")
	// if !ok {
	// 	return
	// }
	projectID := projectUUID.String()

	var body scheme.NewTask
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	title := strings.TrimSpace(body.Title)
	if title == "" {
		helpers.WriteError(w, http.StatusBadRequest, "title is required")
		return
	}
	if len(title) > 200 {
		helpers.WriteError(w, http.StatusBadRequest, "title too long (max 200)")
		return
	}

	status := "TODO"
	if body.Status != nil {
		if norm, ok := helpers.NormalizeStatus(string(*body.Status)); ok {
			status = norm
		} else {
			helpers.WriteError(w, http.StatusBadRequest, "invalid status; use TODO|IN_PROGRESS|DONE")
			return
		}
	}

	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		if _, nf := err.(*helpers.NotFoundErr); nf {
			helpers.WriteError(w, http.StatusNotFound, "project not found")
			return
		}
		helpers.WriteError(w, http.StatusInternalServerError, "failed to verify project")
		return
	}

	id := uuid.New()
	now := helpers.NowRFC3339()

	const q = `
		INSERT INTO tasks (id, project_id, title, description, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	var desc string
	if body.Description != nil {
		desc = strings.TrimSpace(*body.Description)
	}
	if _, err := s.db.ExecContext(ctx, q, id.String(), projectID, title, desc, status, now, now); err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	created := scheme.Task{
		Id:        types.UUID(id),
		ProjectId: types.UUID(projectUUID),
		Title:     title,
		Description: func() *string {
			if desc == "" {
				return nil
			}
			return &desc
		}(),
		Status:    scheme.TaskStatus(status),
		CreatedAt: parseTimeOrNow(now),
		UpdatedAt: parseTimeOrNow(now),
	}
	helpers.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) GetTask(w http.ResponseWriter, r *http.Request, projectUUID openapi_types.UUID, taskUUID openapi_types.UUID) {
	ctx := r.Context()

	// projectUUID, ok := helpers.ParseUUIDParam(w, r, "projectId")
	// if !ok {
	// 	return
	// }
	// taskUUID, ok := helpers.ParseUUIDParam(w, r, "id")
	// if !ok {
	// 	return
	// }

	const q = `
		SELECT id, project_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = ? AND project_id = ?;
	`
	var idStr, projStr, title, desc, status, created, updated string
	err := s.db.QueryRowContext(ctx, q, taskUUID.String(), projectUUID.String()).
		Scan(&idStr, &projStr, &title, &desc, &status, &created, &updated)
	if err == sql.ErrNoRows {
		helpers.WriteError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "failed to fetch task")
		return
	}

	out := scheme.Task{
		Id:        helpers.MustUUID(idStr),
		ProjectId: helpers.MustUUID(projStr),
		Title:     title,
		Description: func() *string {
			if desc == "" {
				return nil
			}
			return &desc
		}(),
		Status:    scheme.TaskStatus(status),
		CreatedAt: parseTimeOrNow(created),
		UpdatedAt: parseTimeOrNow(updated),
	}
	helpers.WriteJSON(w, http.StatusOK, out)
}

func (s *Server) DeleteTask(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, taskId openapi_types.UUID) {
	ctx := r.Context()
	const q = `DELETE FROM tasks WHERE id = ? AND project_id = ?;`

	res, err := s.db.ExecContext(ctx, q, taskId.String(), projectId.String())
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "failed to delete task")
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		helpers.WriteError(w, http.StatusNotFound, "task not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) UpdateTask(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, taskId openapi_types.UUID) {
	ctx := r.Context()

	var body scheme.UpdateTask
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	set := make([]string, 0, 4)
	args := make([]any, 0, 6)

	if body.Title != nil {
		title := strings.TrimSpace(*body.Title)
		if title == "" {
			helpers.WriteError(w, http.StatusBadRequest, "title cannot be empty")
			return
		}
		if len(title) > 200 {
			helpers.WriteError(w, http.StatusBadRequest, "title too long (max 200)")
			return
		}
		set = append(set, "title = ?")
		args = append(args, title)
	}
	if body.Description != nil {
		desc := strings.TrimSpace(*body.Description)
		set = append(set, "description = ?")
		args = append(args, desc)
	}
	if body.Status != nil {
		if norm, ok := helpers.NormalizeStatus(string(*body.Status)); ok {
			set = append(set, "status = ?")
			args = append(args, norm)
		} else {
			helpers.WriteError(w, http.StatusBadRequest, "invalid status; use TODO|IN_PROGRESS|DONE")
			return
		}
	}

	// No-op update?
	if len(set) == 0 {
		// Return current task
		s.GetTask(w, r, projectId, taskId)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	set = append(set, "updated_at = ?")
	args = append(args, now)

	stmt := `
		UPDATE tasks
		SET ` + strings.Join(set, ", ") + `
		WHERE id = ? AND project_id = ?;
	`
	args = append(args, taskId.String(), projectId.String())

	res, err := s.db.ExecContext(ctx, stmt, args...)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "failed to update task")
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		helpers.WriteError(w, http.StatusNotFound, "task not found")
		return
	}

	// Return the updated row
	s.GetTask(w, r, projectId, taskId)
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
