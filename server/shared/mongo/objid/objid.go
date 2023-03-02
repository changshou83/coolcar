package objid

import (
	"fmt"

	"coolcar/shared/id"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func FromID(id fmt.Stringer) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id.String())
}

func MustFromID(id fmt.Stringer) primitive.ObjectID {
	objID, err := FromID(id)
	if err != nil {
		panic(err)
	}
	return objID
}

func ToAccountID(objID primitive.ObjectID) id.AccountID {
	return id.AccountID(objID.Hex())
}

func ToTripID(objID primitive.ObjectID) id.TripID {
	return id.TripID(objID.Hex())
}
