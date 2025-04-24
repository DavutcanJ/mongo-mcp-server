package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DavutcanJ/mongo-mcp-server/internal/cursor"
	svcContext "github.com/DavutcanJ/mongo-mcp-server/internal/service/context"
	"github.com/DavutcanJ/mongo-mcp-server/internal/service/data"
	"github.com/DavutcanJ/mongo-mcp-server/internal/service/model"
	"github.com/DavutcanJ/mongo-mcp-server/internal/service/protocol"
	"github.com/DavutcanJ/mongo-mcp-server/pkg/proto"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	log.Println("Starting MCP Server...")

	// MongoDB bağlantısı
	log.Println("Connecting to MongoDB...")
	mongoURI := "mongodb://localhost:27017"
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("MongoDB client oluşturulamadı: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("MongoDB'ye bağlanılamadı: %v", err)
	}
	defer client.Disconnect(ctx)

	// Ping MongoDB to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("MongoDB connection test failed: %v", err)
	}
	log.Println("Successfully connected to MongoDB")

	// Veritabanı seçimi
	db := client.Database("mcp_db")
	log.Printf("Using database: %s", db.Name())

	// Repository'lerin oluşturulması
	log.Println("Initializing repositories...")
	modelRepo := model.NewModelRepository(db)
	contextRepo := svcContext.NewContextRepository(db)
	protocolRepo := protocol.NewProtocolRepository(db)
	dataRepo := data.NewDataRepository(db)

	// Cursor entegrasyonu
	log.Println("Initializing cursor integration...")
	cursorIntegration, err := cursor.NewIntegration("localhost:50051")
	if err != nil {
		log.Fatalf("Cursor entegrasyonu başlatılamadı: %v", err)
	}

	// gRPC sunucusu oluşturma
	log.Println("Starting gRPC server...")
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("gRPC sunucusu başlatılamadı: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Servisleri kaydetme
	mcpServer := &server{
		modelRepo:    modelRepo,
		contextRepo:  contextRepo,
		protocolRepo: protocolRepo,
		dataRepo:     dataRepo,
		cursor:       cursorIntegration,
	}

	// gRPC servisini kaydet
	proto.RegisterMCPServiceServer(grpcServer, mcpServer)

	// Graceful shutdown için sinyal yakalama
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server
	log.Printf("MCP Server is running on %s", lis.Addr().String())
	log.Println("Use Ctrl+C to stop the server")

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC sunucusu başlatılamadı: %v", err)
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal %v, initiating graceful shutdown...", sig)

	// Graceful shutdown
	grpcServer.GracefulStop()
	log.Println("Server stopped gracefully")
}

// server implements the MCPServiceServer interface
type server struct {
	proto.UnimplementedMCPServiceServer
	modelRepo    *model.ModelRepository
	contextRepo  *svcContext.ContextRepository
	protocolRepo *protocol.ProtocolRepository
	dataRepo     *data.DataRepository
	cursor       *cursor.Integration
}

// CreateModel implements the MCPServiceServer interface
func (s *server) CreateModel(ctx context.Context, req *proto.Model) (*proto.ModelResponse, error) {
	model := &model.Model{
		Name:       req.Name,
		Type:       req.Type,
		Parameters: req.Parameters,
	}

	if err := s.modelRepo.Create(ctx, model); err != nil {
		return &proto.ModelResponse{Error: err.Error()}, nil
	}

	return &proto.ModelResponse{
		Model: &proto.Model{
			Id:         model.ID.Hex(),
			Name:       model.Name,
			Type:       model.Type,
			Parameters: model.Parameters,
		},
	}, nil
}

// CreateContext implements the MCPServiceServer interface
func (s *server) CreateContext(ctx context.Context, req *proto.Context) (*proto.ContextResponse, error) {
	log.Printf("Creating context with name: %s", req.Name)

	// Initialize metadata if nil
	if req.Metadata == nil {
		req.Metadata = make(map[string]string)
	}

	context := &svcContext.Context{
		Name:     req.Name,
		Content:  req.Content, // Add content field
		Metadata: req.Metadata,
	}

	if err := s.contextRepo.Create(ctx, context); err != nil {
		log.Printf("Error creating context: %v", err)
		return &proto.ContextResponse{Error: err.Error()}, nil
	}

	log.Printf("Context created successfully with ID: %s", context.ID.Hex())
	return &proto.ContextResponse{
		Context: &proto.Context{
			Id:       context.ID.Hex(),
			Name:     context.Name,
			Content:  context.Content, // Return content in response
			Metadata: context.Metadata,
		},
	}, nil
}

// ExecuteProtocol implements the MCPServiceServer interface
func (s *server) ExecuteProtocol(ctx context.Context, req *proto.Protocol) (*proto.ProtocolResponse, error) {
	execution, err := s.protocolRepo.ExecuteProtocol(ctx, req.ModelId, req.ContextId, req.Parameters)
	if err != nil {
		return &proto.ProtocolResponse{Error: err.Error()}, nil
	}

	return &proto.ProtocolResponse{
		Id: execution.ID.Hex(),
	}, nil
}

