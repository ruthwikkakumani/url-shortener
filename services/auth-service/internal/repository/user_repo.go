package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruthwikkakumani/redirection-engine/services/auth-service/internal/model"
	"go.uber.org/zap"
)

type UserRepo struct {
	logger *zap.Logger
	db	   *pgxpool.Pool
}

func NewUserRepo(logger *zap.Logger, db *pgxpool.Pool) (*UserRepo){
	return &UserRepo{
		logger: logger,
		db: db,
	}
}

func (r *UserRepo) CreateUser(user *model.User) (error) {
	
	query := `
		INSERT INTO users (name, email, password)
		values ($1, $2, $3)
	`
	
	_, err := r.db.Exec(context.Background(), query, 
		user.Name, 
		user.Email, 
		user.Password)
	
	return err
}

func (r *UserRepo) GetUserByEmail(email string) (*model.User, error) {
	query := `
			SELECT id, name, email, password
			FROM users
			WHERE email = $1
	`
	
	var user model.User
	
	err := r.db.QueryRow(context.Background(), query, email).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (r *UserRepo) UpdatePassword(email, hashedPassword string) error {
	query := `
		UPDATE users
		SET password = $1
		WHERE email = $2
	`
	_, err := r.db.Exec(context.Background(), query, hashedPassword, email)
	return err
}