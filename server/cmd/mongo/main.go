package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	c := context.Background()
	mc, err := mongo.Connect(c, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("cannot connect mongodb: %v", err)
	}
	defer mc.Disconnect(c)

	collection := mc.Database("coolcar").Collection("account")
	findRows(c, collection)
}

func findRows(c context.Context, col *mongo.Collection) {
	cur, err := col.Find(c, bson.M{})
	if err != nil {
		log.Fatalf("cannot find cur row: %v", err)
	}

	for cur.Next(c) {
		var row struct {
			ID     primitive.ObjectID `bson:"_id"`
			OpenID string             `bson:"open_id"`
		}
		err = cur.Decode(&row)
		if err != nil {
			log.Fatalf("cannot decode: %v", err)
		}
		fmt.Printf("%+v\n", row)
	}
}

func insertRows(c context.Context, col *mongo.Collection) {
	res, err := col.InsertMany(c, []interface{}{
		bson.M{
			"open_id": "123",
		},
		bson.M{
			"open_id": "456",
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", res)
}
