package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/DavutcanJ/mongo-mcp-server/internal/cursor"
)

func printUsage() {
	fmt.Println("Usage: mcp-tool <command> [args...]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  model:")
	fmt.Println("    create <name> <type> [parameters]")
	fmt.Println("    get <id>")
	fmt.Println("    list")
	fmt.Println("\n  context:")
	fmt.Println("    create <name> <content> [metadata]")
	fmt.Println("    get <id>")
	fmt.Println("    list")
	fmt.Println("\n  execute <model_id> <context_id> <input> [parameters]")
	fmt.Println("  status <execution_id>")
	fmt.Println("\n  data:")
	fmt.Println("    add <type> <content> [metadata]")
	fmt.Println("    get <id>")
	fmt.Println("    list")
	fmt.Println("    delete <id>")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Create integration
	integration, err := cursor.NewIntegration("localhost:50051")
	if err != nil {
		log.Fatalf("Failed to create integration: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Handle command
	command := os.Args[1]
	args := os.Args[2:]

	result, err := integration.HandleCommand(ctx, command, args)
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}

	// Print result
	if strings.TrimSpace(result) != "" {
		fmt.Println(result)
	}
}
