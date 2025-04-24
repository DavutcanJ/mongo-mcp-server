package server

import (
	"context"
	"fmt"
	"log"
	"mongo-mcp-server/internal/config"
	"mongo-mcp-server/internal/database"
	"mongo-mcp-server/pkg/proto"
	"net"

	"google.golang.org/grpc"
)

// Server implements the MCPServiceServer interface
type Server struct {
	proto.UnimplementedMCPServiceServer
	cfg    *config.Config
	db     *database.MongoDB
	server *grpc.Server
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg: cfg,
	}
}

func (s *Server) Start() error {
	// Initialize MongoDB connection
	db, err := database.NewMongoDB(s.cfg)
	if err != nil {
		return err
	}
	s.db = db

	// Create gRPC server
	s.server = grpc.NewServer()

	// Register services
	proto.RegisterMCPServiceServer(s.server, s)

	// Start listening
	addr := fmt.Sprintf("%s:%d", s.cfg.Connection.Host, s.cfg.Connection.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Printf("Server listening at %v", lis.Addr())
	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
	if s.db != nil {
		s.db.Close()
	}
}

// Model operations
func (s *Server) CreateModel(ctx context.Context, req *proto.Model) (*proto.ModelResponse, error) {
	return &proto.ModelResponse{}, nil
}

func (s *Server) GetModel(ctx context.Context, req *proto.ModelRequest) (*proto.ModelResponse, error) {
	return &proto.ModelResponse{}, nil
}

func (s *Server) ListModels(ctx context.Context, req *proto.ListRequest) (*proto.ModelList, error) {
	return &proto.ModelList{}, nil
}

// Context operations
func (s *Server) CreateContext(ctx context.Context, req *proto.Context) (*proto.ContextResponse, error) {
	return &proto.ContextResponse{}, nil
}

func (s *Server) GetContext(ctx context.Context, req *proto.ContextRequest) (*proto.ContextResponse, error) {
	return &proto.ContextResponse{}, nil
}

func (s *Server) ListContexts(ctx context.Context, req *proto.ListRequest) (*proto.ContextList, error) {
	return &proto.ContextList{}, nil
}

// Protocol operations
func (s *Server) ExecuteProtocol(ctx context.Context, req *proto.Protocol) (*proto.ProtocolResponse, error) {
	return &proto.ProtocolResponse{}, nil
}

func (s *Server) GetProtocolStatus(ctx context.Context, req *proto.ProtocolRequest) (*proto.ProtocolStatus, error) {
	return &proto.ProtocolStatus{}, nil
}

// Data operations
func (s *Server) AddData(ctx context.Context, req *proto.Data) (*proto.DataResponse, error) {
	return &proto.DataResponse{}, nil
}

func (s *Server) GetData(ctx context.Context, req *proto.DataRequest) (*proto.DataResponse, error) {
	return &proto.DataResponse{}, nil
}

func (s *Server) ListData(ctx context.Context, req *proto.ListRequest) (*proto.DataList, error) {
	return &proto.DataList{}, nil
}

func (s *Server) DeleteData(ctx context.Context, req *proto.DataRequest) (*proto.DeleteResponse, error) {
	return &proto.DeleteResponse{}, nil
}
