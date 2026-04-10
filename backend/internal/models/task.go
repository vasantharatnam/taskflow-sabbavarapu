

package models


import "time"

type Task struct {
	ID          string     `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description *string    `json:"description,omitempty" db:"description"`
	Status      string     `json:"status" db:"status"`
	Priority    string     `json:"priority" db:"priority"`
	ProjectID   string     `json:"project_id" db:"project_id"`
	AssigneeID  *string    `json:"assignee_id,omitempty" db:"assignee_id"`
	CreatorID   string     `json:"creator_id" db:"creator_id"`
	DueDate     *time.Time `json:"due_date,omitempty" db:"due_date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}