package apikey

import (
	"context"
	"fmt"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionName = "api_key"

type Repository interface {
	Create(ctx context.Context, key *models.ApiKey) error
	FindAll(ctx context.Context) ([]models.ApiKey, error)
	FindByHash(ctx context.Context, keyHash string) (*models.ApiKey, error)
	Delete(ctx context.Context, id string) error
	UpdateLastUsed(ctx context.Context, id primitive.ObjectID) error
	EnsureIndexes(ctx context.Context) error
}

type repository struct {
	collection *mongodriver.Collection
}

func NewRepository() Repository {
	var coll *mongodriver.Collection
	if mongodb.DB != nil {
		coll = mongodb.DB.Collection(collectionName)
	}
	return &repository{
		collection: coll,
	}
}

func (r *repository) getCollection() *mongodriver.Collection {
	if r.collection == nil && mongodb.DB != nil {
		r.collection = mongodb.DB.Collection(collectionName)
	}
	return r.collection
}

func (r *repository) Create(ctx context.Context, key *models.ApiKey) error {
	coll := r.getCollection()
	if coll == nil {
		return fmt.Errorf("mongodb not connected")
	}
	key.ID = primitive.NewObjectID()
	key.CreatedAt = time.Now()
	key.Enabled = true
	_, err := coll.InsertOne(ctx, key)
	return err
}

func (r *repository) FindAll(ctx context.Context) ([]models.ApiKey, error) {
	coll := r.getCollection()
	if coll == nil {
		return nil, fmt.Errorf("mongodb not connected")
	}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var keys []models.ApiKey
	if err := cursor.All(ctx, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *repository) FindByHash(ctx context.Context, keyHash string) (*models.ApiKey, error) {
	coll := r.getCollection()
	if coll == nil {
		return nil, fmt.Errorf("mongodb not connected")
	}
	var key models.ApiKey
	err := coll.FindOne(ctx, bson.M{"keyHash": keyHash, "enabled": true}).Decode(&key)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	coll := r.getCollection()
	if coll == nil {
		return fmt.Errorf("mongodb not connected")
	}
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = coll.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *repository) UpdateLastUsed(ctx context.Context, id primitive.ObjectID) error {
	coll := r.getCollection()
	if coll == nil {
		return fmt.Errorf("mongodb not connected")
	}
	now := time.Now()
	_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"lastUsedAt": now}})
	return err
}

func (r *repository) EnsureIndexes(ctx context.Context) error {
	coll := r.getCollection()
	if coll == nil {
		return fmt.Errorf("mongodb not connected")
	}
	_, err := coll.Indexes().CreateOne(ctx, mongodriver.IndexModel{
		Keys:    bson.D{{Key: "keyHash", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}
