package cursor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavutcanJ/mongo-mcp-server/pkg/proto"
	"google.golang.org/grpc"
)

// Integration represents the Cursor MCP integration
type Integration struct {
	client proto.MCPServiceClient
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
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %v", err)
	}
	client := proto.NewMCPServiceClient(conn)

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
			return "", fmt.Errorf("create model requires name, type, and parameters")
		}

		params := make(map[string]string)
		if len(args) > 4 {
			err := json.Unmarshal([]byte(args[4]), &params)
			if err != nil {
				return "", fmt.Errorf("invalid parameters JSON: %v", err)
			}
		}

		resp, err := i.client.CreateModel(ctx, &proto.Model{
			Name:       args[1],
			Type:       args[2],
			Parameters: params,
		})
		if err != nil {
			return "", err
		}
		if resp.Error != "" {
			return "", fmt.Errorf(resp.Error)
		}
		return fmt.Sprintf("Model created: %s", resp.Model.Id), nil

	case "get":
		if len(args) < 2 {
			return "", fmt.Errorf("get model requires id")
		}
		resp, err := i.client.GetModel(ctx, &proto.ModelRequest{Id: args[1]})
		if err != nil {
			return "", err
		}
		if resp.Error != "" {
			return "", fmt.Errorf(resp.Error)
		}
		return formatModel(resp.Model), nil

	case "list":
		resp, err := i.client.ListModels(ctx, &proto.ListRequest{
			PageSize: 10,
			Filters:  make(map[string]string),
		})
		if err != nil {
			return "", err
		}
		if resp.Error != "" {
			return "", fmt.Errorf(resp.Error)
		}
		return formatModels(resp.Models), nil

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
			return "", fmt.Errorf("create context requires name and content")
		}

		metadata := make(map[string]string)
		if len(args) > 3 {
			err := json.Unmarshal([]byte(args[3]), &metadata)
			if err != nil {
				return "", fmt.Errorf("invalid metadata JSON: %v", err)
			}
		}

		resp, err := i.client.CreateContext(ctx, &proto.Context{
			Name:     args[1],
			Content:  args[2],
			Metadata: metadata,
		})
		if err != nil {
			return "", err
		}
		if resp.Error != "" {
			return "", fmt.Errorf(resp.Error)
		}
		return fmt.Sprintf("Context created: %s", resp.Context.Id), nil

	case "get":
		if len(args) < 2 {
			return "", fmt.Errorf("get context requires id")
		}
		resp, err := i.client.GetContext(ctx, &proto.ContextRequest{Id: args[1]})
		if err != nil {
			return "", err
		}
		if resp.Error != "" {
			return "", fmt.Errorf(resp.Error)
		}
		return formatContext(resp.Context), nil

	case "list":
		resp, err := i.client.ListContexts(ctx, &proto.ListRequest{
			PageSize: 10,
			Filters:  make(map[string]string),
		})
		if err != nil {
			return "", err
		}
		if resp.Error != "" {
			return "", fmt.Errorf(resp.Error)
		}
		return formatContexts(resp.Contexts), nil

	default:
		return "", fmt.Errorf("unknown context subcommand: %s", args[0])
	}
}

// handleExecuteCommand handles protocol execution commands
func (i *Integration) handleExecuteCommand(ctx context.Context, args []string) (string, error) {
	if len(args) < 3 {
		return "", fmt.Errorf("execute command requires model_id, context_id and input")
	}

	params := make(map[string]string)
	if len(args) > 3 {
		err := json.Unmarshal([]byte(args[3]), &params)
		if err != nil {
			return "", fmt.Errorf("invalid parameters JSON: %v", err)
		}
	}

	resp, err := i.client.ExecuteProtocol(ctx, &proto.Protocol{
		ModelId:    args[0],
		ContextId:  args[1],
		Input:      args[2],
		Parameters: params,
	})
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", fmt.Errorf(resp.Error)
	}
	return fmt.Sprintf("Protocol execution started: %s", resp.Id), nil
}

// handleStatusCommand handles protocol status commands
func (i *Integration) handleStatusCommand(ctx context.Context, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("status command requires execution_id")
	}

	resp, err := i.client.GetProtocolStatus(ctx, &proto.ProtocolRequest{Id: args[0]})
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", fmt.Errorf(resp.Error)
	}
	return fmt.Sprintf("Status: %s", resp.Status), nil
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

		metadata := make(map[string]string)
		if len(args) > 3 {
			err := json.Unmarshal([]byte(args[3]), &metadata)
			if err != nil {
				return "", fmt.Errorf("invalid metadata JSON: %v", err)
			}
		}

		resp, err := i.client.AddData(ctx, &proto.Data{
			Type:     args[1],
			Content:  []byte(args[2]),
			Metadata: metadata,
		})
		if err != nil {
			return "", err
		}
		if resp.Error != "" {
			return "", fmt.Errorf(resp.Error)
		}
		return fmt.Sprintf("Data added: %s", resp.Data.Id), nil

	case "get":
		if len(args) < 2 {
			return "", fmt.Errorf("get data requires id")
		}
		resp, err := i.client.GetData(ctx, &proto.DataRequest{Id: args[1]})
		if err != nil {
			return "", err
		}
		if resp.Error != "" {
			return "", fmt.Errorf(resp.Error)
		}
		return formatData(resp.Data), nil

	case "list":
		resp, err := i.client.ListData(ctx, &proto.ListRequest{
			PageSize: 10,
			Filters:  make(map[string]string),
		})
		if err != nil {
			return "", err
		}
		if resp.Error != "" {
			return "", fmt.Errorf(resp.Error)
		}
		return formatDataList(resp.Data), nil

	case "delete":
		if len(args) < 2 {
			return "", fmt.Errorf("delete data requires id")
		}
		resp, err := i.client.DeleteData(ctx, &proto.DataRequest{Id: args[1]})
		if err != nil {
			return "", err
		}
		if resp.Error != "" {
			return "", fmt.Errorf(resp.Error)
		}
		return "Data deleted successfully", nil

	default:
		return "", fmt.Errorf("unknown data subcommand: %s", args[0])
	}
}

// loadConfig loads the Cursor MCP configuration
func loadConfig() (*Config, error) {
	configPath := filepath.Join("configs", "cursor.json")
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
func formatModel(m *proto.Model) string {
	return fmt.Sprintf("ID: %s\nName: %s\nType: %s\nParameters: %v\n",
		m.Id, m.Name, m.Type, m.Parameters)
}

func formatModels(models []*proto.Model) string {
	var result string
	for _, m := range models {
		result += formatModel(m) + "\n"
	}
	return result
}

func formatContext(c *proto.Context) string {
	return fmt.Sprintf("ID: %s\nName: %s\nContent: %s\nMetadata: %v\n",
		c.Id, c.Name, c.Content, c.Metadata)
}

func formatContexts(contexts []*proto.Context) string {
	var result string
	for _, c := range contexts {
		result += formatContext(c) + "\n"
	}
	return result
}

func formatData(d *proto.Data) string {
	return fmt.Sprintf("ID: %s\nType: %s\nContent: %s\nMetadata: %v\n",
		d.Id, d.Type, string(d.Content), d.Metadata)
}

func formatDataList(data []*proto.Data) string {
	var result string
	for _, d := range data {
		result += formatData(d) + "\n"
	}
	return result
}
