package main

import (
	"crypto/tls"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"os"
	"time"
)

// Liquidation Определяем структуру для данных
type Liquidation struct {
	Minute        string  `db:"minute"`         // Соответствует алиасу "minute" в SQL-запросе
	TotalQuantity float64 `db:"total_quantity"` // Соответствует алиасу "total_quantity" в SQL-запросе
}

func setupRoutes(app *fiber.App, conn *sqlx.DB) {
	app.Get("/liquidation", func(c *fiber.Ctx) error {
		start := time.Now() // Start measuring elapsed time

		// SQL-запрос
		query := `
		SELECT 
			toStartOfMinute(created_at) AS minute, 
			SUM(orig_quantity) AS total_quantity 
		FROM liquidations 
		WHERE symbol = 'ADAUSDT' 
		GROUP BY toStartOfMinute(created_at) 
		ORDER BY minute DESC 
		LIMIT 100
		`

		// Define slice for result data
		var liquidations []Liquidation
		err := conn.Select(&liquidations, query)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Failed to fetch liquidations",
				"details": err.Error(),
			})
		}

		// Compute elapsed time
		elapsed := time.Since(start)

		// Return the data along with the elapsed time
		return c.JSON(fiber.Map{
			"liquidations": liquidations,
			"elapsed":      elapsed.String(),
		})
	})
}

func loadEnvData() (string, string, string) {
	clickhouseAddress := os.Getenv("CLICKHOUSE_ADDRESS")
	clickhouseUsername := os.Getenv("CLICKHOUSE_USERNAME")
	clickhousePassword := os.Getenv("CLICKHOUSE_PASSWORD")
	return clickhouseAddress, clickhouseUsername, clickhousePassword
}

func main() {
	address, username, password := loadEnvData()

	conn := sqlx.NewDb(clickhouse.OpenDB(&clickhouse.Options{
		Addr:     []string{address},
		Protocol: clickhouse.Native,
		TLS:      &tls.Config{}, // Enable secure TLS
		Auth: clickhouse.Auth{
			Username: username,
			Password: password,
		},
	}), "clickhouse")

	// Initialize Fiber app
	app := fiber.New()

	// Setup Fiber routes
	setupRoutes(app, conn)

	// Start the server
	if err := app.Listen(":3000"); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
