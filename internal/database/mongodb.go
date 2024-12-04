package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"parkit-data-ETL/internal/models"
)

type MongoDB struct {
	client     *mongo.Client
	database   string
	collection *mongo.Collection
}

type Config struct {
	URI      string
	Database string
}

func Connect(config *Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.URI))
	if err != nil {
		return nil, err
	}

	// Create indexes
	collection := client.Database(config.Database).Collection("parking_meters")

	// Create a unique index on objectId
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "objectId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, err
	}

	// Create a 2dsphere index on location
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
	})
	if err != nil {
		return nil, err
	}

	return &MongoDB{
		client:     client,
		database:   config.Database,
		collection: collection,
	}, nil
}

func (m *MongoDB) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}

func (m *MongoDB) UpsertParkingMeters(meters []models.ParkingMeter) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var operations []mongo.WriteModel

	for _, meter := range meters {
		filter := bson.M{"objectId": meter.ObjectID}
		update := bson.M{"$set": meter}
		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
		operations = append(operations, model)
	}

	if len(operations) == 0 {
		return nil
	}

	_, err := m.collection.BulkWrite(ctx, operations)
	return err
}
