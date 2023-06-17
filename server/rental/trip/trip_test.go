package trip

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	locdesc "coolcar/rental/trip/client/locDesc"
	"coolcar/rental/trip/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/mongotesting"
	"coolcar/shared/server"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
)

func TestCreateTrip(t *testing.T) {
	c := context.Background()

	pm := &profileManager{}
	cm := &carManager{}
	s := newService(c, t, pm, cm)

	nowFunc = func() int64 {
		return 1605695246
	}
	req := &rentalpb.CreateTripRequest{
		CarId: "car1",
		Start: &rentalpb.Location{
			Latitude:  32,
			Longitude: 114,
		},
	}

	pm.iID = "identity1"
	golden := `{"account_id":%q,"car_id":"car1","identity_id":"identity1","status":1,"start":{"location":{"latitude":32,"longitude":114},"fee_cent":7000,"km_driven":200,"loc_desc":"天安门","timestamp_sec":10000},"current":{"location":{"latitude":32,"longitude":114},"fee_cent":7000,"km_driven":200,"loc_desc":"天安门","timestamp_sec":10000}}`
	cases := []struct {
		name         string
		now          int64
		accountID    string
		tripID       string
		profileErr   error
		carVerifyErr error
		carUnlockErr error
		want         string
		wantErr      bool
	}{
		{
			name:      "normal_create",
			now:       10000,
			accountID: "account1",
			tripID:    "5f8132eb12714bf629489054",
			want:      fmt.Sprintf(golden, "account1"),
		},
		{
			name:       "profile_err",
			now:        10000,
			accountID:  "account2",
			tripID:     "5f8132eb12714bf629489055",
			profileErr: fmt.Errorf("profile"),
			wantErr:    true,
		},
		{
			name:         "car_verify_err",
			now:          10000,
			accountID:    "account3",
			tripID:       "5f8132eb12714bf629489056",
			carVerifyErr: fmt.Errorf("verify"),
			wantErr:      true,
		},
		{
			name:         "car_unlock_err",
			now:          10000,
			accountID:    "account4",
			tripID:       "5f8132eb12714bf629489057",
			carUnlockErr: fmt.Errorf("unlock"),
			want:         fmt.Sprintf(golden, "account4"),
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			mgutil.NewObjIDWithValue(id.TripID(cc.tripID))
			nowFunc = func() int64 {
				return cc.now
			}
			pm.err = cc.profileErr
			cm.unlockErr = cc.carUnlockErr
			cm.verifyErr = cc.carVerifyErr

			cwa := auth.ContextWithAccountID(context.Background(), id.AccountID(cc.accountID))
			res, err := s.CreateTrip(cwa, req)
			// 创建失败
			if cc.wantErr {
				if err == nil {
					t.Errorf("want error; got none")
				} else {
					return
				}
			}
			if err != nil {
				t.Errorf("error creating trip: %v", err)
			}
			if res.Id != cc.tripID {
				t.Errorf("incorrect id; want %q, got %q", cc.tripID, res.Id)
			}
			// 解析 trip 失败
			b, err := json.Marshal(res.Trip)
			if err != nil {
				t.Errorf("cannot marshall response: %v", err)
			}
			got := string(b)
			if cc.want != got {
				t.Errorf("incorrect response: want %q, got %q", cc.want, got)
			}
		})
	}
}

