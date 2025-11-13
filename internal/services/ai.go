package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/raflinoob132/SmartExpenseAI/internal/models"
)

type OpenRouterRequest struct {
	Model         string      `json:"model"`
	Messages      []Message   `json:"messages"`
	ResponseFormat interface{} `json:"response_format,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}


func ParseExpense(text string) (models.Expense, error) {
	var expense models.Expense

	// Get the API key from environment
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return expense, fmt.Errorf("OPENROUTER_API_KEY environment variable is not set")
	}

	// Prepare the prompt for the AI - focus only on expense extraction
	prompt := fmt.Sprintf(`Extract expense information from the following text. If the text doesn't contain expense information, return default values.

	Text: "%s"

	Respond in JSON format with the following structure:
	{
		"description": "the item or service purchased",
		"category": "the category of expense (e.g., Food, Transport, etc.)",
		"amount": "the numeric amount in rupiah (as a number)",
		"date": "the date in YYYY-MM-DD format (use today's date if not specified)"
	}

	If no expense information is found, return:
	{
		"description": "",
		"category": "",
		"amount": 0,
		"date": "%s"
	}`, text, time.Now().Format("2006-01-02"))

	// Prepare the request body
	requestBody := OpenRouterRequest{
		Model: "openai/gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		ResponseFormat: map[string]interface{}{
			"type": "json_object",
		},
	}

	// Convert request body to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return expense, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return expense, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the API call
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return expense, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return expense, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Decode the response
	var openRouterResp OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&openRouterResp); err != nil {
		return expense, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return expense, fmt.Errorf("no choices in AI response")
	}

	// Parse the JSON response from AI
	var expenseResp struct {
		Description string  `json:"description"`
		Category    string  `json:"category"`
		Amount      float64 `json:"amount"`
		Date        string  `json:"date"`
	}

	responseContent := openRouterResp.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(responseContent), &expenseResp); err != nil {
		return expense, fmt.Errorf("failed to unmarshal expense data: %w", err)
	}

	// Convert the date string to time.Time
	var date time.Time
	if expenseResp.Date != "" {
		var err error
		date, err = time.Parse("2006-01-02", expenseResp.Date)
		if err != nil {
			// If date parsing fails, use current date
			date = time.Now()
		}
	} else {
		// If no date provided, use current date
		date = time.Now()
	}

	// Create and return the Expense model
	expense = models.Expense{
		Description: expenseResp.Description,
		Category:    expenseResp.Category,
		Amount:      expenseResp.Amount,
		Date:        date,
		CreatedAt:   time.Now(),
	}

	return expense, nil
}