// GetProtocolStatus implements the MCPServiceServer interface
func (s *server) GetProtocolStatus(ctx context.Context, req *proto.ProtocolRequest) (*proto.ProtocolStatus, error) {
	execution, err := s.protocolRepo.GetExecutionStatus(ctx, req.Id)
	if err != nil {
		return &proto.ProtocolStatus{Error: err.Error()}, nil
	}

	return &proto.ProtocolStatus{
		Status: execution.Status,
	}, nil
}

// AddData implements the MCPServiceServer interface
func (s *server) AddData(ctx context.Context, req *proto.Data) (*proto.DataResponse, error) {
	data := &data.Data{
		Type:     req.Type,
		Content:  string(req.Content),
		Metadata: req.Metadata,
	}

	if err := s.dataRepo.Add(ctx, data); err != nil {
		return &proto.DataResponse{Error: err.Error()}, nil
	}

	return &proto.DataResponse{
		Data: &proto.Data{
			Id:       data.ID.Hex(),
			Type:     data.Type,
			Content:  []byte(data.Content),
			Metadata: data.Metadata,
		},
	}, nil
}

// GetData implements the MCPServiceServer interface
func (s *server) GetData(ctx context.Context, req *proto.DataRequest) (*proto.DataResponse, error) {
	data, err := s.dataRepo.Get(ctx, req.Id)
	if err != nil {
		return &proto.DataResponse{Error: err.Error()}, nil
	}

	return &proto.DataResponse{
		Data: &proto.Data{
			Id:       data.ID.Hex(),
			Type:     data.Type,
			Content:  []byte(data.Content),
			Metadata: data.Metadata,
		},
	}, nil
}

// ListData implements the MCPServiceServer interface
func (s *server) ListData(ctx context.Context, req *proto.ListRequest) (*proto.DataList, error) {
	data, nextPageToken, err := s.dataRepo.List(ctx, req.Filters["type"], int32(req.PageSize), req.Filters["pageToken"])
	if err != nil {
		return &proto.DataList{Error: err.Error()}, nil
	}

	var protoData []*proto.Data
	for _, d := range data {
		protoData = append(protoData, &proto.Data{
			Id:       d.ID.Hex(),
			Type:     d.Type,
			Content:  []byte(d.Content),
			Metadata: d.Metadata,
		})
	}

	if nextPageToken != "" {
		if protoData[0].Metadata == nil {
			protoData[0].Metadata = make(map[string]string)
		}
		protoData[0].Metadata["nextPageToken"] = nextPageToken
	}

	return &proto.DataList{
		Data: protoData,
	}, nil
}

// DeleteData implements the MCPServiceServer interface
func (s *server) DeleteData(ctx context.Context, req *proto.DataRequest) (*proto.DeleteResponse, error) {
	if err := s.dataRepo.Delete(ctx, req.Id); err != nil {
		return &proto.DeleteResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.DeleteResponse{
		Success: true,
	}, nil
}

// ListModels implements the MCPServiceServer interface
func (s *server) ListModels(ctx context.Context, req *proto.ListRequest) (*proto.ModelList, error) {
	log.Printf("ListModels called with page size: %d", req.PageSize)

	if req.PageSize == 0 {
		req.PageSize = 10 // Default page size
	}

	models, nextPageToken, err := s.modelRepo.List(ctx, int32(req.PageSize), req.Filters["pageToken"])
	if err != nil {
		log.Printf("Error listing models: %v", err)
		return &proto.ModelList{Error: err.Error()}, nil
	}

	var protoModels []*proto.Model
	for _, m := range models {
		protoModels = append(protoModels, &proto.Model{
			Id:         m.ID.Hex(),
			Name:       m.Name,
			Type:       m.Type,
			Parameters: m.Parameters,
		})
	}

	if nextPageToken != "" && len(protoModels) > 0 {
		if protoModels[0].Parameters == nil {
			protoModels[0].Parameters = make(map[string]string)
		}
		protoModels[0].Parameters["nextPageToken"] = nextPageToken
	}

	log.Printf("Found %d models", len(protoModels))
	return &proto.ModelList{
		Models: protoModels,
	}, nil
}

// GetModel implements the MCPServiceServer interface
func (s *server) GetModel(ctx context.Context, req *proto.ModelRequest) (*proto.ModelResponse, error) {
	model, err := s.modelRepo.Get(ctx, req.Id)
	if err != nil {
		return &proto.ModelResponse{Error: err.Error()}, nil
	}

	return &proto.ModelResponse{
		Model: &proto.Model{
			Id:         model.ID.Hex(),
			Name:       model.Name,
			Type:       model.Type,
			Parameters: model.Parameters,
		},
	}, nil
}
