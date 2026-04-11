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

type ProjectHandler struct {
	projectRepo *repository.ProjectRepository
}

func NewProjectHandler(projectRepo *repository.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{projectRepo: projectRepo}
}

type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type ProjectDetailResponse struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description *string       `json:"description,omitempty"`
	OwnerID     string        `json:"owner_id"`
	CreatedAt   time.Time     `json:"created_at"`
	Tasks       []models.Task `json:"tasks"`
}


func (h *ProjectHandler) ListProjects (w http.ResponseWriter, r *http.Request) {
	 user, ok := auth.UserFromContext(r.Context())
	 if !ok {
		response.WriteError(w , http.StatusUnauthorized , "unauthorized")
		return
	 }

	 ctx, cancel  := context.WithTimeout(r.Context() , 5*time.Second)
	 defer cancel()

    projects , err := h.projectRepo.GetProjectsAccessedByUserID(ctx , user.UserID)
	if err != nil {
		response.WriteError(w , http.StatusInternalServerError , "internal server error")
		return
	}

	response.WriteJSON(w , http.StatusOK , map[string]any{
		"projects" : projects,
	})
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	fields := make(map[string]string)
	if strings.TrimSpace(req.Name) == "" {
		fields["name"] = "is required"
	}

	if len(fields) > 0 {
		response.WriteValidationError(w, fields)
		return
	}

	project := &models.Project{
		Name : req.Name,
		Description : req.Description,
		OwnerID : user.UserID,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.projectRepo.CreateProject(ctx, project); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to create project")
		return
	}

	response.WriteJSON(w , http.StatusCreated , project)
	
}

func (h *ProjectHandler) GetProjectByID(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
    
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return 
	}

	projectID := projectIDFromPath(r.URL.Path)
	if projectID == "" {
		response.WriteError(w , http.StatusNotFound , "project not found")
		return
	}

	ctx , cancel := context.WithTimeout(r.Context() , 5*time.Second)
	defer cancel()

	hasAccess, err := h.projectRepo.IsUserHasAccess(ctx , projectID , user.UserID)
	if err != nil {
       response.WriteError(w , http.StatusInternalServerError , "internal server error")
	   return
	}
	if !hasAccess {
		response.WriteError(w , http.StatusForbidden , "forbidden")
		return 
	}

	project , err := h.projectRepo.GetProjectByID(ctx , projectID)
	if err != nil {
		if errors.Is(err , pgx.ErrNoRows) {
			response.WriteError(w , http.StatusNotFound , "project not found")
			return
		}
		response.WriteError(w , http.StatusInternalServerError , "internal server error")
		return
	}

    tasks, err := h.projectRepo.GetTasksByProjectID(ctx , projectID)
	if err != nil {
		response.WriteError(w , http.StatusInternalServerError , "internal server error")
		return
	}
	
	response.WriteJSON(w , http.StatusOK , ProjectDetailResponse{
		ID: project.ID,
		Name: project.Name,
		Description: project.Description,
		OwnerID: project.OwnerID,
		CreatedAt: project.CreatedAt,
		Tasks: tasks,
	})
} 

func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID := projectIDFromPath(r.URL.Path)
	if projectID == "" {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}

	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	fields := make(map[string]string)
	if strings.TrimSpace(req.Name) == "" {
		fields["name"] = "is required"
	}
	if len(fields) > 0 {
		response.WriteValidationError(w, fields)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	project, err := h.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if project.OwnerID != user.UserID {
		response.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	project.Name = strings.TrimSpace(req.Name)
	project.Description = req.Description

	if err := h.projectRepo.UpdateProject(ctx, project); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to update project")
		return
	}

	response.WriteJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID := projectIDFromPath(r.URL.Path)
	if projectID == "" {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	project, err := h.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if project.OwnerID != user.UserID {
		response.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.projectRepo.DeleteProject(ctx, projectID); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to delete project")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func projectIDFromPath(path string) string {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		return ""
	}

	if parts[0] != "projects" {
		return ""
	}

	return parts[1]
}