package main

import (
	"habit-tracker/config"
	"habit-tracker/models"
	"habit-tracker/routes"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	godotenv.Load()

	// Connect to database
	config.ConnectDatabase()

	// Auto migrate models
	if err := models.MigrateDB(config.DB); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Setup Gin router
	r := gin.Default()

	// Enable CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Setup API routes FIRST
	routes.SetupRoutes(r)

	// Serve static files LAST
	r.Static("/static", "./frontend")
	r.GET("/", func(c *gin.Context) {
		c.File("./frontend/index.html")
	})

	// Start server
	log.Println("Server starting on :8080")
	log.Println("API available at: http://localhost:8080/api/v1/")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
