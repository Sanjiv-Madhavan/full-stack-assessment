// package api

// import (
// 	"context"
// 	"sync"
// 	"time"

// 	"full-stack-assesment/internal/scheme"

// 	"github.com/google/uuid"
// 	"github.com/oapi-codegen/runtime/types"
// )

// var _ StrictServerInterface = (*Server)(nil)

// type Server struct {
// 	mu    sync.RWMutex
// 	todos map[string]scheme.Todo
// }

// func NewServer() *Server {
// 	return &Server{
// 		todos: make(map[string]scheme.Todo),
// 	}
// }

// func (s *Server) GetHealth(ctx context.Context, _ GetHealthRequestObject) (GetHealthResponseObject, error) {
// 	return GetHealth200JSONResponse{Status: scheme.Ok}, nil
// }

// func (s *Server) GetTodos(ctx context.Context, _ GetTodosRequestObject) (GetTodosResponseObject, error) {
// 	s.mu.RLock()
// 	defer s.mu.RUnlock()
// 	out := make([]scheme.Todo, 0, len(s.todos))
// 	for _, t := range s.todos {
// 		out = append(out, t)
// 	}
// 	return GetTodos200JSONResponse(out), nil
// }

// func (s *Server) PostTodos(ctx context.Context, req PostTodosRequestObject) (PostTodosResponseObject, error) {
// 	if req.Body == nil || req.Body.Title == "" {
// 		return PostTodos400JSONResponse(scheme.Error{Code: 400, Message: "title is required"}), nil
// 	}
// 	u := uuid.New() // uuid.UUID (not string)
// 	todo := scheme.Todo{
// 		Id:        types.UUID(u), // 👈 match the generated field type
// 		Title:     req.Body.Title,
// 		Completed: false,
// 		CreatedAt: time.Now().UTC(),
// 	}
// 	s.mu.Lock()
// 	s.todos[u.String()] = todo
// 	s.mu.Unlock()
// 	return PostTodos201JSONResponse(todo), nil
// }

// func (s *Server) PutTodosId(ctx context.Context, req PutTodosIdRequestObject) (PutTodosIdResponseObject, error) {
// 	id := req.Id
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	cur, ok := s.todos[id.String()]
// 	if !ok {
// 		return PutTodosId404JSONResponse(scheme.Error{Code: 404, Message: "todo not found"}), nil
// 	}
// 	if req.Body != nil {
// 		if req.Body.Title != nil {
// 			cur.Title = *req.Body.Title
// 		}
// 		if req.Body.Completed != nil {
// 			cur.Completed = *req.Body.Completed
// 		}
// 	}
// 	s.todos[id.String()] = cur
// 	return PutTodosId200JSONResponse(cur), nil
// }

// func (s *Server) DeleteTodosId(ctx context.Context, req DeleteTodosIdRequestObject) (DeleteTodosIdResponseObject, error) {
// 	id := req.Id
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	if _, ok := s.todos[id.String()]; !ok {
// 		return DeleteTodosId404JSONResponse(scheme.Error{Code: 404, Message: "todo not found"}), nil
// 	}
// 	delete(s.todos, id.String())
// 	return DeleteTodosId204Response{}, nil
// }

package api

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"

	"full-stack-assesment/internal/scheme"
)

var _ StrictServerInterface = (*Server)(nil)

type Server struct {
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}

func (s *Server) GetHealth(ctx context.Context, _ GetHealthRequestObject) (GetHealthResponseObject, error) {
	return GetHealth200JSONResponse{Status: scheme.Ok}, nil
}

func (s *Server) GetTodos(ctx context.Context, _ GetTodosRequestObject) (GetTodosResponseObject, error) {
	const q = `SELECT id, title, completed, created_at FROM todos ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]scheme.Todo, 0, 16)
	for rows.Next() {
		var id, title string
		var completedInt int
		var createdAt string
		if err := rows.Scan(&id, &title, &completedInt, &createdAt); err != nil {
			return nil, err
		}
		out = append(out, scheme.Todo{
			Id:        parseUUIDOrZero(id),
			Title:     title,
			Completed: completedInt == 1,
			CreatedAt: mustParseTime(createdAt),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return GetTodos200JSONResponse(out), nil
}

func (s *Server) PostTodos(ctx context.Context, req PostTodosRequestObject) (PostTodosResponseObject, error) {
	if req.Body == nil || req.Body.Title == "" {
		return PostTodos400JSONResponse(scheme.Error{Code: 400, Message: "title is required"}), nil
	}
	id := uuid.New()
	now := time.Now().UTC()

	const q = `INSERT INTO todos(id, title, completed, created_at) VALUES(?, ?, 0, ?)`
	if _, err := s.db.ExecContext(ctx, q, id, req.Body.Title, now.Format(time.RFC3339Nano)); err != nil {
		return nil, err
	}

	return PostTodos201JSONResponse(scheme.Todo{
		Id:        id,
		Title:     req.Body.Title,
		Completed: false,
		CreatedAt: now,
	}), nil
}

func (s *Server) PutTodosId(ctx context.Context, req PutTodosIdRequestObject) (PutTodosIdResponseObject, error) {
	// fetch current
	const sel = `SELECT id, title, completed, created_at FROM todos WHERE id = ?`
	var id, title string
	var completedInt int
	var createdAt string
	err := s.db.QueryRowContext(ctx, sel, req.Id).Scan(&id, &title, &completedInt, &createdAt)
	if err == sql.ErrNoRows {
		return PutTodosId404JSONResponse(scheme.Error{Code: 404, Message: "todo not found"}), nil
	}
	if err != nil {
		return nil, err
	}

	// apply patch
	if req.Body != nil {
		if req.Body.Title != nil {
			title = *req.Body.Title
		}
		if req.Body.Completed != nil {
			if *req.Body.Completed {
				completedInt = 1
			} else {
				completedInt = 0
			}
		}
	}

	const upd = `UPDATE todos SET title = ?, completed = ? WHERE id = ?`
	if _, err := s.db.ExecContext(ctx, upd, title, completedInt, id); err != nil {
		return nil, err
	}

	return PutTodosId200JSONResponse(scheme.Todo{
		Id:        parseUUIDOrZero(id),
		Title:     title,
		Completed: completedInt == 1,
		CreatedAt: mustParseTime(createdAt),
	}), nil
}

func (s *Server) DeleteTodosId(ctx context.Context, req DeleteTodosIdRequestObject) (DeleteTodosIdResponseObject, error) {
	const del = `DELETE FROM todos WHERE id = ?`
	res, err := s.db.ExecContext(ctx, del, req.Id)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return DeleteTodosId404JSONResponse(scheme.Error{Code: 404, Message: "todo not found"}), nil
	}
	return DeleteTodosId204Response{}, nil
}

// helpers

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		// if someone later inserts not-ISO timestamps, fall back to now
		return time.Now().UTC()
	}
	return t
}

func parseUUIDOrZero(s string) types.UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil // still a valid UUID value; decide if you want to 500 instead
	}
	return u
}
