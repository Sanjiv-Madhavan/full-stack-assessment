package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/oapi-codegen/runtime/types"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"full-stack-assesment/internal/apierrors"
	"full-stack-assesment/internal/helpers"
	"full-stack-assesment/internal/scheme"
	service "full-stack-assesment/internal/service/projects"
	taskservice "full-stack-assesment/internal/service/task"
)

var _ ServerInterface = (*Server)(nil)

type Server struct {
	projectsService service.ProjectsService
	tasksService    taskservice.TaskService
}

func NewServer(projectSvc service.ProjectsService, taskSvc taskservice.TaskService) *Server {
	return &Server{
		projectsService: projectSvc,
		tasksService:    taskSvc,
	}
}

func (s *Server) GetHealth(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, "OK")
}

func (s *Server) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body scheme.NewProject
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	project, err := s.projectsService.CreateProject(ctx, body)
	if err != nil {
		// If time permitted, I would have logged here. For now just focused on error handlers
		if err == apierrors.ErrProjectNameExists {
			helpers.WriteError(w, http.StatusConflict, "project name exists")
			return
		}
		if err == apierrors.ErrProjectNameTooLong || err == apierrors.ErrProjectNameRequired {
			helpers.WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		helpers.WriteJSON(w, http.StatusInternalServerError, err)
		return
	}

	created := scheme.Project{
		Id:        project.Id,
		Name:      project.Name,
		CreatedAt: project.CreatedAt,
		UpdatedAt: project.UpdatedAt,
	}
	helpers.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) ListProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projects, err := s.projectsService.ListProject(ctx)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "internal server error")
	}

	helpers.WriteJSON(w, http.StatusOK, projects)
}

func (s *Server) ListTasks(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, params scheme.ListTasksParams) {
	ctx := r.Context()

	tasks, err := s.tasksService.ListTasks(ctx, projectId.String(), params)
	if err != nil {
		if err == apierrors.ErrProjectNotFound {
			helpers.WriteError(w, http.StatusNotFound, "project not found")
			return
		}
		if err == apierrors.ErrorTaskStatusInvalid {
			helpers.WriteError(w, http.StatusBadRequest, "invalid status options")
			return
		}
		helpers.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	helpers.WriteJSON(w, http.StatusOK, tasks)
}

func (s *Server) CreateTask(w http.ResponseWriter, r *http.Request, projectUUID types.UUID) {
	ctx := r.Context()

	projectID := projectUUID.String()

	var body scheme.NewTask
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	task, err := s.tasksService.CreateTask(ctx, body, projectID)
	if err != nil {
		if err == apierrors.ErrTaskTitleTooLong || err == apierrors.ErrorTaskStatusInvalid || err == apierrors.ErrorTaskTitleNotFound {
			helpers.WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err == apierrors.ErrProjectNotFound {
			helpers.WriteError(w, http.StatusNotFound, "project not found")
			return
		}
		helpers.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, task)
}

func (s *Server) GetTask(w http.ResponseWriter, r *http.Request, projectUUID openapi_types.UUID, taskUUID openapi_types.UUID) {
	ctx := r.Context()

	projectUUID, ok := helpers.ParseUUIDParam(w, r, "projectId")
	if !ok {
		return
	}
	taskUUID, ok = helpers.ParseUUIDParam(w, r, "taskId")
	if !ok {
		return
	}

	task, err := s.tasksService.GetTask(ctx, taskUUID.String(), projectUUID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.WriteError(w, http.StatusBadRequest, "invalid task id")
			return
		}
		helpers.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	helpers.WriteJSON(w, http.StatusOK, task)
}

func (s *Server) DeleteTask(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, taskId openapi_types.UUID) {
	ctx := r.Context()
	if err := s.tasksService.DeleteTask(ctx, taskId.String(), projectId.String()); err != nil {
		if err == apierrors.ErrProjectNotFound {
			helpers.WriteError(w, http.StatusNotFound, "project not found")
			return
		}
		if err == apierrors.ErrorTaskTitleNotFound {
			helpers.WriteError(w, http.StatusBadRequest, "invalid task")
			return
		}
		helpers.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	helpers.WriteJSON(w, http.StatusNoContent, "Deletion: OK")
}

func (s *Server) UpdateTask(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, taskId openapi_types.UUID) {
	ctx := r.Context()

	if err := s.projectsService.EnsureProjectExists(ctx, projectId.String()); err != nil {
		if err == apierrors.ErrProjectNotFound {
			helpers.WriteError(w, http.StatusNotFound, "project not found")
			return
		}
		helpers.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

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

	if len(set) == 0 {
		s.GetTask(w, r, projectId, taskId)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	set = append(set, "updated_at = ?")
	args = append(args, now)
	args = append(args, taskId.String(), projectId.String())

	if err := s.tasksService.UpdateTask(ctx, args, set); err != nil {
		if err == apierrors.ErrorTaskTitleNotFound {
			helpers.WriteError(w, http.StatusBadRequest, "invalid task")
			return
		}
		helpers.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	s.GetTask(w, r, projectId, taskId)
}