func TestTripLifecycle(t *testing.T) {
	c := auth.ContextWithAccountID(
		context.Background(), id.AccountID("account_for_lifecycle"))
	s := newService(c, t, &profileManager{}, &carManager{})

	tid := id.TripID("5f8132eb22714bf629489056")
	mgutil.NewObjIDWithValue(tid)
	cases := []struct {
		name    string
		now     int64
		op      func() (*rentalpb.Trip, error)
		want    string
		wantErr bool
	}{
		{
			name: "create_trip",
			now:  10000,
			op: func() (*rentalpb.Trip, error) {
				e, err := s.CreateTrip(c, &rentalpb.CreateTripRequest{
					CarId: "car1",
					Start: &rentalpb.Location{
						Latitude:  32.123,
						Longitude: 114.2525,
					},
				})
				if err != nil {
					return nil, err
				}
				return e.Trip, nil
			},
			want: `{"account_id":"account_for_lifecycle","car_id":"car1","start":{"location":{"latitude":32.123,"longitude":114.2525},"poi_name":"天安门","timestamp_sec":10000},"current":{"location":{"latitude":32.123,"longitude":114.2525},"poi_name":"天安门","timestamp_sec":10000},"status":1}`,
		},
		{
			name: "update_trip",
			now:  20000,
			op: func() (*rentalpb.Trip, error) {
				return s.UpdateTrip(c, &rentalpb.UpdateTripRequest{
					Id: tid.String(),
					Current: &rentalpb.Location{
						Latitude:  28.234234,
						Longitude: 123.243255,
					},
				})
			},
			want: `{"account_id":"account_for_lifecycle","car_id":"car1","start":{"location":{"latitude":32.123,"longitude":114.2525},"poi_name":"天安门","timestamp_sec":10000},"current":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":10677,"km_driven":100,"poi_name":"中关村","timestamp_sec":20000},"status":1}`,
		},
		{
			name: "finish_trip",
			now:  30000,
			op: func() (*rentalpb.Trip, error) {
				return s.UpdateTrip(c, &rentalpb.UpdateTripRequest{
					Id:      tid.String(),
					EndTrip: true,
				})
			},
			want: `{"account_id":"account_for_lifecycle","car_id":"car1","start":{"location":{"latitude":32.123,"longitude":114.2525},"poi_name":"天安门","timestamp_sec":10000},"current":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":24674,"km_driven":100,"poi_name":"中关村","timestamp_sec":30000},"end":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":24674,"km_driven":100,"poi_name":"中关村","timestamp_sec":30000},"status":2}`,
		},
		{
			name: "query_trip",
			now:  40000,
			op: func() (*rentalpb.Trip, error) {
				return s.GetTrip(c, &rentalpb.GetTripRequest{
					Id: tid.String(),
				})
			},
			want: `{"account_id":"account_for_lifecycle","car_id":"car1","start":{"location":{"latitude":32.123,"longitude":114.2525},"poi_name":"天安门","timestamp_sec":10000},"current":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":24674,"km_driven":100,"poi_name":"中关村","timestamp_sec":30000},"end":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":24674,"km_driven":100,"poi_name":"中关村","timestamp_sec":30000},"status":2}`,
		},
		{
			name: "update_after_finished",
			now:  50000,
			op: func() (*rentalpb.Trip, error) {
				return s.UpdateTrip(c, &rentalpb.UpdateTripRequest{
					Id: tid.String(),
				})
			},
			wantErr: true,
		},
	}
	rand.Seed(1345)
	for _, cc := range cases {
		nowFunc = func() int64 {
			return cc.now
		}
		trip, err := cc.op()
		if cc.wantErr {
			if err == nil {
				t.Errorf("%s: want error; got none", cc.name)
			} else {
				continue
			}
		}
		if err != nil {
			t.Errorf("%s: operation failed: %v", cc.name, err)
			continue
		}
		b, err := json.Marshal(trip)
		if err != nil {
			t.Errorf("%s: failed marshalling response: %v", cc.name, err)
		}
		got := string(b)
		if cc.want != got {
			t.Errorf("%s: incorrect response; want: %s, got: %s", cc.name, cc.want, got)
		}
	}
}

// func TestGetTrips(t *testing.T) {
// 	rows := []struct {
// 		id        string
// 		accountID string
// 	}{
// 		{
// 			id:        "5f8132eb10714bf629489051",
// 			accountID: "account_for_get_trips",
// 		},
// 		{
// 			id:        "5f8132eb10714bf629489055",
// 			accountID: "account_for_get_trips_1",
// 		},
// 	}

// 	c := auth.ContextWithAccountID(context.Background(), id.AccountID("account_for_get_trips"))
// 	s := newService(c, t, &profileManager{}, &carManager{})
// 	for _, r := range rows {
// 		mgutil.NewObjIDWithValue(id.TripID(r.id))
// 		nowFunc = func() int64 {
// 			return 10000
// 		}

