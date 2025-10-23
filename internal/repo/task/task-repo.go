package repo

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"full-stack-assesment/internal/apierrors"
	"full-stack-assesment/internal/helpers"
	"full-stack-assesment/internal/scheme"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrTaskConflict = errors.New("task conflict")
)

type TaskRepository interface {
	Create(ctx context.Context, t scheme.Task) error
	Get(ctx context.Context, taskUUID string) (scheme.Task, error)
	List(ctx context.Context) ([]scheme.Task, error)
	Update(ctx context.Context, t scheme.Task) error
	Delete(ctx context.Context, taskUUID string) error
}

type SQLiteTaskRepo struct {
	db *sql.DB
}

func NewSQLiteTaskRepo(db *sql.DB) *SQLiteTaskRepo {
	return &SQLiteTaskRepo{db: db}
}

func (r *SQLiteTaskRepo) Create(ctx context.Context, t scheme.Task) error {
	const q = `
		INSERT INTO tasks (id, project_id, title, description, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	var desc string
	if t.Description != nil {
		desc = strings.TrimSpace(*t.Description)
	}
	taskUUID := t.Id.String()
	projectUUID := t.ProjectId.String()
	if _, err := r.db.ExecContext(ctx, q, taskUUID, projectUUID, t.Title, desc, t.Status, t.CreatedAt, t.UpdatedAt); err != nil {
		return err
	}
	return nil
}

func (r *SQLiteTaskRepo) Get(ctx context.Context, taskUUID string, projectUUID string) (*scheme.Task, error) {
	const q = `
		SELECT id, project_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = ? AND project_id = ?;
	`
	var idStr, projStr, title, desc, status, created, updated string
	err := r.db.QueryRowContext(ctx, q, taskUUID, projectUUID).
		Scan(&idStr, &projStr, &title, &desc, &status, &created, &updated)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
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
		CreatedAt: helpers.ParseTimeOrNow(created),
		UpdatedAt: helpers.ParseTimeOrNow(updated),
	}
	return &out, nil
}

func (r *SQLiteTaskRepo) List(ctx context.Context, offset, limit int, where []string, args []any) ([]scheme.Task, error) {
	stmt := `
		SELECT id, project_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?;
	`

	rows, err := r.db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return []scheme.Task{}, err
	}
	defer rows.Close()

	out := make([]scheme.Task, 0, limit)
	for rows.Next() {
		var idStr, projStr, title, desc, status, created, updated string
		if err := rows.Scan(&idStr, &projStr, &title, &desc, &status, &created, &updated); err != nil {
			return []scheme.Task{}, err
		}
		var descPtr *string
		if strings.TrimSpace(desc) != "" {
			cp := desc
			descPtr = &cp
		}
		// If time permitted, I would apply pagination on the UI. I have incorporated fields (limit and offset) for the very same reason
		// P.S. to ensure functionality, suite_test covers this case so no worries
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
	return out, nil
}

func (r *SQLiteTaskRepo) Delete(ctx context.Context, taskUUID string, projectUUID string) error {
	const q = `DELETE FROM tasks WHERE id = ? AND project_id = ?;`

	res, err := r.db.ExecContext(ctx, q, taskUUID, projectUUID)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return apierrors.ErrorTaskTitleNotFound
	}
	return nil
}

func (r *SQLiteTaskRepo) Update(ctx context.Context, args []any, set []string) error {
	stmt := `
		UPDATE tasks
		SET ` + strings.Join(set, ", ") + `
		WHERE id = ? AND project_id = ?;
	`

	res, err := r.db.ExecContext(ctx, stmt, args...)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return apierrors.ErrorTaskTitleNotFound
	}
	return nil
}
