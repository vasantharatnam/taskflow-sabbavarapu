package repository

import (
	"context"
	"fmt"
	
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/models"
)


type TaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) CreateTask(ctx context.Context, task *models.Task) error {
	query := `
		INSERT INTO tasks (
			title, description, status, priority, project_id, assignee_id, creator_id, due_date
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx , query , task.Title , task.Description , task.Status , task.Priority , task.ProjectID , task.AssigneeID , task.CreatorID , task.DueDate).
	Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}


func (r *TaskRepository) GetTasksByProjectID(ctx context.Context , projectID string , status ,  assigneeID string)([]models.Task , error) {
	baseQuery := `
		SELECT id, title, description, status, priority, project_id, assignee_id, creator_id, due_date, created_at, updated_at
		FROM tasks
		WHERE project_id = $1
	`
	args := []any{projectID}
	argPos := 2

	if status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	if assigneeID != "" {
		baseQuery += fmt.Sprintf(" AND assignee_id = $%d", argPos)
		args = append(args, assigneeID)
		argPos++
	}

	baseQuery += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, baseQuery, args...)
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

	return tasks, nil

}

func (r *TaskRepository) GetTaskByID(ctx context.Context , taskID string) (*models.Task , error) {
	query := `
		SELECT id, title, description, status, priority, project_id, assignee_id, creator_id, due_date, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	var t models.Task

	err := r.db.QueryRow(ctx, query, taskID).Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.Priority,
		&t.ProjectID,
		&t.AssigneeID,
		&t.CreatorID,
		&t.DueDate,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *TaskRepository) UpdateTask(ctx context.Context , task *models.Task) error {
	query := `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			priority = $4,
			assignee_id = $5,
			due_date = $6,
			updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at
	`

	return r.db.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.AssigneeID,
		task.DueDate,
		task.ID,
	).Scan(&task.UpdatedAt)
}

func (r *TaskRepository) DeleteTask(ctx context.Context, taskID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, taskID)
	return err
}

func (r *TaskRepository) AssigneeExists(ctx context.Context, assigneeID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, assigneeID).Scan(&exists)
	return exists, err
}


// keep compiler happy if pgx imported but not used in future refactors
var _ = pgx.ErrNoRows