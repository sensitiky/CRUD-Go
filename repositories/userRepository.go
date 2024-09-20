package repositories

import (
	"context"
	"database/sql"
	"log"
	"server-go/models"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id int) (*models.User, error)
	RegisterUser(ctx context.Context, name string, lastName string, email string, password string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
}

type userRepositoryImpl struct {
	DB *sql.DB
}

func NewUserRepository(DB *sql.DB) UserRepository {
	return &userRepositoryImpl{DB: DB}
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, id int) (*models.User, error) {
	query := "SELECT id FROM users WHERE id =$1"
	user := &models.User{}

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&user.Id,
		&user.Name,
		&user.LastName,
		&user.Email,
		&user.Password,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("User not found")
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := "SELECT id, name, lastName, email, password FROM users WHERE email = $1"
	user := &models.User{}

	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&user.Id,
		&user.Name,
		&user.LastName,
		&user.Email,
		&user.Password,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("User not found")
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *userRepositoryImpl) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	query := `UPDATE users SET name = $1, lastName = $2, email = $3, password = $4 WHERE id = $5 RETURNING id, name, lastName, email, password`

	err := r.DB.QueryRowContext(ctx, query, user.Name, user.LastName, user.Email, user.Password, user.Id).Scan(
		&user.Id,
		&user.Name,
		&user.LastName,
		&user.Email,
		&user.Password,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("User not found")
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *userRepositoryImpl) RegisterUser(ctx context.Context, name string, lastName string, email string, password string) (*models.User, error) {
	query := `INSERT INTO users (name, lastName, email, password)
        VALUES ($1, $2, $3, $4)
        RETURNING id, name, lastName, email, password;`

	user := &models.User{
		Name:     name,
		LastName: lastName,
		Email:    email,
		Password: password,
	}

	err := r.DB.QueryRowContext(ctx, query, name, lastName, email, password).Scan(
		&user.Id,
		&user.Name,
		&user.LastName,
		&user.Email,
		&user.Password,
	)

	if err != nil {
		log.Printf("Error registering user: %v", err)
		return nil, err
	}

	log.Printf("User registered successfully: %+v", user)
	return user, nil
}
