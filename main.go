package main

import (
	"fmt"
	"log"
	"os"
	"tx-building-go/transactions"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Cardano Transaction Builder using Apollo")
	fmt.Println("==========================================\n")

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	// Get required environment variables
	mnemonic := os.Getenv("MNEMONIC")
	if mnemonic == "" {
		log.Fatal("MNEMONIC environment variable is required")
	}

	maestroAPIKey := os.Getenv("MAESTRO_API_KEY")
	if maestroAPIKey == "" {
		log.Fatal("MAESTRO_API_KEY environment variable is required")
	}

	blockfrostProjectID := os.Getenv("BLOCKFROST_PROJECT_ID")
	if blockfrostProjectID == "" {
		log.Fatal("BLOCKFROST_PROJECT_ID environment variable is required")
	}

	fmt.Println("Environment variables loaded successfully\n")

	// Build and send transaction
	fmt.Println("Building and sending transaction...")
	txID, err := transactions.SendAda(mnemonic, maestroAPIKey, blockfrostProjectID)
	if err != nil {
		log.Fatalf("Transaction failed: %v", err)
	}

	fmt.Printf("\nTransaction completed successfully!\n")
	fmt.Printf("Transaction ID: %s\n", txID)
}
