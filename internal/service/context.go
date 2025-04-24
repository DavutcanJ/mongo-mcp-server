package internal

import (
	"context"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Context represents a context in the system
type Context struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name"`
	Description string                 `bson:"description" json:"description"`
	ModelIDs    []string              `bson:"model_ids" json:"model_ids"`
	Metadata    map[string]string      `bson:"metadata" json:"metadata"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
}

// ContextRepository handles database operations for contexts
type ContextRepository struct {
	collection *mongo.Collection
}

// NewContextRepository creates a new ContextRepository
func NewContextRepository(db *mongo.Database) *ContextRepository {
	return &ContextRepository{
		collection: db.Collection("contexts"),
	}
}

// Create creates a new context
func (r *ContextRepository) Create(ctx context.Context, context *Context) error {
	context.CreatedAt = time.Now()
	context.UpdatedAt = time.Now()
	
	_, err := r.collection.InsertOne(ctx, context)
	return err
}

// Get retrieves a context by ID
func (r *ContextRepository) Get(ctx context.Context, id string) (*Context, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var context Context
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&context)
	if err != nil {
		return nil, err
	}

	return &context, nil
}

// List retrieves all contexts with pagination
func (r *ContextRepository) List(ctx context.Context, pageSize int32, pageToken string) ([]*Context, string, error) {
	filter := bson.M{}
	if pageToken != "" {
		objectID, err := primitive.ObjectIDFromHex(pageToken)
		if err == nil {
			filter["_id"] = bson.M{"$gt": objectID}
		}
	}

	opts := options.Find().SetLimit(int64(pageSize))
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}
	defer cursor.Close(ctx)

	var contexts []*Context
	if err = cursor.All(ctx, &contexts); err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if len(contexts) == int(pageSize) {
		nextPageToken = contexts[len(contexts)-1].ID.Hex()
	}

	return contexts, nextPageToken, nil
}