// 		cwa := auth.ContextWithAccountID(c, id.AccountID(r.accountID))
// 		_, err := s.CreateTrip(cwa, &rentalpb.CreateTripRequest{
// 			CarId: "car1",
// 			Start: &rentalpb.Location{
// 				Latitude:  32,
// 				Longitude: 114,
// 			},
// 		})
// 		if err != nil {
// 			t.Fatalf("cannot create trip: %v", err)
// 		}
// 	}

// 	cases := []struct {
// 		name      string
// 		accountID string
// 		idList    []string
// 		wantCount int
// 		want      string
// 	}{
// 		{
// 			name:      "get_5f8132eb10714bf629489051",
// 			accountID: "account_for_get_trips",
// 			idList:    []string{"5f8132eb10714bf629489051"},
// 			wantCount: 1,
// 			want:      `{"id":"5f8132eb10714bf629489051","trip":{"account_id":"account_for_get_trips","car_id":"car1","status":1,"start":{"location":{"latitude":32,"longitude":114},"fee_cent":7000,"km_driven":200,"loc_desc":"天安门","timestamp_sec":10000},"current":{"location":{"latitude":32,"longitude":114},"fee_cent":7000,"km_driven":200,"loc_desc":"天安门","timestamp_sec":10000}}}`,
// 		},
// 	}

// 	for _, cc := range cases {
// 		t.Run(cc.name, func(t *testing.T) {
// 			c := auth.ContextWithAccountID(context.Background(), id.AccountID(cc.accountID))
// 			res, err := s.GetTrips(
// 				c,
// 				&rentalpb.GetTripsRequest{
// 					IdList: cc.idList,
// 				},
// 			)
// 			if err != nil {
// 				t.Errorf("cannot get trips: %v", err)
// 			}
// 			if cc.wantCount != len(res.Trips) {
// 				t.Errorf("incorrect count: want:%d, got: %d", cc.wantCount, len(res.Trips))
// 			}
// 			b, err := json.Marshal(res.Trips[0])
// 			if err != nil {
// 				t.Errorf("failed marshalling response: %v", err)
// 			}
// 			if cc.want != string(b) {
// 				t.Errorf("incorrect response; want: %s, got: %s", cc.want, string(b))
// 			}
// 		})
// 	}
// }

// func TestTripLifecycle(t *testing.T) {
// 	c := auth.ContextWithAccountID(
// 		context.Background(), id.AccountID("account_for_lifecycle"))
// 	s := newService(c, t, &profileManager{}, &carManager{})

