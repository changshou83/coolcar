package dao

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	tripField      = "trip"
	accountIDField = tripField + ".accountid"
	statusField    = tripField + ".status"
)

// Mongo defines a mongo dao.
type Mongo struct {
	collection *mongo.Collection
}

// NewMongo creates a mongo dao.
func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		collection: db.Collection("trip"),
	}
}

// TripRecord defines a trip record in mongo db.
type TripRecord struct {
	mgutil.IDField       `bson:"inline"`
	mgutil.UpdateAtField `bson:"inline"`
	Trip                 *rentalpb.Trip `bson:"trip"`
}

// CreateTrip create a trip.
func (m *Mongo) CreateTrip(
	c context.Context,
	trip *rentalpb.Trip,
) (*TripRecord, error) {
	record := &TripRecord{
		Trip: trip,
	}
	record.ID = mgutil.NewObjID()
	record.UpdatedAt = mgutil.UpdatedAt()

	_, err := m.collection.InsertOne(c, record)
	if err != nil {
		return nil, err // 直接返回错误，可能根据不同的错误返回不同的错误码
	}

	return record, nil
}

// GetTrip gets a trip.
func (m *Mongo) GetTrip(c context.Context, id id.TripID, accountID id.AccountID) (*TripRecord, error) {
	objID, err := objid.FromID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %v", err)
	}
	res := m.collection.FindOne(c, bson.M{
		mgutil.IDFieldName: objID,
		accountIDField:     accountID,
	})

	if err := res.Err(); err != nil {
		return nil, err
	}

	var tr TripRecord
	err = res.Decode(&tr)
	if err != nil {
		return nil, fmt.Errorf("cannot decode: %v", err)
	}
	return &tr, nil
}

// GetTrips gets trips for the account by status.
// If status is not specified, gets all trips for the account.
// func (m *Mongo) GetTrips(c context.Context, accountID id.AccountID, status rentalpb.TripStatus) ([]*TripRecord, error) {
// 	filter := bson.M{
// 		accountIDField: accountID.String(),
// 	}
// 	if status != rentalpb.TripStatus_TS_NOT_SPECIFIED {
// 		filter[statusField] = status
// 	}

// 	res, err := m.collection.Find(c, filter, options.Find().SetSort(bson.M{
// 		mgutil.IDFieldName: -1,
// 	}))
// 	if err != nil {
// 		return nil, err
// 	}

// 	var trips []*TripRecord
// 	for res.Next(c) {
// 		var trip TripRecord
// 		err := res.Decode(&trip)
// 		if err != nil {
// 			return nil, err
// 		}
// 		trips = append(trips, &trip)
// 	}
// 	return trips, nil
// }

// GetTrips gets trips for the account by id list,
// If id list is empty, gets all trips for the account.
func (m *Mongo) GetTrips(
	c context.Context,
	accountID id.AccountID,
	idList []id.TripID,
) ([]*TripRecord, error) {
	filter := bson.M{
		accountIDField: accountID.String(),
	}
	if len(idList) != 0 {
		var tripIDList []primitive.ObjectID
		for _, id := range idList {
			tripID, err := objid.FromID(id)
			if err != nil {
				return nil, fmt.Errorf("invalid id: %v", err)
			}
			tripIDList = append(tripIDList, tripID)
		}
		filter[mgutil.IDFieldName] = mgutil.In(tripIDList)
	}
	res, err := m.collection.Find(c, filter, options.Find().SetSort(bson.M{
		mgutil.IDFieldName: -1,
	}))
	if err != nil {
		return nil, err
	}

	var trips []*TripRecord
	for res.Next(c) {
		var trip TripRecord
		err := res.Decode(&trip)
		if err != nil {
			return nil, err
		}
		trips = append(trips, &trip)
	}
	return trips, nil
}

// UpdateTrip update a trip.
func (m *Mongo) UpdateTrip(
	c context.Context,
	tripID id.TripID,
	accountID id.AccountID,
	updatedAt int64,
	trip *rentalpb.Trip,
) error {
	objID, err := objid.FromID(tripID)
	if err != nil {
		return fmt.Errorf("invalid id: %v", err)
	}

	// 使用 UpdatedAt 作为标识实现乐观锁，如果在并发更新时，当前资源已被更新，那么就直接查不到资源，保证了资源的一致性和完整性
	newUpdatedAt := mgutil.UpdatedAt()
	// UpdateOne 具有原子性， FindOneAndUpdate 不具有原子性
	res, err := m.collection.UpdateOne(c, bson.M{
		mgutil.IDFieldName:        objID,
		accountIDField:            accountID.String(),
		mgutil.UpdatedAtFieldName: updatedAt,
	}, mgutil.Set(bson.M{
		tripField:                 trip,
		mgutil.UpdatedAtFieldName: newUpdatedAt,
	}))
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
