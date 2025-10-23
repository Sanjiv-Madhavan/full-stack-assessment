package apierrors

import "errors"

var (
	ErrProjectNameRequired = errors.New("project name is required")
	ErrProjectNameTooLong  = errors.New("project name is too long (max 128)")
	ErrProjectNameExists   = errors.New("project name already exists")
	ErrProjectNotFound     = errors.New("project not found")
	ErrorTaskTitleNotFound = errors.New("title is required")
	ErrTaskTitleTooLong    = errors.New("title too long (max 200)")
	ErrorTaskStatusInvalid = errors.New("invalid status; use TODO|IN_PROGRESS|DONE")
)