// 	tid := id.TripID("5f8132eb22714bf629489056")
// 	mgutil.NewObjIDWithValue(tid)
// 	cases := []struct {
// 		name    string
// 		now     int64
// 		op      func() (*rentalpb.Trip, error)
// 		want    string
// 		wantErr bool
// 	}{
// 		{
// 			name: "create_trip",
// 			now:  10000,
// 			op: func() (*rentalpb.Trip, error) {
// 				e, err := s.CreateTrip(c, &rentalpb.CreateTripRequest{
// 					CarId: "car1",
// 					Start: &rentalpb.Location{
// 						Latitude:  32.123,
// 						Longitude: 114.2525,
// 					},
// 				})
// 				if err != nil {
// 					return nil, err
// 				}
// 				return e.Trip, nil
// 			},
// 			want: `{"account_id":"account_for_lifecycle","car_id":"car1","status":1,"start":{"location":{"latitude":32.123,"longitude":114.2525},"fee_cent":7000,"km_driven":200,"loc_desc":"天安门","timestamp_sec":10000},"current":{"location":{"latitude":32.123,"longitude":114.2525},"fee_cent":7000,"km_driven":200,"loc_desc":"天安门","timestamp_sec":10000}}`,
// 		},
// 		{
// 			name: "update_trip",
// 			now:  20000,
// 			op: func() (*rentalpb.Trip, error) {
// 				return s.UpdateTrip(c, &rentalpb.UpdateTripRequest{
// 					Id: tid.String(),
// 					Current: &rentalpb.Location{
// 						Latitude:  28.234234,
// 						Longitude: 123.243255,
// 					},
// 				})
// 			},
// 			want: `{"account_id":"account_for_lifecycle","car_id":"car1","status":1,"start":{"location":{"latitude":32.123,"longitude":114.2525},"fee_cent":7000,"km_driven":200,"loc_desc":"天安门","timestamp_sec":10000},"current":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":14000,"km_driven":400,"loc_desc":"中关村","timestamp_sec":20000}}`,
// 		},
// 		{
// 			name: "finish_trip",
// 			now:  30000,
// 			op: func() (*rentalpb.Trip, error) {
// 				return s.UpdateTrip(c, &rentalpb.UpdateTripRequest{
// 					Id:      tid.String(),
// 					EndTrip: true,
// 				})
// 			},
// 			want: `{"account_id":"account_for_lifecycle","car_id":"car1","status":2,"start":{"location":{"latitude":32.123,"longitude":114.2525},"fee_cent":7000,"km_driven":200,"loc_desc":"天安门","timestamp_sec":10000},"end":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":21000,"km_driven":600,"loc_desc":"中关村","timestamp_sec":30000},"current":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":21000,"km_driven":600,"loc_desc":"中关村","timestamp_sec":30000}}`,
// 		},
// 		{
// 			name: "query_trip",
// 			now:  40000,
// 			op: func() (*rentalpb.Trip, error) {
// 				res, err := s.GetTrips(c, &rentalpb.GetTripsRequest{
// 					IdList: []string{tid.String()},
// 				})
// 				if err != nil {
// 					return nil, err
// 				}
// 				return res.Trips[0].Trip, nil
// 			},
// 			want: `{"account_id":"account_for_lifecycle","car_id":"car1","status":2,"start":{"location":{"latitude":32.123,"longitude":114.2525},"fee_cent":7000,"km_driven":200,"loc_desc":"天安门","timestamp_sec":10000},"end":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":21000,"km_driven":600,"loc_desc":"中关村","timestamp_sec":30000},"current":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":21000,"km_driven":600,"loc_desc":"中关村","timestamp_sec":30000}}`,
// 		},
// 		{
// 			name: "update_after_finished",
// 			now:  50000,
// 			op: func() (*rentalpb.Trip, error) {
// 				return s.UpdateTrip(c, &rentalpb.UpdateTripRequest{
// 					Id: tid.String(),
// 				})
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	rand.Seed(1345)
// 	for _, cc := range cases {
// 		nowFunc = func() int64 {
// 			return cc.now
// 		}
// 		trip, err := cc.op()
// 		if cc.wantErr {
// 			if err == nil {
// 				t.Errorf("%s: want error; got none", cc.name)
// 			} else {
// 				continue
// 			}
// 		}
// 		if err != nil {
// 			t.Errorf("%s: operation failed: %v", cc.name, err)
// 			continue
// 		}
// 		b, err := json.Marshal(trip)
// 		if err != nil {
// 			t.Errorf("%s: failed marshalling response: %v", cc.name, err)
// 		}
// 		got := string(b)
// 		if cc.want != got {
// 			t.Errorf("%s: incorrect response; want: %s, got: %s", cc.name, cc.want, got)
// 		}
// 	}
// }

/* 测试用 */

func newService(c context.Context, t *testing.T, pm ProfileManager, cm CarManager) *Service {
	mc, err := mongotesting.NewClient(c)
	if err != nil {
		t.Fatalf("cannot create mongo client: %v", err)
	}

	logger, err := server.NewZapLogger()
	if err != nil {
		t.Fatalf("cannot create logger: %v", err)
	}

	db := mc.Database("coolcar")
	mongotesting.SetupIndexes(c, db)
	return &Service{
		ProfileManager: pm,
		CarManager:     cm,
		LocDescManager: &locdesc.Manager{},
		Mongo:          dao.NewMongo(db),
		Logger:         logger,
	}
}

type profileManager struct {
	iID id.IdentityID
	err error
}

func (p *profileManager) Verify(context.Context, id.AccountID) (id.IdentityID, error) {
	return p.iID, p.err
}

type carManager struct {
	verifyErr error
	unlockErr error
}

func (m *carManager) Verify(context.Context, id.CarID, *rentalpb.Location) error {
	return m.verifyErr
}

func (m *carManager) Unlock(c context.Context, cid id.CarID, aid id.AccountID, tid id.TripID, avatarURL string) error {
	return m.unlockErr
}

func (m *carManager) Lock(c context.Context, cid id.CarID) error {
	return nil
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m))
}
