package dao

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	accountIDField      = "accountid"
	profileField        = "profile"
	identityStatusField = profileField + ".status"
	// photoBlobIDField    = "photoblobid"
)

type Mongo struct {
	collection *mongo.Collection
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		collection: db.Collection("profile"),
	}
}

type ProfileRecord struct {
	AccountID string            `bson:"accountid"`
	Profile   *rentalpb.Profile `bson:"profile"`
}

func (m *Mongo) GetProfile(
	c context.Context,
	aid id.AccountID,
) (*ProfileRecord, error) {
	res := m.collection.FindOne(c, bson.M{
		accountIDField: aid.String(),
	})
	if err := res.Err(); err != nil {
		return nil, err
	}
	var record ProfileRecord
	err := res.Decode(record)
	if err != nil {
		return nil, fmt.Errorf("cannot decode profile: %v", err)
	}
	return &record, nil
}

func (m *Mongo) UpdateProfile(
	c context.Context,
	aid id.AccountID,
	prevState rentalpb.IdentityStatus,
	profile *rentalpb.Profile,
) error {
	// create filter
	filter := bson.M{
		identityStatusField: prevState,
	}
	if prevState == rentalpb.IdentityStatus_UNSUBMITTED {
		filter = mgutil.ZeroOrDoesNotExist(identityStatusField, prevState)
	}
	filter[accountIDField] = aid.String()
	// create || update
	_, err := m.collection.UpdateOne(c, filter,
		mgutil.Set(bson.M{
			accountIDField: aid.String(),
			profileField:   profile,
		}),
		options.Update().SetUpsert(true),
	)
	return err
}

// func (m *Mongo) updateProfilePhoto(
// 	c context.Context,
// 	aid id.AccountID,
// 	bid id.BlobID,
// ) error {
// 	_, err := m.collection.UpdateOne(c, bson.M{
// 		accountIDField: aid.String(),
// 	}, mgutil.Set(bson.M{
// 		accountIDField:   aid.String(),
// 		photoBlobIDField: bid.String(),
// 	}), options.Update().SetUpsert(true))
// 	return err
// }
