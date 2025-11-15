package controllers

import (
	"habit-tracker/config"
	"habit-tracker/models"
	"habit-tracker/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Auth Controllers
func Register(c *gin.Context) {
	var input struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Name     string `json:"name" binding:"required"`
		Bio      string `json:"bio"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Login:    input.Login,
		Password: string(hashedPassword),
		Name:     input.Name,
		Bio:      input.Bio,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login already exists"})
		return
	}

	token, _ := utils.GenerateToken(user.ID, user.Login)

	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

func Login(c *gin.Context) {
	var input struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("login = ?", input.Login).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, _ := utils.GenerateToken(user.ID, user.Login)

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

func GetCurrentUser(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Public Profile
func GetPublicProfile(c *gin.Context) {
	login := c.Param("login")

	var user models.User
	if err := config.DB.Where("login = ?", login).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var habits []models.Habit
	config.DB.Where("user_id = ? AND is_public = ?", user.ID, true).
		Preload("Completions").
		Find(&habits)

	c.JSON(http.StatusOK, gin.H{
		"user":   user,
		"habits": habits,
	})
}

// Habit CRUD
func CreateHabit(c *gin.Context) {
	userID := c.GetUint("user_id")

	var input struct {
		Name     string `json:"name" binding:"required"`
		Goal     string `json:"goal"`
		IsPublic bool   `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	habit := models.Habit{
		UserID:   userID,
		Name:     input.Name,
		Goal:     input.Goal,
		IsPublic: input.IsPublic,
	}

	if err := config.DB.Create(&habit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create habit"})
		return
	}

	c.JSON(http.StatusCreated, habit)
}

func GetUserHabits(c *gin.Context) {
	userID := c.GetUint("user_id")

	var habits []models.Habit
	config.DB.Where("user_id = ?", userID).
		Preload("Completions").
		Find(&habits)

	c.JSON(http.StatusOK, habits)
}

func GetHabit(c *gin.Context) {
	userID := c.GetUint("user_id")
	habitID := c.Param("id")

	var habit models.Habit
	if err := config.DB.Where("id = ? AND user_id = ?", habitID, userID).
		Preload("Completions").
		First(&habit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
		return
	}

	c.JSON(http.StatusOK, habit)
}

func UpdateHabit(c *gin.Context) {
	userID := c.GetUint("user_id")
	habitID := c.Param("id")

	var habit models.Habit
	if err := config.DB.Where("id = ? AND user_id = ?", habitID, userID).First(&habit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
		return
	}

	var input struct {
		Name     string `json:"name"`
		Goal     string `json:"goal"`
		IsPublic *bool  `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name != "" {
		habit.Name = input.Name
	}
	if input.Goal != "" {
		habit.Goal = input.Goal
	}
	if input.IsPublic != nil {
		habit.IsPublic = *input.IsPublic
	}

	config.DB.Save(&habit)

	c.JSON(http.StatusOK, habit)
}

func DeleteHabit(c *gin.Context) {
	userID := c.GetUint("user_id")
	habitID := c.Param("id")

	result := config.DB.Where("id = ? AND user_id = ?", habitID, userID).Delete(&models.Habit{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Habit deleted"})
}

// Completion Management
func ToggleCompletion(c *gin.Context) {
	userID := c.GetUint("user_id")
	habitID := c.Param("id")

	var input struct {
		Date string `json:"date"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	var habit models.Habit
	if err := config.DB.Where("id = ? AND user_id = ?", habitID, userID).First(&habit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
		return
	}

	var completion models.Completion
	result := config.DB.Where("habit_id = ? AND date_completed = ?", habitID, date).First(&completion)

	if result.Error == nil {
		config.DB.Delete(&completion)
		c.JSON(http.StatusOK, gin.H{"message": "Completion removed", "completed": false})
	} else {
		newCompletion := models.Completion{
			HabitID:       habit.ID,
			DateCompleted: date,
		}
		config.DB.Create(&newCompletion)
		c.JSON(http.StatusOK, gin.H{"message": "Completion added", "completed": true})
	}
}

func UndoLastCompletion(c *gin.Context) {
	userID := c.GetUint("user_id")
	habitID := c.Param("id")

	var habit models.Habit
	if err := config.DB.Where("id = ? AND user_id = ?", habitID, userID).First(&habit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
		return
	}

	var completion models.Completion
	if err := config.DB.Where("habit_id = ?", habitID).
		Order("date_completed DESC").
		First(&completion).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No completions to undo"})
		return
	}

	config.DB.Delete(&completion)

	c.JSON(http.StatusOK, gin.H{"message": "Last completion undone"})
}
