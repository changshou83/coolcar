package dao

import (
	"context"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/mongotesting"
	"coolcar/shared/mongo/objid"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestResolveAccountID(t *testing.T) {
	c := context.Background()
	mc, err := mongotesting.NewClient(c)
	if err != nil {
		t.Errorf("cannot connect mongodb: %v", err)
	}

	mdao := NewMongo(mc.Database("coolcar"))
	if err != nil {
		panic(err)
	}
	// 初始数据
	_, err = mdao.collection.InsertMany(c, []interface{}{
		bson.M{
			mgutil.IDFieldName: objid.MustFromID(id.AccountID("5f7c245ab0361e00ffb9fd6f")),
			openIDField:        "openid_1",
		},
		bson.M{
			mgutil.IDFieldName: objid.MustFromID(id.AccountID("5f7c245ab0361e00ffb9fd70")),
			openIDField:        "openid_2",
		},
	})
	if err != nil {
		t.Errorf("cannot insert initial values: %v", err)
	}
	// 测试数据
	cases := []struct {
		name   string
		openID string
		want   string
	}{
		{
			name:   "existing_user",
			openID: "openid_1",
			want:   "5f7c245ab0361e00ffb9fd6f",
		},
		{
			name:   "another_existing_user",
			openID: "openid_2",
			want:   "5f7c245ab0361e00ffb9fd70",
		},
		{
			name:   "new_user",
			openID: "openid_3",
			want:   "5f7c245ab0361e00ffb9fd71",
		},
	}

	// set new user id
	mgutil.NewObjIDWithValue(id.AccountID("5f7c245ab0361e00ffb9fd71"))

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			id, err := mdao.ResolveAccountID(context.Background(), cc.openID)
			if err != nil {
				t.Errorf("failed resolve account id for %q: %v", cc.openID, err)
			}

			if id.String() != cc.want {
				t.Errorf("resolve account id: want: %q; got: %q", cc.want, id)
			}
		})
	}
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m))
}
