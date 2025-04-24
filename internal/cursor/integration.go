package cursor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"mongo-mcp-server/internal/proto"
)

// Integration represents the Cursor MCP integration
type Integration struct {
	client *Client
	config *Config
}

// Config represents the Cursor MCP configuration
type Config struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Features    []string `json:"features"`
	Commands    struct {
		Prefix    string   `json:"prefix"`
		Available []string `json:"available"`
	} `json:"commands"`
}

// NewIntegration creates a new Cursor MCP integration
func NewIntegration(addr string) (*Integration, error) {
	client, err := NewClient(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	return &Integration{
		client: client,
		config: config,
	}, nil
}

// HandleCommand handles Cursor MCP commands
func (i *Integration) HandleCommand(ctx context.Context, command string, args []string) (string, error) {
	switch command {
	case "model":
		return i.handleModelCommand(ctx, args)
	case "context":
		return i.handleContextCommand(ctx, args)
	case "execute":
		return i.handleExecuteCommand(ctx, args)
	case "status":
		return i.handleStatusCommand(ctx, args)
	case "data":
		return i.handleDataCommand(ctx, args)
	default:
		return "", fmt.Errorf("unknown command: %s", command)
	}
}

// handleModelCommand handles model-related commands
func (i *Integration) handleModelCommand(ctx context.Context, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("model command requires subcommand")
	}

	switch args[0] {
	case "create":
		if len(args) < 4 {
			return "", fmt.Errorf("create model requires name, type, and description")
		}
		model, err := i.client.CreateModel(ctx, args[1], args[2], args[3], nil)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Model created: %s", model.Id), nil
	case "list":
		models, _, err := i.client.ListModels(ctx, 10, "")
		if err != nil {
			return "", err
		}
		return formatModels(models), nil
	default:
		return "", fmt.Errorf("unknown model subcommand: %s", args[0])
	}
}

// handleContextCommand handles context-related commands
func (i *Integration) handleContextCommand(ctx context.Context, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("context command requires subcommand")
	}

	switch args[0] {
	case "create":
		if len(args) < 3 {
			return "", fmt.Errorf("create context requires name and description")
		}
		context, err := i.client.CreateContext(ctx, args[1], args[2], nil, nil)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Context created: %s", context.Id), nil
	case "list":
		contexts, _, err := i.client.ListContexts(ctx, 10, "")
		if err != nil {
			return "", err
		}
		return formatContexts(contexts), nil
	default:
		return "", fmt.Errorf("unknown context subcommand: %s", args[0])
	}
}

// handleExecuteCommand handles protocol execution commands
func (i *Integration) handleExecuteCommand(ctx context.Context, args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("execute command requires protocol_id and context_id")
	}

	executionID, err := i.client.ExecuteProtocol(ctx, args[0], args[1], nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Protocol execution started: %s", executionID), nil
}

// handleStatusCommand handles protocol status commands
func (i *Integration) handleStatusCommand(ctx context.Context, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("status command requires execution_id")
	}

	status, err := i.client.GetProtocolStatus(ctx, args[0])
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Status: %s\nResult: %s", status.Status, status.Result), nil
}

// handleDataCommand handles data-related commands
func (i *Integration) handleDataCommand(ctx context.Context, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("data command requires subcommand")
	}

	switch args[0] {
	case "add":
		if len(args) < 3 {
			return "", fmt.Errorf("add data requires type and content")
		}
		data, err := i.client.AddData(ctx, args[1], args[2], nil)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Data added: %s", data.Id), nil
	case "list":
		data, _, err := i.client.ListData(ctx, "", 10, "")
		if err != nil {
			return "", err
		}
		return formatData(data), nil
	default:
		return "", fmt.Errorf("unknown data subcommand: %s", args[0])
	}
}

// loadConfig loads the Cursor MCP configuration
func loadConfig() (*Config, error) {
	configPath := filepath.Join("pkg", "cursor", "mcp.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Helper functions for formatting output
func formatModels(models []*proto.Model) string {
	var result string
	for _, m := range models {
		result += fmt.Sprintf("ID: %s\nName: %s\nType: %s\nDescription: %s\n\n",
			m.Id, m.Name, m.Type, m.Description)
	}
	return result
}

func formatContexts(contexts []*proto.Context) string {
	var result string
	for _, c := range contexts {
		result += fmt.Sprintf("ID: %s\nName: %s\nDescription: %s\nModel IDs: %v\n\n",
			c.Id, c.Name, c.Description, c.ModelIds)
	}
	return result
}

func formatData(data []*proto.Data) string {
	var result string
	for _, d := range data {
		result += fmt.Sprintf("ID: %s\nType: %s\nContent: %s\n\n",
			d.Id, d.Type, d.Content)
	}
	return result
}
