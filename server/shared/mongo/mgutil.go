package mgutil

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"coolcar/shared/mongo/objid"
)

const (
	IDFieldName       = "_id"
	UpdateAtFieldName = "updatedAt"
)

// IDField defines the ObjectID field
type IDField struct {
	ID primitive.ObjectID `bson:"_id"`
}

// UpdateAtField defines the updateAt field
type UpdateAtField struct {
	UpdatedAt int64 `bson="updated_at"`
}

// NewObjID generates a new object id.
var NewObjID = primitive.NewObjectID

// NewObjIDWithValue sets id for next objectID generation.
func NewObjIDWithValue(id fmt.Stringer) {
	NewObjID = func() primitive.ObjectID {
		return objid.MustFromID(id)
	}
}

// UpdateAt returns a value suitable for UpdatedAt field.
var UpdateAt = func() int64 {
	return time.Now().UnixNano()
}

// Set returns a $set update document
func Set(v interface{}) bson.M {
	return bson.M{
		"$set": v,
	}
}

// SetOnInsert returns a $setOnInsert update document
func SetOnInsert(v interface{}) bson.M {
	return bson.M{
		"$setOnInsert": v,
	}
}

// ZeroOrDoesNotExist generates a filter expression with
// field equal to zero or field dose not exist.
func ZeroOrDoesNotExist(field string, zero interface{}) bson.M {
	return bson.M{
		"$or": []bson.M{
			{
				field: zero,
			},
			{
				field: bson.M{
					"$exists": false,
				},
			},
		},
	}
}
