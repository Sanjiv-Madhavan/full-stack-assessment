package repo

import (
	"context"
	"strings"
	"time"

	"full-stack-assesment/internal/apierrors"
	"full-stack-assesment/internal/helpers"
	repo "full-stack-assesment/internal/repo/task"
	"full-stack-assesment/internal/scheme"
	projectsSvc "full-stack-assesment/internal/service/projects"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

type TaskService struct {
	projectsService projectsSvc.ProjectsService
	repo            repo.SQLiteTaskRepo
}

func NewService(repo repo.SQLiteTaskRepo, projectsService projectsSvc.ProjectsService) *TaskService {
	return &TaskService{
		repo:            repo,
		projectsService: projectsService,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, newTask scheme.NewTask, projectID string) (*scheme.Task, error) {

	title := strings.TrimSpace(newTask.Title)
	if title == "" {
		return nil, apierrors.ErrorTaskTitleNotFound
	}
	if len(title) > 200 {
		return nil, apierrors.ErrTaskTitleTooLong
	}
	status := "TODO"
	if newTask.Status != nil {
		if norm, ok := helpers.NormalizeStatus(string(*newTask.Status)); ok {
			status = norm
		} else {
			return nil, apierrors.ErrorTaskStatusInvalid
		}
	}

	if err := s.projectsService.EnsureProjectExists(ctx, projectID); err != nil {
		return nil, err
	}

	id := uuid.New()
	now := time.Now().UTC()

	task := scheme.Task{
		Id:          types.UUID(id),
		ProjectId:   helpers.MustUUID(projectID),
		Title:       newTask.Title,
		Description: newTask.Description,
		Status:      scheme.TaskStatus(status),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *TaskService) GetTask(ctx context.Context, taskUUID string, projectUUID string) (*scheme.Task, error) {
	task, err := s.repo.Get(ctx, taskUUID, projectUUID)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) ListTasks(ctx context.Context, projectId string, params scheme.ListTasksParams) ([]scheme.Task, error) {

	if err := s.projectsService.EnsureProjectExists(ctx, projectId); err != nil {
		return []scheme.Task{}, err
	}

	var (
		where  = []string{"project_id = ?"}
		args   = []any{projectId}
		limit  = 50
		offset = 0
	)

	if params.Status != nil {
		if norm, ok := helpers.NormalizeStatus(string(*params.Status)); ok {
			where = append(where, "status = ?")
			args = append(args, norm)
		} else {
			return []scheme.Task{}, apierrors.ErrorTaskStatusInvalid
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

	tasks, err := s.repo.List(ctx, offset, limit, where, args)
	if err != nil {
		return []scheme.Task{}, err
	}

	return tasks, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, taskUUID string, projectUUID string) error {
	if err := s.projectsService.EnsureProjectExists(ctx, projectUUID); err != nil {
		return apierrors.ErrProjectNotFound
	}
	if err := s.repo.Delete(ctx, taskUUID, projectUUID); err != nil {
		return err
	}
	return nil
}

func (s *TaskService) UpdateTask(ctx context.Context, args []any, set []string) error {

	if err := s.repo.Update(ctx, args, set); err != nil {
		return err
	}
	return nil

}
