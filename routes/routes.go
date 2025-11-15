package routes

import (
	"github.com/gin-gonic/gin"
	"habit-tracker/controllers"
	"habit-tracker/middleware"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		// Public routes
		api.POST("/register", controllers.Register)
		api.POST("/login", controllers.Login)
		api.GET("/:login/habits", controllers.GetPublicProfile)

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired())
		{
			// User routes
			protected.GET("/me", controllers.GetCurrentUser)

			// Habit routes
			protected.POST("/habits", controllers.CreateHabit)
			protected.GET("/habits", controllers.GetUserHabits)
			protected.GET("/habits/:id", controllers.GetHabit)
			protected.PUT("/habits/:id", controllers.UpdateHabit)
			protected.DELETE("/habits/:id", controllers.DeleteHabit)

			// Completion routes
			protected.POST("/habits/:id/toggle", controllers.ToggleCompletion)
			protected.POST("/habits/:id/undo", controllers.UndoLastCompletion)
		}
	}
}
