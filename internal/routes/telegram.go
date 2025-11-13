package routes

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gofiber/fiber/v2"

	"SmartExpenseAI/internal/database"
	"SmartExpenseAI/internal/services"
)

// Global variables to hold bot and allowed user ID
var (
	bot           *tgbotapi.BotAPI
	allowedUserID int64
)

func InitTelegram(botInstance *tgbotapi.BotAPI, userID int64) {
	bot = botInstance
	allowedUserID = userID
}

func TelegramRoutes(app *fiber.App) {
	// Webhook endpoint for Telegram
	app.Post("/webhook", func(c *fiber.Ctx) error {
		var update tgbotapi.Update
		if err := c.BodyParser(&update); err != nil {
			log.Printf("Failed to parse update: %v", err)
			return c.Status(400).SendString("Bad Request")
		}

		// Process text messages only
		if update.Message != nil && update.Message.Text != "" {
			// Check if user is authorized
			if int64(update.Message.From.ID) != allowedUserID {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are not authorized to use this bot.")
				bot.Send(msg)
				return c.SendString("OK")
			}

			// Check if it's a command
			if update.Message.IsCommand() {
				command := update.Message.Command()
				go handleCommand(bot, update.Message, command)
			} else {
				// Process as natural language expense (only for expense entries)
				go func() {
					text := update.Message.Text
					chatID := update.Message.Chat.ID

					log.Printf("Received text: %s", text)

					// Parse the expense using AI (only for expense extraction)
					expense, err := services.ParseExpense(text)
					if err != nil {
						log.Printf("Error parsing expense: %v", err)

						// Instead of generic error message, provide helpful guidance
						responseText := "🤖 Halo! Saya SmartExpenseAI, asisten yang membantu kamu mencatat pengeluaran.\n\n" +
							"Kamu bisa kirim pesan seperti:\n" +
							"• \"makan nasi padang 25000\"\n" +
							"• \"beli buku 50k\"\n\n" +
							"Untuk fitur lainnya, gunakan perintah:\n" +
							"• /lihat - Lihat pengeluaran terakhir\n" +
							"• /bulan - Rekap bulan ini\n" +
							"• /hapus - Hapus pengeluaran"

						msg := tgbotapi.NewMessage(chatID, responseText)
						bot.Send(msg)
						return
					}

					// Only save if there's actual expense data (amount > 0)
					if expense.Amount <= 0 {
						// No expense data found, provide helpful response
						responseText := "🤖 Tidak bisa mengenali pengeluaran dari pesanmu.\n\n" +
							"Contoh format yang benar:\n" +
							"• \"makan nasi padang 25000\"\n" +
							"• \"beli buku 50k\"\n\n" +
							"Untuk fitur lainnya, gunakan perintah:\n" +
							"• /lihat - Lihat pengeluaran terakhir\n" +
							"• /bulan - Rekap bulan ini\n" +
							"• /hapus - Hapus pengeluaran"

						msg := tgbotapi.NewMessage(chatID, responseText)
						bot.Send(msg)
						return
					}

					log.Printf("Parsed expense: %+v", expense)

					// Set the user ID for the expense
					expense.UserID = uint(allowedUserID)

					// Save to database
					err = database.DB.Create(&expense).Error
					if err != nil {
						log.Printf("Error saving expense to database: %v", err)

						msg := tgbotapi.NewMessage(chatID, "Error saving your expense. Please try again.")
						bot.Send(msg)
						return
					}

					log.Printf("Saved expense to database")

					// Format the amount with thousands separator
					amountStr := formatCurrency(expense.Amount)

					log.Printf("AmountStr: %s", amountStr)

					// Send response to user
					responseText := fmt.Sprintf("✅ Disimpan:\nKategori: %s\nJumlah: Rp%s\nDeskripsi: %s",
						expense.Category,
						amountStr,
						expense.Description)

					log.Printf("Response text: %s", responseText)

					msg := tgbotapi.NewMessage(chatID, responseText)
					sendResult, err := bot.Send(msg)
					if err != nil {
						log.Printf("Error sending message to user: %v", err)
					} else {
						log.Printf("Message sent successfully: %+v", sendResult)
					}
				}()
			}
		}

		return c.SendString("OK")
	})

	// Setup webhook route
	app.Get("/setup-webhook", func(c *fiber.Ctx) error {
		webhookURL := fmt.Sprintf("%s/webhook", c.BaseURL())

		_, err := bot.SetWebhook(tgbotapi.NewWebhook(webhookURL))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to set webhook: %v", err),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Webhook set successfully",
			"url":     webhookURL,
		})
	})
}

func handleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message, command string) {
	chatID := message.Chat.ID

	switch command {
	case "start":
		helpText := "🤖 Selamat datang di SmartExpenseAI!\n\n" +
			"Fitur yang tersedia:\n" +
			"• Kirim pesan biasa untuk mencatat pengeluaran\n" +
			"• /lihat - Lihat 10 pengeluaran terakhir\n" +
			"• /bulan - Lihat rekap pengeluaran 30 hari terakhir\n" +
			"• /hapus - Hapus pengeluaran (contoh: /hapus 5)\n" +
			"• /update - Update pengeluaran (contoh: /update 5 beli buku 50000 Pendidikan)\n" +
			"• /bantuan - Tampilkan bantuan ini"
		msg := tgbotapi.NewMessage(chatID, helpText)
		bot.Send(msg)

	case "bantuan":
		helpText := "🤖 Bantuan SmartExpenseAI:\n\n" +
			"Cara mencatat pengeluaran:\n" +
			"• Kirim pesan seperti: \"makan nasi padang 25000\" atau \"beli buku 50k\"\n\n" +
			"Perintah yang tersedia:\n" +
			"• /lihat - Lihat 10 pengeluaran terakhir kamu\n" +
			"• /bulan - Lihat rekap pengeluaran 30 hari terakhir per bulan\n" +
			"• /hapus ID - Hapus pengeluaran, ganti ID dengan nomor pengeluaran\n" +
			"• /update ID deskripsi jumlah kategori - Update pengeluaran\n" +
			"• /bantuan - Tampilkan pesan bantuan ini"
		msg := tgbotapi.NewMessage(chatID, helpText)
		bot.Send(msg)

	case "lihat":
		services.ListExpenses(bot, chatID)

	case "bulan":
		services.GenerateMonthlyRecap(bot, chatID)

	case "hapus":
		// Extract expense ID from command arguments
		args := message.CommandArguments()
		if args == "" {
			msg := tgbotapi.NewMessage(chatID, "Silakan berikan ID pengeluaran yang ingin dihapus.\nContoh: /hapus 5")
			bot.Send(msg)
			return
		}

		// Parse the expense ID
		expenseID, err := strconv.ParseUint(args, 10, 32)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "ID pengeluaran harus berupa angka.\nContoh: /hapus 5")
			bot.Send(msg)
			return
		}

		services.DeleteExpense(bot, chatID, uint(expenseID))

	case "update":
		// Extract arguments from command
		args := message.CommandArguments()
		if args == "" {
			msg := tgbotapi.NewMessage(chatID, "Format salah. Gunakan: /update ID deskripsi jumlah kategori\nContoh: /update 5 beli buku 50000 Pendidikan")
			bot.Send(msg)
			return
		}

		// Parse the arguments (ID, description, amount, category)
		parts := strings.SplitN(args, " ", 4)
		if len(parts) < 4 {
			msg := tgbotapi.NewMessage(chatID, "Format salah. Gunakan: /update ID deskripsi jumlah kategori\nContoh: /update 5 beli buku 50000 Pendidikan")
			bot.Send(msg)
			return
		}

		// Parse expense ID
		expenseID, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "ID pengeluaran harus berupa angka.\nContoh: /update 5 beli buku 50000 Pendidikan")
			bot.Send(msg)
			return
		}

		description := parts[1]
		amountStr := parts[2]
		category := parts[3]

		// Parse amount
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Jumlah harus berupa angka.\nContoh: /update 5 beli buku 50000 Pendidikan")
			bot.Send(msg)
			return
		}

		services.UpdateExpense(bot, chatID, uint(expenseID), description, amount, category)

	default:
		msg := tgbotapi.NewMessage(chatID, "Perintah tidak dikenali. Gunakan /bantuan untuk melihat bantuan.")
		bot.Send(msg)
	}
}

