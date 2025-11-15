package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Login     string    `gorm:"uniqueIndex;not null" json:"login"`
	Password  string    `gorm:"not null" json:"-"`
	Name      string    `json:"name"`
	Bio       string    `json:"bio"`
	AvatarURL string    `json:"avatar_url"`
	Habits    []Habit   `gorm:"foreignKey:UserID" json:"habits,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Habit struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	UserID      uint         `gorm:"not null;index" json:"user_id"`
	User        User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Name        string       `gorm:"not null" json:"name"`
	Goal        string       `json:"goal"`
	IsPublic    bool         `gorm:"default:false" json:"is_public"`
	Completions []Completion `gorm:"foreignKey:HabitID" json:"completions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Completion struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	HabitID       uint      `gorm:"not null;index" json:"habit_id"`
	Habit         Habit     `gorm:"foreignKey:HabitID" json:"habit,omitempty"`
	DateCompleted time.Time `gorm:"type:date;not null" json:"date_completed"`
	CreatedAt     time.Time `json:"created_at"`
}

// Add unique constraint for habit + date
func (Completion) TableName() string {
	return "completions"
}

func MigrateDB(db *gorm.DB) error {
	err := db.AutoMigrate(&User{}, &Habit{}, &Completion{})
	if err != nil {
		return err
	}

	// Add unique constraint
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_habit_date ON completions(habit_id, date_completed)")

	return nil
}
