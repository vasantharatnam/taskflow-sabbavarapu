package api

import (
	"time"

	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/models"
)

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
