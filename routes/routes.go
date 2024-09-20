package routes

import (
	"server-go/controllers"
	"server-go/middlewares"

	"github.com/gin-gonic/gin"
)

func SetUpRoutes(r *gin.Engine, userController *controllers.UserController) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello",
		})
	})
	r.POST("/login", userController.Login)
	r.POST("/register", userController.Register)
	r.PUT("/user/:id", middlewares.AuthMiddleware(), userController.UpdateUser)
	r.GET("/me", middlewares.AuthMiddleware(), userController.Me)
	r.POST("/logout", middlewares.AuthMiddleware(), userController.Logout)
}
