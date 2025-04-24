package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"internal"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	// MongoDB bağlantısı
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

	// Veritabanı seçimi
	db := client.Database("mcp_db")
	// Repository'lerin oluşturulması
	modelRepo := model.NewModelRepository(db)
	contextRepo := context.NewContextRepository(db)
	protocolRepo := protocol.NewProtocolRepository(db)
	dataRepo := data.NewDataRepository(db)

	// Cursor entegrasyonu
	cursorIntegration, err := cursor.NewIntegration("localhost:50051")
	if err != nil {
		log.Fatalf("Cursor entegrasyonu başlatılamadı: %v", err)
	}
	defer cursorIntegration.client.Close()

	// gRPC sunucusu oluşturma
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("gRPC sunucusu başlatılamadı: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Servisleri kaydetme
	proto.RegisterMCPServiceServer(grpcServer, &server{
		modelRepo:    modelRepo,
		contextRepo:  contextRepo,
		protocolRepo: protocolRepo,
		dataRepo:     dataRepo,
		cursor:       cursorIntegration,
	})

	// Graceful shutdown için sinyal yakalama
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("gRPC sunucusu başlatılıyor: %s", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC sunucusu başlatılamadı: %v", err)
		}
	}()

	// Sinyal bekleme
	<-sigChan
	log.Println("Sunucu kapatılıyor...")
	grpcServer.GracefulStop()
}

// server implements the MCPServiceServer interface
type server struct {
	proto.UnimplementedMCPServiceServer
	modelRepo    *model.ModelRepository
	contextRepo  *context.ContextRepository
	protocolRepo *protocol.ProtocolRepository
	dataRepo     *data.DataRepository
	cursor       *cursor.Integration
}

// CreateModel implements the MCPServiceServer interface
func (s *server) CreateModel(ctx context.Context, req *proto.CreateModelRequest) (*proto.CreateModelResponse, error) {
	model := &model.Model{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Parameters:  req.Parameters,
	}

	if err := s.modelRepo.Create(ctx, model); err != nil {
		return nil, err
	}

	return &proto.CreateModelResponse{
		Model: &proto.Model{
			Id:          model.ID.Hex(),
			Name:        model.Name,
			Type:        model.Type,
			Description: model.Description,
			Parameters:  model.Parameters,
			CreatedAt:   model.CreatedAt.Unix(),
			UpdatedAt:   model.UpdatedAt.Unix(),
		},
	}, nil
}

// CreateContext implements the MCPServiceServer interface
func (s *server) CreateContext(ctx context.Context, req *proto.CreateContextRequest) (*proto.CreateContextResponse, error) {
	context := &context.Context{
		Name:        req.Name,
		Description: req.Description,
		ModelIDs:    req.ModelIds,
		Metadata:    req.Metadata,
	}

	if err := s.contextRepo.Create(ctx, context); err != nil {
		return nil, err
	}

	return &proto.CreateContextResponse{
		Context: &proto.Context{
			Id:          context.ID.Hex(),
			Name:        context.Name,
			Description: context.Description,
			ModelIds:    context.ModelIDs,
			Metadata:    context.Metadata,
			CreatedAt:   context.CreatedAt.Unix(),
			UpdatedAt:   context.UpdatedAt.Unix(),
		},
	}, nil
}

// ExecuteProtocol implements the MCPServiceServer interface
func (s *server) ExecuteProtocol(ctx context.Context, req *proto.ExecuteProtocolRequest) (*proto.ExecuteProtocolResponse, error) {
	execution, err := s.protocolRepo.ExecuteProtocol(ctx, req.ProtocolId, req.ContextId, req.Parameters)
	if err != nil {
		return nil, err
	}

	return &proto.ExecuteProtocolResponse{
		ExecutionId: execution.ID.Hex(),
		Status:      execution.Status,
	}, nil
}

// GetProtocolStatus implements the MCPServiceServer interface
func (s *server) GetProtocolStatus(ctx context.Context, req *proto.GetProtocolStatusRequest) (*proto.GetProtocolStatusResponse, error) {
	execution, err := s.protocolRepo.GetExecutionStatus(ctx, req.ExecutionId)
	if err != nil {
		return nil, err
	}

	return &proto.GetProtocolStatusResponse{
		Status:   execution.Status,
		Result:   execution.Result,
		Metadata: execution.Metadata,
	}, nil
}

// AddData implements the MCPServiceServer interface
func (s *server) AddData(ctx context.Context, req *proto.AddDataRequest) (*proto.AddDataResponse, error) {
	data := &data.Data{
		Type:     req.Type,
		Content:  req.Content,
		Metadata: req.Metadata,
	}

	if err := s.dataRepo.Add(ctx, data); err != nil {
		return nil, err
	}

	return &proto.AddDataResponse{
		Data: &proto.Data{
			Id:        data.ID.Hex(),
			Type:      data.Type,
			Content:   data.Content,
			Metadata:  data.Metadata,
			CreatedAt: data.CreatedAt.Unix(),
			UpdatedAt: data.UpdatedAt.Unix(),
		},
	}, nil
}

// GetData implements the MCPServiceServer interface
func (s *server) GetData(ctx context.Context, req *proto.GetDataRequest) (*proto.GetDataResponse, error) {
	data, err := s.dataRepo.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &proto.GetDataResponse{
		Data: &proto.Data{
			Id:        data.ID.Hex(),
			Type:      data.Type,
			Content:   data.Content,
			Metadata:  data.Metadata,
			CreatedAt: data.CreatedAt.Unix(),
			UpdatedAt: data.UpdatedAt.Unix(),
		},
	}, nil
}

// ListData implements the MCPServiceServer interface
func (s *server) ListData(ctx context.Context, req *proto.ListDataRequest) (*proto.ListDataResponse, error) {
	data, nextPageToken, err := s.dataRepo.List(ctx, req.Type, req.PageSize, req.PageToken)
	if err != nil {
		return nil, err
	}

	var protoData []*proto.Data
	for _, d := range data {
		protoData = append(protoData, &proto.Data{
			Id:        d.ID.Hex(),
			Type:      d.Type,
			Content:   d.Content,
			Metadata:  d.Metadata,
			CreatedAt: d.CreatedAt.Unix(),
			UpdatedAt: d.UpdatedAt.Unix(),
		})
	}

	return &proto.ListDataResponse{
		Data:          protoData,
		NextPageToken: nextPageToken,
	}, nil
}

// DeleteData implements the MCPServiceServer interface
func (s *server) DeleteData(ctx context.Context, req *proto.DeleteDataRequest) (*proto.DeleteDataResponse, error) {
	if err := s.dataRepo.Delete(ctx, req.Id); err != nil {
		return nil, err
	}

	return &proto.DeleteDataResponse{
		Success: true,
	}, nil
}
