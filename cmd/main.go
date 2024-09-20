package main

import (
	"log"
	"server-go/config"
	"server-go/controllers"
	"server-go/repositories"
	"server-go/routes"
	"server-go/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting the server")

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(config.CORSmiddleware())
	DB, err := config.DatabaseConnection()
	if err != nil {
		log.Fatalf("Could not connect to the database")
	}
	userRepo := repositories.NewUserRepository(DB)
	userService := services.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	routes.SetUpRoutes(r, userController)

	r.Run(":4000")
}
