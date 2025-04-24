package internal

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Model represents a machine learning model in the system
type Model struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name"`
	Type        string                 `bson:"type" json:"type"`
	Description string                 `bson:"description" json:"description"`
	Parameters  map[string]string      `bson:"parameters" json:"parameters"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
}

// ModelRepository handles database operations for models
type ModelRepository struct {
	collection *mongo.Collection
}

// NewModelRepository creates a new ModelRepository
func NewModelRepository(db *mongo.Database) *ModelRepository {
	return &ModelRepository{
		collection: db.Collection("models"),
	}
}

// Create creates a new model
func (r *ModelRepository) Create(ctx context.Context, model *Model) error {
	model.CreatedAt = time.Now()
	model.UpdatedAt = time.Now()
	
	_, err := r.collection.InsertOne(ctx, model)
	return err
}

// Get retrieves a model by ID
func (r *ModelRepository) Get(ctx context.Context, id string) (*Model, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var model Model
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&model)
	if err != nil {
		return nil, err
	}

	return &model, nil
}

// List retrieves all models with pagination
func (r *ModelRepository) List(ctx context.Context, pageSize int32, pageToken string) ([]*Model, string, error) {
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

	var models []*Model
	if err = cursor.All(ctx, &models); err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if len(models) == int(pageSize) {
		nextPageToken = models[len(models)-1].ID.Hex()
	}

	return models, nextPageToken, nil
}