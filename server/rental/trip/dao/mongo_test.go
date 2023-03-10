package dao

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/mongotesting"
	"os"
	"testing"
)

func TestCreateTrip(t *testing.T) {
	c := context.Background()
	mc, err := mongotesting.NewClient(c)
	if err != nil {
		t.Errorf("cannot connect mongodb: %v", err)
	}

	db := mc.Database("coolcar")
	err = mongotesting.SetupIndexes(c, db)
	if err != nil {
		t.Fatalf("cannot setup indexes: %v", err)
	}
	m := NewMongo(db)

	if err != nil {
		panic(err)
	}
	cases := []struct {
		name       string
		tripID     string
		accountID  string
		tripStatus rentalpb.TripStatus
		wantErr    bool
	}{
		{
			name:       "finished",
			tripID:     "5f8132eb00714bf62948905c",
			accountID:  "account1",
			tripStatus: rentalpb.TripStatus_FINISHED,
		},
		{
			name:       "another_finished",
			tripID:     "5f8132eb00714bf62948905d",
			accountID:  "account1",
			tripStatus: rentalpb.TripStatus_FINISHED,
		},
		{
			name:       "in_progress",
			tripID:     "5f8132eb00714bf62948905e",
			accountID:  "account1",
			tripStatus: rentalpb.TripStatus_IN_PROGRESS,
		},
		{
			name:       "another_in_progress",
			tripID:     "5f8132eb00714bf62948905f",
			accountID:  "account1",
			tripStatus: rentalpb.TripStatus_IN_PROGRESS,
			wantErr:    true,
		},
		{
			name:       "in_progress_by_another_account",
			tripID:     "5f8132eb00714bf629489060",
			accountID:  "account2",
			tripStatus: rentalpb.TripStatus_IN_PROGRESS,
		},
	}

	for _, cc := range cases {
		mgutil.NewObjIDWithValue(id.TripID(cc.tripID))
		trip, err := m.CreateTrip(context.Background(), &rentalpb.Trip{
			AccountId: cc.accountID,
			Status:    cc.tripStatus,
		})
		if cc.wantErr {
			if err == nil {
				t.Errorf("%s: error expected; got none", cc.name)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: error creating trip: %v", cc.name, err)
			continue
		}
		if trip.ID.Hex() != cc.tripID {
			t.Errorf("%s: incorrect trip id; want: %q; got: %q",
				cc.name, cc.tripID, trip.ID.Hex())
		}
	}
}

func TestGetTrips(t *testing.T) {
	rows := []struct {
		id        string
		accountID string
	}{
		{
			id:        "5f8132eb10714bf629489051",
			accountID: "account_id_for_get_trips",
		},
		{
			id:        "5f8132eb10714bf629489052",
			accountID: "account_id_for_get_trips",
		},
		{
			id:        "5f8132eb10714bf629489055",
			accountID: "account_id_for_get_trips_1",
		},
	}

	c := context.Background()
	mc, err := mongotesting.NewClient(c)
	if err != nil {
		t.Errorf("cannot connect mongodb: %v", err)
	}

	m := NewMongo(mc.Database("coolcar"))

	for _, r := range rows {
		mgutil.NewObjIDWithValue(id.TripID(r.id))
		_, err := m.CreateTrip(c, &rentalpb.Trip{
			AccountId: r.accountID,
		})
		if err != nil {
			t.Fatalf("cannot create rows: %v", err)
		}
	}

	type IDList = []id.TripID
	cases := []struct {
		name      string
		accountID string
		idList    []id.TripID
		wantCount int
	}{
		{
			name:      "get_all",
			accountID: "account_id_for_get_trips",
			idList:    IDList{},
			wantCount: 2,
		},
		{
			name:      "get_5f8132eb10714bf629489051",
			accountID: "account_id_for_get_trips",
			idList:    IDList{id.TripID("5f8132eb10714bf629489051")},
			wantCount: 1,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			res, err := m.GetTrips(
				context.Background(),
				id.AccountID(cc.accountID),
				cc.idList,
			)
			if err != nil {
				t.Errorf("cannot get trips: %v", err)
			}
			if cc.wantCount != len(res) {
				t.Errorf("incorrect result count; want:%d, got: %d", len(cc.idList), len(res))
			}
		})
	}
}

func TestUpdateTrip(t *testing.T) {
	c := context.Background()
	mc, err := mongotesting.NewClient(c)
	if err != nil {
		t.Fatalf("cannot connect mongodb: %v", err)
	}

	m := NewMongo(mc.Database("coolcar"))
	tripID := id.TripID("5f8132eb12714bf629489054")
	accountID := id.AccountID("account_for_update")

	mgutil.NewObjIDWithValue(tripID)
	mgutil.NewUpdatedAtWithValue(10000)

	trip, err := m.CreateTrip(c, &rentalpb.Trip{
		AccountId: accountID.String(),
		Status:    rentalpb.TripStatus_IN_PROGRESS,
		Start: &rentalpb.LocationStatus{
			LocDesc: "start_loc",
		},
	})
	if err != nil {
		t.Fatalf("cannot create trip: %v", err)
	}
	if trip.UpdatedAt != 10000 {
		t.Fatalf("wrong updatedat; want: 10000; got: %d", trip.UpdatedAt)
	}

	update := &rentalpb.Trip{
		AccountId: accountID.String(),
		Status:    rentalpb.TripStatus_IN_PROGRESS,
		Start: &rentalpb.LocationStatus{
			LocDesc: "start_loc_updated",
		},
	}
	cases := []struct {
		name          string
		now           int64
		withUpdatedAt int64
		wantErr       bool
	}{
		{
			name:          "normal_update",
			now:           20000,
			withUpdatedAt: 10000,
		},
		{
			name:          "update_with_stale_timestamp",
			now:           30000,
			withUpdatedAt: 10000,
			wantErr:       true,
		},
		{
			name:          "update_with_refetch",
			now:           40000,
			withUpdatedAt: 20000,
		},
	}

	for _, cc := range cases {
		mgutil.NewUpdatedAtWithValue(cc.now)
		err := m.UpdateTrip(c, tripID, accountID, cc.withUpdatedAt, update)
		if cc.wantErr {
			if err == nil {
				t.Errorf("%s: want error; got none", cc.name)
			} else {
				continue
			}
		} else {
			if err != nil {
				t.Errorf("%s: cannot update: %v", cc.name, err)
			}
		}

		idList := []id.TripID{tripID}
		updatedTrips, err := m.GetTrips(c, accountID, idList)
		if err != nil {
			t.Errorf("%s: cannot get trips after update: %v", cc.name, err)
		}
		if cc.now != updatedTrips[0].UpdatedAt {
			t.Errorf("%s: incorrect updatedat: want %d, got %d",
				cc.name, cc.now, updatedTrips[0].UpdatedAt)
		}
	}
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m))
}
