package controllers

import (
	"log"
	"net/http"
	"server-go/models"
	"server-go/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService services.UserService
}

func NewUserController(userService services.UserService) *UserController {
	return &UserController{userService: userService}
}

func (crtl *UserController) Login(c *gin.Context) {
	var loginData models.LoginUser

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	// Validate input
	if loginData.Email == "" || loginData.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}

	token, err := crtl.userService.Login(loginData, c.Writer)
	if err != nil {
		switch err.Error() {
		case "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case "invalid credentials":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "message": "Login successful"})
}

func (crtl *UserController) Register(c *gin.Context) {
	var registerData models.User

	if err := c.ShouldBindJSON(&registerData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data", "details": err.Error()})
		return
	}

	// Basic input validation
	if registerData.Name == "" || registerData.LastName == "" || registerData.Email == "" || registerData.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	token, err := crtl.userService.Register(registerData)
	if err != nil {
		log.Printf("Error in Register: %v", err)
		switch err.Error() {
		case "user already exists":
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		case "failed to hash password":
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register the user"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "token": token})
}

// handle the user session
func (crtl *UserController) Me(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userModel, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user model"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userID":   userModel.Id,
		"name":     userModel.Name,
		"lastName": userModel.LastName,
		"Email":    userModel.Email,
		"password": userModel.Password,
	})
}

func (ctrl *UserController) UpdateUser(c *gin.Context) {
	var user models.User
	// Obtain the user ID from the route
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	user.Id = id
	// Bind JSON data to the user model
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data", "details": err.Error()})
		return
	}
	// Call the UpdateUser method in the UserService
	updatedUser, err := ctrl.userService.UpdateUser(c.Request.Context(), &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated the user", "user": updatedUser})
}

func (crtl *UserController) Logout(c *gin.Context) {
	err := crtl.userService.Logout(c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
