package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/auth"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/models"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/repository"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/response"	
)

type TaskHandler struct {
	taskRepo *repository.TaskRepository
	projectRepo *repository.ProjectRepository
}

func NewTaskHandler(taskRepo *repository.TaskRepository, projectRepo *repository.ProjectRepository) *TaskHandler {
	return &TaskHandler{
		taskRepo : taskRepo,
		projectRepo : projectRepo,
	}
}

type CreateTaskRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

func (h *TaskHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID := projectIDFromTaskListPath(r.URL.Path)
	if projectID == "" {
		response.WriteError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	status := r.URL.Query().Get("status")
	assigneeID := r.URL.Query().Get("assignee_id")

	if status != "" && !isValidStatus(status) {
		response.WriteValidationError(w, map[string]string{
			"status": "is invalid",
		})
		return
	}

	ctx , cancel := context.WithTimeout(r.Context() , 5*time.Second)
	defer cancel()

	// Check if user has access to the project
	hasAccess, err := h.projectRepo.IsUserHasAccess(ctx, projectID, user.UserID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !hasAccess {
		response.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	tasks, err := h.taskRepo.GetTasksByProjectID(ctx, projectID, status, assigneeID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"tasks": tasks,
	})
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID := projectIDFromTaskListPath(r.URL.Path)
	if projectID == "" {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	fields := validateCreateTaskRequest(req)
	if len(fields) > 0 {
		response.WriteValidationError(w, fields)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	hasAccess, err := h.projectRepo.IsUserHasAccess(ctx, projectID, user.UserID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !hasAccess {
		response.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	if req.AssigneeID != nil && strings.TrimSpace(*req.AssigneeID) != "" {
		exists, err := h.taskRepo.AssigneeExists(ctx, strings.TrimSpace(*req.AssigneeID))
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if !exists {
			response.WriteValidationError(w, map[string]string{
				"assignee_id": "does not exist",
			})
			return
		}
	}

	dueDate, err := parseOptionalDate(req.DueDate)
	if err != nil {
		response.WriteValidationError(w, map[string]string{
			"due_date": "must be in YYYY-MM-DD format",
		})
		return
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "todo"
	}

	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		Priority:    req.Priority,
		ProjectID:   projectID,
		AssigneeID:  normalizeOptionalString(req.AssigneeID),
		CreatorID:   user.UserID,
		DueDate:     dueDate,
	}

	if err := h.taskRepo.CreateTask(ctx, task); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	response.WriteJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	taskID := taskIDFromPath(r.URL.Path)
	if taskID == "" {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	fields := validateUpdateTaskRequest(req)
	if len(fields) > 0 {
		response.WriteValidationError(w, fields)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	task, err := h.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	project, err := h.projectRepo.GetProjectByID(ctx, task.ProjectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if project.OwnerID != user.UserID && task.CreatorID != user.UserID {
		response.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	if req.AssigneeID != nil && strings.TrimSpace(*req.AssigneeID) != "" {
		exists, err := h.taskRepo.AssigneeExists(ctx, strings.TrimSpace(*req.AssigneeID))
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if !exists {
			response.WriteValidationError(w, map[string]string{
				"assignee_id": "does not exist",
			})
			return
		}
	}

	dueDate, err := parseOptionalDate(req.DueDate)
	if err != nil {
		response.WriteValidationError(w, map[string]string{
			"due_date": "must be in YYYY-MM-DD format",
		})
		return
	}

	task.Title = req.Title
	task.Description = req.Description
	task.Status = req.Status
	task.Priority = req.Priority
	task.AssigneeID = normalizeOptionalString(req.AssigneeID)
	task.DueDate = dueDate

	if err := h.taskRepo.UpdateTask(ctx, task); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to update task")
		return
	}

	response.WriteJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	taskID := taskIDFromPath(r.URL.Path)
	if taskID == "" {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	task, err := h.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	project, err := h.projectRepo.GetProjectByID(ctx, task.ProjectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if project.OwnerID != user.UserID && task.CreatorID != user.UserID {
		response.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.taskRepo.DeleteTask(ctx, taskID); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to delete task")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validateCreateTaskRequest(req CreateTaskRequest) map[string]string {
	fields := make(map[string]string)

	if strings.TrimSpace(req.Title) == "" {
		fields["title"] = "is required"
	}

	status := strings.TrimSpace(req.Status)
	if status != "" && !isValidStatus(status) {
		fields["status"] = "is invalid"
	}

	if strings.TrimSpace(req.Priority) == "" {
		fields["priority"] = "is required"
	} else if !isValidPriority(req.Priority) {
		fields["priority"] = "is invalid"
	}

	return fields
}

func validateUpdateTaskRequest(req UpdateTaskRequest) map[string]string {
	fields := make(map[string]string)

	if strings.TrimSpace(req.Title) == "" {
		fields["title"] = "is required"
	}

	if !isValidStatus(req.Status) {
		fields["status"] = "is invalid"
	}

	if !isValidPriority(req.Priority) {
		fields["priority"] = "is invalid"
	}

	return fields
}

func isValidStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "todo", "in_progress", "done":
		return true
	default:
		return false
	}
}

func isValidPriority(priority string) bool {
	switch strings.TrimSpace(priority) {
	case "low", "medium", "high":
		return true
	default:
		return false
	}
}

func parseOptionalDate(dateStr *string) (*time.Time, error) {
	if dateStr == nil || strings.TrimSpace(*dateStr) == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*dateStr))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func normalizeOptionalString(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func projectIDFromTaskListPath(path string) string {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 3 {
		return ""
	}

	if parts[0] != "projects" || parts[2] != "tasks" {
		return ""
	}

	return parts[1]
}

func taskIDFromPath(path string) string {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		return ""
	}

	if parts[0] != "tasks" {
		return ""
	}

	return parts[1]
}