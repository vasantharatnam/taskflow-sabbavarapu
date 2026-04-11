package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/models"
)

type ProjectWithTasks struct {
	Project models.Project `json:"project"`
	Tasks   []models.Task  `json:"tasks"`
}

type ProjectRepository struct {
	db *pgxpool.Pool
}

func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) CreateProject(ctx context.Context,  project *models.Project) error {
	 
	query := `INSERT INTO projects (name, description , owner_id) 
	           VALUES ($1, $2, $3) 
			   RETURNING id, created_at`

	return r.db.QueryRow(ctx , query , project.Name , project.Description , project.OwnerID).
		Scan(&project.ID, &project.CreatedAt)		
}

func (r *ProjectRepository) GetProjectsAccessedByUserID(ctx context.Context, userID string) ([]models.Project, error) {

	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1 OR t.assignee_id = $1 OR t.creator_id = $1
		ORDER BY p.created_at DESC
	`

	rows , err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
    
	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	return  projects , rows.Err()
}


func (r *ProjectRepository) GetProjectByID(ctx context.Context , projectID string) (*models.Project , error) {
	query := `
		SELECT id, name, description, owner_id, created_at
		FROM projects
		WHERE id = $1
	`

	var p models.Project
	err := r.db.QueryRow(ctx , query , projectID).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &p , nil
}

func (r *ProjectRepository) IsUserHasAccess(ctx context.Context , projectID string , userID string) (bool , error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM projects p
			LEFT JOIN tasks t ON t.project_id = p.id
			WHERE p.id = $1
			  AND (p.owner_id = $2 OR t.assignee_id = $2 OR t.creator_id = $2)
		)
	`

	var hasAccess bool
	err := r.db.QueryRow(ctx, query, projectID, userID).Scan(&hasAccess)
	if err != nil {
		return false, err
	}
	return hasAccess, nil
}


func (r *ProjectRepository) GetTasksByProjectID(ctx context.Context , projectID string) ([]models.Task , error) {
	query := `
		SELECT id, title, description, status, priority, project_id, assignee_id, creator_id, due_date, created_at, updated_at
		FROM tasks
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.ProjectID, &t.AssigneeID, &t.CreatorID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks , rows.Err()
}

func (r *ProjectRepository) UpdateProject(ctx context.Context , project *models.Project) error {
	query := `
		UPDATE projects
		SET name = $1, description = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, project.Name, project.Description, project.ID)
	return err
}


func (r *ProjectRepository) DeleteProject(ctx context.Context , projectID string) error {
	query := `
		DELETE FROM projects
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, projectID)
	return err
}