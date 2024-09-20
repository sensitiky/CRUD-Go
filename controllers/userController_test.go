package controllers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"server-go/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocking the UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Login(loginData models.LoginUser, writer http.ResponseWriter) (string, error) {
	args := m.Called(loginData, writer)
	return args.String(0), args.Error(1)
}

func (m *MockUserService) Register(user models.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Logout(writer http.ResponseWriter) error {
	args := m.Called(writer)
	return args.Error(0)
}

func TestLogin(t *testing.T) {
	mockUserService := new(MockUserService)
	controller := NewUserController(mockUserService)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/login", controller.Login)

	t.Run("successful login", func(t *testing.T) {
		loginData := models.LoginUser{Email: "john@example.com", Password: "password"}
		mockUserService.On("Login", loginData).Return("token123", nil)

		body := bytes.NewBufferString(`{"email":"john@example.com","password":"password"}`)
		req, _ := http.NewRequest(http.MethodPost, "/login", body)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), "Login successful")
	})

	t.Run("invalid input format", func(t *testing.T) {
		body := bytes.NewBufferString(`{"email":"john@example.com"}`)
		req, _ := http.NewRequest(http.MethodPost, "/login", body)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "Email and password are required")
	})
}

func TestRegister(t *testing.T) {
	mockUserService := new(MockUserService)
	controller := NewUserController(mockUserService)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/register", controller.Register)

	t.Run("successful registration", func(t *testing.T) {
		registerData := models.User{Name: "John", LastName: "Doe", Email: "john@example.com", Password: "password"}
		mockUserService.On("Register", registerData).Return("token123", nil)

		body := bytes.NewBufferString(`{"name":"John","lastName":"Doe","email":"john@example.com","password":"password"}`)
		req, _ := http.NewRequest(http.MethodPost, "/register", body)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)
		assert.Contains(t, resp.Body.String(), "User registered successfully")
	})

	t.Run("invalid input data", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"John"}`)
		req, _ := http.NewRequest(http.MethodPost, "/register", body)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "All fields are required")
	})
}

func TestMe(t *testing.T) {
	controller := NewUserController(nil)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/me", controller.Me)

	t.Run("unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/me", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Contains(t, resp.Body.String(), "Unauthorized")
	})

	t.Run("successful retrieval", func(t *testing.T) {
		router.Use(func(c *gin.Context) {
			c.Set("user", &models.User{Id: 1, Name: "John", LastName: "Doe", Email: "john@example.com", Password: "password"})
			c.Next()
		})

		req, _ := http.NewRequest(http.MethodGet, "/me", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), "John")
	})
}

func TestUpdateUser(t *testing.T) {
	mockUserService := new(MockUserService)
	controller := NewUserController(mockUserService)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PUT("/users/:id", controller.UpdateUser)

	t.Run("successful update", func(t *testing.T) {
		user := models.User{Id: 1, Name: "John", LastName: "Doe", Email: "john@example.com", Password: "password"}
		updatedUser := models.User{Id: 1, Name: "John", LastName: "Doe", Email: "john@example.com", Password: "newpassword"}
		mockUserService.On("UpdateUser", mock.Anything, &user).Return(&updatedUser, nil)

		body := bytes.NewBufferString(`{"name":"John","lastName":"Doe","email":"john@example.com","password":"password"}`)
		req, _ := http.NewRequest(http.MethodPut, "/users/1", body)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), "Successfully updated the user")
	})

	t.Run("invalid user ID", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"John","lastName":"Doe","email":"john@example.com","password":"password"}`)
		req, _ := http.NewRequest(http.MethodPut, "/users/abc", body)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "Invalid user ID")
	})
}

func TestLogout(t *testing.T) {
	mockUserService := new(MockUserService)
	controller := NewUserController(mockUserService)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/logout", controller.Logout)

	t.Run("successful logout", func(t *testing.T) {
		mockUserService.On("Logout", mock.Anything).Return(nil)

		req, _ := http.NewRequest(http.MethodPost, "/logout", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), "Successfully logged out")
	})

	t.Run("failed logout", func(t *testing.T) {
		mockUserService.On("Logout", mock.Anything).Return(assert.AnError)

		req, _ := http.NewRequest(http.MethodPost, "/logout", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Contains(t, resp.Body.String(), "Failed to logout")
	})
}
