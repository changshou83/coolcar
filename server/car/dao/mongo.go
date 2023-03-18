package dao

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	carField      = "car"
	statusField   = carField + ".status"
	driverField   = carField + ".driver"
	positionField = carField + ".position"
	tripIDField   = carField + ".tripid"
)

type Mongo struct {
	collection *mongo.Collection
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		collection: db.Collection("car"),
	}
}

type CarRecord struct {
	mgutil.IDField `bson:"inline"`
	Car            *carpb.Car `bson:"car"`
}

func (m *Mongo) CreateCar(c context.Context) (*CarRecord, error) {
	record := &CarRecord{
		Car: &carpb.Car{
			Position: &carpb.Location{
				Latitude:  30,
				Longitude: 120,
			},
			Status: carpb.CarStatus_LOCKED,
		},
	}

	record.ID = mgutil.NewObjID()
	_, err := m.collection.InsertOne(c, record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (m *Mongo) GetCar(c context.Context, cid id.CarID) (*CarRecord, error) {
	id, err := objid.FromID(cid)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %v", err)
	}
	return convertSingleResult(m.collection.FindOne(c, bson.M{
		mgutil.IDFieldName: id,
	}))
}

func (m *Mongo) GetCars(c context.Context) ([]*CarRecord, error) {
	filter := bson.M{}
	res, err := m.collection.Find(c, filter, options.Find())
	if err != nil {
		return nil, err
	}

	var cars []*CarRecord
	for res.Next(c) {
		var record CarRecord
		err := res.Decode(&record)
		if err != nil {
			return nil, err
		}
		cars = append(cars, &record)
	}
	return cars, nil
}

// CarUpdate defines updates to a car.
type CarUpdate struct {
	Status       carpb.CarStatus
	Position     *carpb.Location
	Driver       *carpb.Driver
	TripID       id.TripID
	UpdateTripID bool
}

// UpdateCar updates a car.
// if status is specified, it updates the car
// only when existing record matches the status specified.
func (m *Mongo) UpdateCar(
	c context.Context,
	cid id.CarID,
	prevStatus carpb.CarStatus,
	update *CarUpdate,
) (*CarRecord, error) {
	id, err := objid.FromID(cid)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %v", err)
	}
	// create filter
	filter := bson.M{
		mgutil.IDFieldName: id,
	}
	if prevStatus != carpb.CarStatus_CS_NOT_SPECIFIED {
		filter[statusField] = prevStatus
	}
	// create update
	data := bson.M{}
	if update.Status != carpb.CarStatus_CS_NOT_SPECIFIED {
		data[statusField] = update.Status
	}
	if update.Driver != nil {
		data[driverField] = update.Driver
	}
	if update.Position != nil {
		data[positionField] = update.Position
	}
	if update.UpdateTripID {
		data[tripIDField] = update.TripID.String()
	}
	// find
	res := m.collection.FindOneAndUpdate(
		c, filter, mgutil.Set(data),
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	return convertSingleResult(res)
}

// convertSingleResult convert single result to a car record.
func convertSingleResult(res *mongo.SingleResult) (*CarRecord, error) {
	if err := res.Err(); err != nil {
		return nil, err
	}

	var record CarRecord
	err := res.Decode(&record)
	if err != nil {
		return nil, fmt.Errorf("cannot decode: %v", err)
	}
	return &record, nil
}
