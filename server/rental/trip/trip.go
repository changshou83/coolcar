package trip

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/trip/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	"coolcar/shared/mongo/objid"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	ProfileManager ProfileManager
	CarManager     CarManager
	LocDescManager LocDescManager
	DistanceCalc   DistanceCalc
	Mongo          *dao.Mongo
	Logger         *zap.Logger
	rentalpb.UnimplementedTripServiceServer
}

// ProfileManager defines the ACL (Anti Corruption Layer)
// for profile verification logic.
type ProfileManager interface {
	Verify(c context.Context, aid id.AccountID) (id.IdentityID, error)
}

// CarManager defines the ACL for car management.
type CarManager interface {
	Verify(c context.Context, cid id.CarID, loc *rentalpb.Location) error
	Unlock(c context.Context, cid id.CarID, aid id.AccountID, tid id.TripID, avatarURL string) error
	Lock(c context.Context, cid id.CarID) error
}

// LocDescManager resolves loc_desc
type LocDescManager interface {
	Resolve(c context.Context, loc *rentalpb.Location) (string, error)
}

// DistanceCalc calculates distance between given location
type DistanceCalc interface {
	DistanceKm(context.Context, *rentalpb.Location, *rentalpb.Location) (float64, error)
}

func (s *Service) CreateTrip(
	ctx context.Context,
	req *rentalpb.CreateTripRequest,
) (*rentalpb.TripEntity, error) {
	accountID, err := auth.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if req.CarId == "" || req.Start == nil {
		return nil, status.Error(codes.InvalidArgument, "")
	}

	// 验证驾驶者身份
	identityID, err := s.ProfileManager.Verify(ctx, accountID)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	// 检查车辆状态
	carID := id.CarID(req.CarId)
	err = s.CarManager.Verify(ctx, carID, req.Start)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	// 创建行程：写入数据库，开始计费
	locDesc, err := s.LocDescManager.Resolve(ctx, req.Start)
	if err != nil {
		s.Logger.Info("cannot resolve loc_desc", zap.Stringer("loc", req.Start))
	}
	ls := s.calcCurrentStatus(ctx, &rentalpb.LocationStatus{
		Location: req.Start,
		LocDesc:  locDesc,
	}, req.Start)
	trip, err := s.Mongo.CreateTrip(ctx, &rentalpb.Trip{
		AccountId:  accountID.String(),
		CarId:      carID.String(),
		IdentityId: identityID.String(),
		Status:     rentalpb.TripStatus_IN_PROGRESS,
		Start:      ls,
		Current:    ls,
	})
	if err != nil {
		s.Logger.Warn("cannot create trip", zap.Error(err))
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}

	// 车辆开锁
	go func() {
		err := s.CarManager.Unlock(context.Background(), carID, accountID, objid.ToTripID(trip.ID), req.AvatarUrl)
		if err != nil {
			s.Logger.Error("cannot unlock car", zap.Error(err))
		}
	}()

	return &rentalpb.TripEntity{
		Id:   trip.ID.Hex(),
		Trip: trip.Trip,
	}, nil
}

func (s *Service) GetTrips(
	ctx context.Context,
	req *rentalpb.GetTripsRequest,
) (*rentalpb.GetTripsResponse, error) {
	aid, err := auth.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var idList []id.TripID
	for _, str := range req.IdList {
		tripID := id.TripID(str)
		idList = append(idList, tripID)
	}
	trips, err := s.Mongo.GetTrips(ctx, aid, idList)
	if err != nil {
		s.Logger.Error("cannot get trips", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	res := &rentalpb.GetTripsResponse{}
	for _, trip := range trips {
		res.Trips = append(res.Trips, &rentalpb.TripEntity{
			Id:   trip.ID.Hex(),
			Trip: trip.Trip,
		})
	}
	return res, nil
}

func (s *Service) UpdateTrip(
	ctx context.Context,
	req *rentalpb.UpdateTripRequest,
) (*rentalpb.Trip, error) {
	// get account id
	aid, err := auth.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// get trip
	idList := []id.TripID{id.TripID(req.Id)}
	trips, err := s.Mongo.GetTrips(ctx, aid, idList)
	if err != nil {
		return nil, status.Error(codes.NotFound, "")
	}
	trip := trips[0]
	// verify
	if trip.Trip.Status == rentalpb.TripStatus_FINISHED {
		return nil, status.Error(codes.FailedPrecondition, "cannot update a finished trip")
	}
	if trip.Trip.Current == nil {
		s.Logger.Error("trip without current set", zap.String("id", idList[0].String()))
	}
	// update
	cur := trip.Trip.Current.Location
	if req.Current != nil {
		cur = req.Current
	}
	trip.Trip.Current = s.calcCurrentStatus(ctx, trip.Trip.Current, cur)

	if req.EndTrip {
		trip.Trip.End = trip.Trip.Current
		trip.Trip.Status = rentalpb.TripStatus_FINISHED

		err := s.CarManager.Lock(ctx, id.CarID(trip.Trip.CarId))
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "cannot lock car: %v", err)
		}
	}

	err = s.Mongo.UpdateTrip(ctx, idList[0], aid, trip.UpdatedAt, trip.Trip)
	if err != nil {
		return nil, status.Error(codes.Aborted, "")
	}
	return trip.Trip, nil
}

var nowFunc = func() int64 {
	return time.Now().Unix()
}

const centsPerSec = 0.7
const kmPerSec = 0.02

func (s *Service) calcCurrentStatus(
	c context.Context,
	prev *rentalpb.LocationStatus,
	cur *rentalpb.Location,
) *rentalpb.LocationStatus {
	now := nowFunc()
	elapsedSec := float64(now - prev.TimestampSec)

	// dist, err := s.DistanceCalc.DistanceKm(c, prev.Location, cur)
	// if err != nil {
	// 	s.Logger.Warn("cannot calculate distance", zap.Error(err))
	// }

	locDesc, err := s.LocDescManager.Resolve(c, cur)
	if err != nil {
		s.Logger.Info("cannot resolve loc_desc", zap.Error(err))
	}

	return &rentalpb.LocationStatus{
		Location: cur,
		FeeCent:  prev.FeeCent + int32(centsPerSec*elapsedSec),
		// KmDriven:     prev.KmDriven + dist,
		KmDriven:     prev.KmDriven + kmPerSec*elapsedSec,
		TimestampSec: now,
		LocDesc:      locDesc,
	}
}
