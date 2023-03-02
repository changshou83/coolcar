package dao

import (
	"context"
	"fmt"

	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const openIDField = "open_id"

// Mongo defines a mongodb dao
type Mongo struct {
	collection *mongo.Collection
}

// NewMongo creates a mongodb dao
func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		collection: db.Collection("account"),
	}
}

// ResolveAccountID resolves the account id from open id.
func (m *Mongo) ResolveAccountID(
	ctx context.Context,
	openID string,
) (id.AccountID, error) {
	insertedID := mgutil.NewObjID()
	res := m.collection.FindOneAndUpdate(ctx, bson.M{
		openIDField: openID,
	}, mgutil.SetOnInsert(bson.M{
		openIDField:        openID,
		mgutil.IDFieldName: insertedID,
	}), options.
		FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After),
	)
	if err := res.Err(); err != nil {
		return "", fmt.Errorf("cannot find one and update: %v", err)
	}

	var row mgutil.IDField
	err := res.Decode(&row)
	if err != nil {
		return "", fmt.Errorf("cannot decode result: %v", err)
	}

	return objid.ToAccountID(row.ID), nil
}
