package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"server-go/models"
	"server-go/repositories"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type UserService interface {
	Login(input models.LoginUser, w http.ResponseWriter) (string, error)
	Register(input models.User) (string, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	Logout(w http.ResponseWriter) error
}

type userService struct {
	userRepository repositories.UserRepository
}

// Login implements AuthService.
func (s *userService) Login(input models.LoginUser, w http.ResponseWriter) (string, error) {
	log.Printf("Login attempt for Email: %s", input.Email)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := s.userRepository.FindByEmail(ctx, input.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User not found for Email: %s", input.Email)
			return "", errors.New("user not found")
		}
		log.Printf("Database error while finding user: %v", err)
		return "", fmt.Errorf("database error: %v", err)
	}

	if user == nil {
		log.Printf("User is nil for Email: %s", input.Email)
		return "", errors.New("user not found")
	}

	log.Printf("User found for Email: %s", input.Email)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		log.Printf("Invalid credentials for Email: %s", input.Email)
		return "", errors.New("invalid credentials")
	}

	log.Printf("Password comparison successful for Email: %s", input.Email)

	token, err := generateJWT(user)
	if err != nil {
		log.Printf("Error generating JWT for Email: %s. Error: %v", input.Email, err)
		return "", fmt.Errorf("error generating token: %v", err)
	}

	log.Printf("JWT generated successfully for Email: %s", input.Email)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(72 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	log.Printf("Login successful for Email: %s", input.Email)
	return token, nil
}

func generateJWT(user *models.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":           "server-go",
		"sub":           user.Id,
		"iat":           now.Unix(),
		"exp":           now.Add(time.Hour * 24).Unix(),
		"nbf":           now.Unix(),
		"user_id":       user.Id,
		"user_Name":     user.Name,
		"user_LastName": user.LastName,
		"user_Email":    user.Email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(jwtSecret)
}

func (s *userService) Register(input models.User) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user already exists
	existingUser, err := s.userRepository.FindByEmail(ctx, input.Email)
	if err != nil {
		return "", err
	}
	if existingUser != nil {
		return "", errors.New("user already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("failed to hash password")
	}
	input.Password = string(hashedPassword)

	// Register the user
	registeredUser, err := s.userRepository.RegisterUser(ctx, input.Name, input.LastName, input.Email, input.Password)
	if err != nil {
		return "", err
	}

	// Generate JWT token (assuming you have a function to do this)
	token, err := generateJWT(registeredUser)
	if err != nil {
		return "", err
	}

	return token, nil
}

// update implements AuthService
func (s *userService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	// Hash the password before updating the user
	if user.Password != "" {
		hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		user.Password = string(hashPassword)
	}

	// Update the user in the repository
	updatedUser, err := s.userRepository.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Fetch the updated user from the repository to ensure we have the latest data
	updatedUser, err = s.userRepository.FindByID(ctx, user.Id)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

// Logout implements AuthService.
func (s *userService) Logout(w http.ResponseWriter) error {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// use the repo to generate the service
func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{userRepository: userRepo}
}
