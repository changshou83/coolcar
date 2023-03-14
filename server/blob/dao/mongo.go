package dao

import (
	"context"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Mongo struct {
	collection *mongo.Collection
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		collection: db.Collection("blob"),
	}
}

type BlobRecord struct {
	mgutil.IDField `bson:"inline"`
	AccountID      string `bson:"accountid"`
	Path           string `bson:"path"`
}

func (m *Mongo) CreateBlob(c context.Context, aid id.AccountID) (*BlobRecord, error) {
	record := &BlobRecord{
		AccountID: aid.String(),
	}
	blobID := mgutil.NewObjID()
	record.ID = blobID
	record.Path = fmt.Sprintf("%s/%s", aid.String(), blobID.Hex())

	_, err := m.collection.InsertOne(c, record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (m *Mongo) GetBlob(c context.Context, bid id.BlobID) (*BlobRecord, error) {
	blobID, err := objid.FromID(bid)
	if err != nil {
		return nil, fmt.Errorf("invalid object id: %v", err)
	}

	res := m.collection.FindOne(c, bson.M{
		mgutil.IDFieldName: blobID,
	})
	if err := res.Err(); err != nil {
		return nil, err
	}

	var record BlobRecord
	err = res.Decode(&record)
	if err != nil {
		return nil, fmt.Errorf("cannot decode result: %v", err)
	}
	return &record, nil
}
