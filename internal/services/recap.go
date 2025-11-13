package services

import (
	"fmt"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/go-co-op/gocron"

	"SmartExpenseAI/internal/database"
	"SmartExpenseAI/internal/models"
)

// GenerateWeeklyRecap generates a weekly recap of expenses and sends it to the specified chat
func GenerateWeeklyRecap(bot *tgbotapi.BotAPI, chatID int64) {
	// Calculate the date 7 days ago
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	// Query expenses from the last 7 days for this user
	var expenses []models.Expense
	result := database.DB.Where("user_id = ? AND date >= ?", uint(chatID), sevenDaysAgo).Find(&expenses)
	if result.Error != nil {
		log.Printf("Error fetching expenses: %v", result.Error)
		return
	}

	if len(expenses) == 0 {
		// No expenses found for the period
		msg := tgbotapi.NewMessage(chatID, "Tidak ada pengeluaran dalam 7 hari terakhir.")
		bot.Send(msg)
		return
	}

	// Group total per category
	categoryTotals := make(map[string]float64)
	totalAmount := 0.0

	for _, expense := range expenses {
		categoryTotals[expense.Category] += expense.Amount
		totalAmount += expense.Amount
	}

	// Format the recap message
	recapText := "🧾 Rekap Mingguan:\n"
	
	// Add each category with its total
	for category, amount := range categoryTotals {
		recapText += fmt.Sprintf("- %s: Rp%s\n", category, formatCurrency(amount))
	}
	
	// Add total amount
	recapText += fmt.Sprintf("Total: Rp%s", formatCurrency(totalAmount))

	// Send the message to the chat
	msg := tgbotapi.NewMessage(chatID, recapText)
	bot.Send(msg)
}

