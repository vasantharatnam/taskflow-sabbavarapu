package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (name, email, password_hash) 
	           VALUES ($1, $2, $3) 
			   RETURNING id, created_at`

	return r.db.QueryRow(ctx, query, user.Name, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt)
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, created_at 
	           FROM users 
			   WHERE email = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, email).
		Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) IsUserExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT COUNT(1) FROM users WHERE email = $1`
	var count int
	err := r.db.QueryRow(ctx, query, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
