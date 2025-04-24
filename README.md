# MongoDB Model Context Protocol Server

A powerful server implementation for managing models, contexts, and protocols with MongoDB integration and Cursor support.

## Features

- Model management (create, list, update, delete)
- Context management with multiple model support
- Protocol execution and status tracking
- Data storage and retrieval
- Cursor integration for enhanced code completion and analysis
- gRPC API for all operations
- MongoDB backend for persistent storage

## Requirements

- Go 1.16 or higher
- MongoDB 4.4 or higher
- Cursor IDE (for full integration features)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/mongo-mcp-server.git
cd mongo-mcp-server
```

2. Install dependencies:
```bash
go mod download
```

3. Start MongoDB:
```bash
mongod --dbpath /path/to/data/directory
```

4. Build and run the server:
```bash
go build -o mcp-server cmd/server/main.go
./mcp-server
```

## Cursor Integration

The server includes built-in support for Cursor IDE integration. To enable Cursor features:

1. Install the Cursor IDE from [cursor.sh](https://cursor.sh)
2. The server will automatically detect and integrate with Cursor when it's running
3. Use the `/mcp` prefix in Cursor to access MCP features:
   - `/mcp model` - Manage models
   - `/mcp context` - Manage contexts
   - `/mcp execute` - Execute protocols
   - `/mcp status` - Check protocol status
   - `/mcp data` - Manage data

## API Usage

### Models

Create a new model:
```bash
grpcurl -plaintext -d '{
  "name": "gpt-4",
  "type": "llm",
  "description": "OpenAI GPT-4 model",
  "parameters": {
    "temperature": 0.7,
    "max_tokens": 2048
  }
}' localhost:50051 proto.MCPService/CreateModel
```

### Contexts

Create a new context:
```bash
grpcurl -plaintext -d '{
  "name": "code-review",
  "description": "Code review context",
  "model_ids": ["model_id_1", "model_id_2"],
  "metadata": {
    "language": "go",
    "framework": "grpc"
  }
}' localhost:50051 proto.MCPService/CreateContext
```

### Protocols

Execute a protocol:
```bash
grpcurl -plaintext -d '{
  "protocol_id": "protocol_id_1",
  "context_id": "context_id_1",
  "parameters": {
    "input": "Review this code",
    "style": "detailed"
  }
}' localhost:50051 proto.MCPService/ExecuteProtocol
```

Check protocol status:
```bash
grpcurl -plaintext -d '{
  "execution_id": "execution_id_1"
}' localhost:50051 proto.MCPService/GetProtocolStatus
```

### Data

Add data:
```bash
grpcurl -plaintext -d '{
  "type": "code",
  "content": "package main\n\nfunc main() {\n    println(\"Hello, World!\")\n}",
  "metadata": {
    "language": "go",
    "author": "John Doe"
  }
}' localhost:50051 proto.MCPService/AddData
```

Get data:
```bash
grpcurl -plaintext -d '{
  "id": "data_id_1"
}' localhost:50051 proto.MCPService/GetData
```

List data:
```bash
grpcurl -plaintext -d '{
  "type": "code",
  "page_size": 10,
  "page_token": ""
}' localhost:50051 proto.MCPService/ListData
```

Delete data:
```bash
grpcurl -plaintext -d '{
  "id": "data_id_1"
}' localhost:50051 proto.MCPService/DeleteData
```

## Security

- All API endpoints require authentication
- MongoDB connection uses secure defaults
- gRPC communication is encrypted
- Cursor integration uses secure channels

## License

MIT License - see LICENSE file for details
