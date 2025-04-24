package internal

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Data represents data in the system
type Data struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Type      string                 `bson:"type" json:"type"`
	Content   string                 `bson:"content" json:"content"`
	Metadata  map[string]string      `bson:"metadata" json:"metadata"`
	CreatedAt time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time              `bson:"updated_at" json:"updated_at"`
}

// DataRepository handles database operations for data
type DataRepository struct {
	collection *mongo.Collection
}

// NewDataRepository creates a new DataRepository
func NewDataRepository(db *mongo.Database) *DataRepository {
	return &DataRepository{
		collection: db.Collection("data"),
	}
}

// Add adds new data
func (r *DataRepository) Add(ctx context.Context, data *Data) error {
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	
	_, err := r.collection.InsertOne(ctx, data)
	return err
}

// Get retrieves data by ID
func (r *DataRepository) Get(ctx context.Context, id string) (*Data, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var data Data
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// List retrieves data with pagination and optional type filter
func (r *DataRepository) List(ctx context.Context, dataType string, pageSize int32, pageToken string) ([]*Data, string, error) {
	filter := bson.M{}
	if dataType != "" {
		filter["type"] = dataType
	}
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

	var data []*Data
	if err = cursor.All(ctx, &data); err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if len(data) == int(pageSize) {
		nextPageToken = data[len(data)-1].ID.Hex()
	}

	return data, nextPageToken, nil
}

// Delete removes data by ID
func (r *DataRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}