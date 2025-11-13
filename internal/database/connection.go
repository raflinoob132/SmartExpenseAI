package database

import (
	"log"
	"os"

	"SmartExpenseAI/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Migrate the schema
	DB.AutoMigrate(&models.Expense{})

	log.Println("Database connected successfully")
}

func GetExpensesByUserID(userID uint) ([]models.Expense, error) {
	var expenses []models.Expense
	result := DB.Where("user_id = ?", userID).Order("date DESC").Find(&expenses)
	return expenses, result.Error
}

func GetExpenseByID(userID uint, expenseID uint) (*models.Expense, error) {
	var expense models.Expense
	result := DB.Where("user_id = ? AND id = ?", userID, expenseID).First(&expense)
	if result.Error != nil {
		return nil, result.Error
	}
	return &expense, nil
}

func UpdateExpense(expense *models.Expense) error {
	result := DB.Save(expense)
	return result.Error
}

func DeleteExpense(userID uint, expenseID uint) error {
	result := DB.Where("user_id = ? AND id = ?", userID, expenseID).Delete(&models.Expense{})
	return result.Error
}