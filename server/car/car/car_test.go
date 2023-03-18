package car

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/car/dao"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/mongotesting"
	"coolcar/shared/server"
	"encoding/json"
	"os"
	"testing"
)

func TestCarUpdate(t *testing.T) {
	c := context.Background()
	mc, err := mongotesting.NewClient(c)
	if err != nil {
		t.Errorf("cannot create mongo client: %v", err)
	}

	logger, err := server.NewZapLogger()
	if err != nil {
		t.Fatalf("cannot create logger: %v", err)
	}

	s := &Service{
		Mongo:     dao.NewMongo(mc.Database("coolcar")),
		Logger:    logger,
		Publisher: &testPublisher{},
	}

	carID := id.CarID("5f8132eb22814bf629489056")
	mgutil.NewObjIDWithValue(carID)
	_, err = s.CreateCar(c, &carpb.CreateCarRequest{})
	if err != nil {
		t.Fatalf("cannot create car: %v", err)
	}
	// 先创建，再开锁，最后锁车
	cases := []struct {
		name    string
		fn      func() error
		want    string
		wantErr bool
	}{
		{
			name: "get_car",
			fn: func() error {
				return nil
			},
			want: `{"status":1,"position":{"latitude":30,"longitude":120}}`,
		},
		{
			name: "unlock_car",
			fn: func() error {
				_, err := s.UnlockCar(c, &carpb.UnlockCarRequest{
					Id: carID.String(),
					Driver: &carpb.Driver{
						Id:        "test_driver",
						AvatarUrl: "test_driver",
					},
					TripId: "test_trip",
				})
				return err
			},
			want: `{"status":2,"driver":{"id":"test_driver","avatar_url":"test_driver"},"position":{"latitude":30,"longitude":120},"trip_id":"test_trip"}`,
		},
		{
			name: "unlock_complete",
			fn: func() error {
				_, err := s.UpdateCar(c, &carpb.UpdateCarRequest{
					Id:     carID.String(),
					Status: carpb.CarStatus_UNLOCKED,
					Position: &carpb.Location{
						Latitude:  31,
						Longitude: 121,
					},
				})
				return err
			},
			want: `{"status":3,"driver":{"id":"test_driver","avatar_url":"test_driver"},"position":{"latitude":31,"longitude":121},"trip_id":"test_trip"}`,
		},
		{
			name: "unlock_by_another_driver",
			fn: func() error {
				_, err := s.UnlockCar(c, &carpb.UnlockCarRequest{
					Id: carID.String(),
					Driver: &carpb.Driver{
						Id:        "another_test_driver",
						AvatarUrl: "test_driver",
					},
					TripId: "another_test_trip",
				})
				return err
			},
			wantErr: true,
		},
		{
			name: "lock_car",
			fn: func() error {
				_, err := s.LockCar(c, &carpb.LockCarRequest{
					Id: carID.String(),
				})
				return err
			},
			want: `{"status":4,"driver":{"id":"test_driver","avatar_url":"test_driver"},"position":{"latitude":31,"longitude":121},"trip_id":"test_trip"}`,
		},
		{
			name: "lock_complete",
			fn: func() error {
				_, err := s.UpdateCar(c, &carpb.UpdateCarRequest{
					Id:     carID.String(),
					Status: carpb.CarStatus_LOCKED,
				})
				return err
			},
			want: `{"status":1,"driver":{},"position":{"latitude":31,"longitude":121}}`,
		},
	}

	for _, cc := range cases {
		err := cc.fn()
		if cc.wantErr {
			if err == nil {
				t.Errorf("%s: want error; got none.", cc.name)
			} else {
				continue
			}
		}
		if err != nil {
			t.Errorf("%s: operation failed: %v", cc.name, err)
			continue
		}

		car, err := s.GetCar(c, &carpb.GetCarRequest{
			Id: carID.String(),
		})
		if err != nil {
			t.Errorf("%s: cannot get car after operation: %v", cc.name, err)
		}

		bytes, err := json.Marshal(car)
		if err != nil {
			t.Errorf("%s: failed marshal response: %v", cc.name, err)
		}
		got := string(bytes)
		if got != cc.want {
			t.Errorf("%s: get incorrect response: want: %q, got: %q", cc.name, cc.want, got)
		}
	}
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m))
}

type testPublisher struct{}

func (p *testPublisher) Publish(context.Context, *carpb.CarEntity) error {
	return nil
}
