package objid

import (
	"fmt"

	"coolcar/shared/id"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FromID converts an id(string) to objected id
func FromID(id fmt.Stringer) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id.String())
}

// MustFromID converts an id(string) to objected id, panics on error
func MustFromID(id fmt.Stringer) primitive.ObjectID {
	objID, err := FromID(id)
	if err != nil {
		panic(err)
	}
	return objID
}

// ToAccountID converts object id to account id.
func ToAccountID(objID primitive.ObjectID) id.AccountID {
	return id.AccountID(objID.Hex())
}

// ToTripID converts object id to trip id.
func ToTripID(objID primitive.ObjectID) id.TripID {
	return id.TripID(objID.Hex())
}