// GenerateMonthlyRecap generates a 30-day recap of expenses and sends it to the specified chat
func GenerateMonthlyRecap(bot *tgbotapi.BotAPI, chatID int64) {
	// Calculate the date 30 days ago
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	// Query expenses from the last 30 days for this user
	var expenses []models.Expense
	result := database.DB.Where("user_id = ? AND date >= ?", uint(chatID), thirtyDaysAgo).Order("date DESC").Find(&expenses)
	if result.Error != nil {
		log.Printf("Error fetching expenses: %v", result.Error)
		return
	}

	if len(expenses) == 0 {
		// No expenses found for the period
		msg := tgbotapi.NewMessage(chatID, "Tidak ada pengeluaran dalam 30 hari terakhir.")
		bot.Send(msg)
		return
	}

	// Group expenses by month
	monthlyExpenses := make(map[string][]models.Expense)
	for _, expense := range expenses {
		monthKey := expense.Date.Format("January 2006") // Format: "November 2025"
		monthlyExpenses[monthKey] = append(monthlyExpenses[monthKey], expense)
	}

	// Format the recap message
	recapText := "🧾 Rekap Pengeluaran 30 Hari:\n\n"

	// Sort months and display expenses
	// Get sorted list of months
	var months []string
	for month := range monthlyExpenses {
		months = append(months, month)
	}
	// Sort in reverse chronological order (newest first)
	for i := 0; i < len(months); i++ {
		for j := i + 1; j < len(months); j++ {
			// Convert month strings back to time for comparison
			time1, _ := time.Parse("January 2006", months[i])
			time2, _ := time.Parse("January 2006", months[j])
			if time1.Before(time2) {
				months[i], months[j] = months[j], months[i]
			}
		}
	}

	for _, month := range months {
		monthExpenses := monthlyExpenses[month]
		monthTotal := 0.0

		for _, expense := range monthExpenses {
			monthTotal += expense.Amount
		}

		recapText += fmt.Sprintf("*%s (Total: Rp%s)*\n", month, formatCurrency(monthTotal))
		for _, expense := range monthExpenses {
			recapText += fmt.Sprintf("• %s: Rp%s (%s)\n",
				expense.Date.Format("2 Jan"),
				formatCurrency(expense.Amount),
				expense.Description)
		}
		recapText += "\n"
	}

	// Add total amount for the whole period
	totalAmount := 0.0
	for _, expense := range expenses {
		totalAmount += expense.Amount
	}
	recapText += fmt.Sprintf("*Total 30 Hari Terakhir: Rp%s*", formatCurrency(totalAmount))

	// Send the message to the chat
	msg := tgbotapi.NewMessage(chatID, recapText)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

// Helper function to format currency with thousands separator
func formatCurrency(amount float64) string {
	// Convert to integer to avoid decimal places
	amt := int64(amount)
	
	// Convert to string
	s := strconv.FormatInt(amt, 10)
	
	// Add thousands separators from right to left
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	
	return s
}

// ScheduleWeeklyRecap schedules the weekly recap to run automatically
func ScheduleWeeklyRecap(bot *tgbotapi.BotAPI, chatID int64) {
	// Create a new scheduler
	scheduler := gocron.NewScheduler(time.UTC)

	// Schedule the function to run once a week on Sundays at 8:00 AM
	_, err := scheduler.Cron("0 8 * * 0").Do(func() {
		GenerateWeeklyRecap(bot, chatID)
	})

	if err != nil {
		log.Printf("Error scheduling weekly recap: %v", err)
		return
	}

	// Start the scheduler
	scheduler.StartAsync()
}

// CallWeeklyRecapForUser calls the weekly recap function for a specific user
func CallWeeklyRecapForUser(bot *tgbotapi.BotAPI, chatID int64) {
	GenerateWeeklyRecap(bot, chatID)
}

// ListExpenses sends the last 10 expenses to the user
func ListExpenses(bot *tgbotapi.BotAPI, chatID int64) {
	// Get user's expenses (most recent 10)
	expenses, err := database.GetExpensesByUserID(uint(chatID))
	if err != nil {
		log.Printf("Error fetching expenses: %v", err)
		return
	}

	if len(expenses) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Kamu belum memiliki pengeluaran yang tercatat.")
		bot.Send(msg)
		return
	}

	// Limit to last 10 expenses
	if len(expenses) > 10 {
		expenses = expenses[:10]
	}

	// Format the list message
	listText := "📋 10 Pengeluaran Terakhir Kamu:\n\n"

	for _, expense := range expenses {
		listText += fmt.Sprintf("ID: %d\n   %s\n   Rp%s\n   Kategori: %s\n   Tanggal: %s\n\n",
			expense.ID,
			expense.Description,
			formatCurrency(expense.Amount),
			expense.Category,
			expense.Date.Format("2 Jan 2006"))
	}

	listText += "Kamu bisa hapus dengan perintah: /hapus ID\nContoh: /hapus 5\n\n"
	listText += "Kamu bisa update dengan perintah: /update ID deskripsi jumlah kategori\nContoh: /update 5 beli buku 50000 Pendidikan"

	// Send the message to the chat
	msg := tgbotapi.NewMessage(chatID, listText)
	bot.Send(msg)
}

// DeleteExpense deletes the specified expense
func DeleteExpense(bot *tgbotapi.BotAPI, chatID int64, expenseID uint) {
	err := database.DeleteExpense(uint(chatID), expenseID)
	if err != nil {
		log.Printf("Error deleting expense: %v", err)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Gagal menghapus pengeluaran dengan ID %d.", expenseID))
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Pengeluaran dengan ID %d berhasil dihapus.", expenseID))
	bot.Send(msg)
}

// UpdateExpense updates the specified expense
func UpdateExpense(bot *tgbotapi.BotAPI, chatID int64, expenseID uint, description string, amount float64, category string) {
	// Get the expense first
	expense, err := database.GetExpenseByID(uint(chatID), expenseID)
	if err != nil {
		log.Printf("Error getting expense to update: %v", err)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Pengeluaran dengan ID %d tidak ditemukan.", expenseID))
		bot.Send(msg)
		return
	}

	// Update the fields
	expense.Description = description
	expense.Amount = amount
	expense.Category = category

	// Save the updated expense
	err = database.UpdateExpense(expense)
	if err != nil {
		log.Printf("Error updating expense: %v", err)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Gagal mengupdate pengeluaran dengan ID %d.", expenseID))
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Pengeluaran dengan ID %d berhasil diupdate:\n\nDeskripsi: %s\nJumlah: Rp%s\nKategori: %s",
		expenseID, expense.Description, formatCurrency(expense.Amount), expense.Category))
	bot.Send(msg)
}