// handleNaturalCommand handles commands detected from natural language
func handleNaturalCommand(bot *tgbotapi.BotAPI, chatID int64, command string, argsStr string, originalText string) {
	switch command {
	case "list":
		services.ListExpenses(bot, chatID)
	case "monthly":
		services.GenerateMonthlyRecap(bot, chatID)
	case "delete":
		// Extract ID from args
		// Simplified: assume argsStr contains the ID
		argsStr = strings.Trim(argsStr, "[]\"")
		if argsStr != "" && argsStr != "[]" && argsStr != "" {
			// Remove quotes if present
			argsStr = strings.ReplaceAll(argsStr, "\"", "")
			parts := strings.Split(argsStr, ",")
			if len(parts) > 0 {
				idStr := strings.TrimSpace(parts[0])
				if idStr != "" {
					id, err := strconv.ParseUint(idStr, 10, 32)
					if err != nil {
						msg := tgbotapi.NewMessage(chatID, "ID pengeluaran tidak valid.")
						bot.Send(msg)
						return
					}
					services.DeleteExpense(bot, chatID, uint(id))
					return
				}
			}
		}
		// If no ID provided or invalid
		msg := tgbotapi.NewMessage(chatID, "Silakan berikan ID pengeluaran yang ingin dihapus.")
		bot.Send(msg)
	case "update":
		// For update, we need to parse the original text to extract ID, description, amount, and category
		// This is more complex and we'll implement a simplified version
		parts := strings.Split(originalText, " ")
		if len(parts) >= 4 {
			// Look for a number in the text (the ID)
			idStr := ""
			for _, part := range parts {
				if _, err := strconv.Atoi(part); err == nil {
					idStr = part
					break
				}
			}

			if idStr != "" {
				id, err := strconv.ParseUint(idStr, 10, 32)
				if err != nil {
					msg := tgbotapi.NewMessage(chatID, "ID pengeluaran tidak valid.")
					bot.Send(msg)
					return
				}

				// For now, we'll just send a message to indicate update functionality
				// In a real implementation, we'd parse the description, amount, and category
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Fitur pengubahan pengeluaran ID %d sedang dalam pengembangan. Silakan gunakan perintah /update untuk saat ini.", id))
				bot.Send(msg)
				return
			}
		}

		msg := tgbotapi.NewMessage(chatID, "Format tidak dikenali. Untuk mengupdate pengeluaran, sebutkan ID pengeluaran yang ingin diubah.")
		bot.Send(msg)
	case "help":
		helpText := "🤖 Bantuan SmartExpenseAI:\n\n" +
			"Cara mencatat pengeluaran:\n" +
			"• Kirim pesan seperti: \"makan nasi padang 25000\" atau \"beli buku 50k\"\n\n" +
			"Perintah alami yang bisa kamu gunakan:\n" +
			"• \"lihat pengeluaranku\" - Lihat 10 pengeluaran terakhir kamu\n" +
			"• \"rekap bulan ini\" - Lihat rekap pengeluaran 30 hari terakhir\n" +
			"• \"hapus pengeluaran 5\" - Hapus pengeluaran dengan ID tertentu\n" +
			"• \"bantuan\" - Tampilkan pesan bantuan ini"
		msg := tgbotapi.NewMessage(chatID, helpText)
		bot.Send(msg)
	case "weekly":
		// Call the weekly recap function for the user
		services.CallWeeklyRecapForUser(bot, chatID)
	default:
		// For unknown commands, send a message
		msg := tgbotapi.NewMessage(chatID, "Perintah tidak dikenali. Gunakan perintah seperti 'lihat pengeluaranku' atau kirim pesan untuk mencatat pengeluaran baru.")
		bot.Send(msg)
	}
}

// isLikelyCommand checks if text likely contains a command
func isLikelyCommand(text string) bool {
	lowerText := strings.ToLower(text)

	// Check for keywords that indicate commands
	commandKeywords := []string{
		"lihat", "tampilkan", "pengeluaranku", "daftar", "list",
		"rekap", "ringkasan", "summary", "bulan", "minggu",
		"hapus", "delete", "remove", "bantuan", "help",
		"apa yang", "perintah", "commands", "cara", "histori", "semua",
		"terakhir", "pengeluaran", "my expenses", "show expenses",
	}

	for _, keyword := range commandKeywords {
		if strings.Contains(lowerText, keyword) {
			return true
		}
	}
	return false
}

// detectLikelyCommand detects what type of command this likely is
func detectLikelyCommand(text string) (string, []string) {
	lowerText := strings.ToLower(text)

	// Check for list commands
	if containsAny(lowerText, []string{
		"lihat pengeluaran", "tampilkan pengeluaran", "pengeluaranku", "daftar pengeluaran",
		"list pengeluaran", "show expenses", "my expenses", "lihat rekap", "lihat daftar",
		"tampilkan daftar", "daftar terakhir", "lihat terakhir", "pengeluaran terakhir",
		"lihat semua", "tampilkan semua", "lihat histori", "tampilkan histori"}) {
		return "list", []string{}
	}

	// Check for monthly recap commands
	if containsAny(lowerText, []string{"rekap bulan", "pengeluaran bulan", "ringkasan bulan", "summary bulan", "monthly recap", "month summary"}) {
		return "monthly", []string{}
	}

	// Check for delete commands (with ID)
	if containsAny(lowerText, []string{"hapus", "delete"}) {
		// Extract ID from text
		re := regexp.MustCompile(`\d+`)
		matches := re.FindAllString(text, -1)
		if len(matches) > 0 {
			return "delete", []string{matches[0]}
		}
		return "delete", []string{}
	}

	// Check for help commands
	if containsAny(lowerText, []string{"bantuan", "help", "cara pakai", "perintah", "commands", "tutor", "cara guna"}) {
		return "help", []string{}
	}

	// Check for weekly recap commands
	if containsAny(lowerText, []string{"rekap minggu", "pengeluaran minggu", "ringkasan minggu", "summary minggu"}) {
		return "weekly", []string{}
	}

	return "", []string{}
}

// containsAny checks if the text contains any of the keywords
func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
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
