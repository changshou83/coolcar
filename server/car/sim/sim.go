package sim

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/car/mq"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Controller struct {
	CarService    carpb.CarServiceClient
	CarSubscriber mq.Subscriber
	Logger        *zap.Logger
}

func (c *Controller) RunSimulations(ctx context.Context) {
	var cars []*carpb.CarEntity
	for {
		time.Sleep(3 * time.Second)
		res, err := c.CarService.GetCars(ctx, &carpb.GetCarsRequest{})
		if err != nil {
			c.Logger.Error("cannot get cars", zap.Error(err))
			continue
		}
		cars = res.Cars
		break
	}

	c.Logger.Info("Running car simulations.", zap.Int("car_count", len(cars)))
	// subscribe coolcar queue
	carCh, carCleanUp, err := c.CarSubscriber.Subscribe(ctx)
	defer carCleanUp()
	if err != nil {
		c.Logger.Error("cannot subscribe to car", zap.Error(err))
		return
	}
	// simulate cars
	carChans := make(map[string]chan *carpb.Car)
	for _, car := range cars {
		carFanoutCh := make(chan *carpb.Car)
		carChans[car.Id] = carFanoutCh
		go c.SimulateCar(context.Background(), car, carFanoutCh)
	}

	// receive message from go channel
	for carUpdate := range carCh {
		ch := carChans[carUpdate.Id]
		if ch != nil {
			ch <- carUpdate.Car
		}
	}
}

// SimulateCar simulates a real car.
func (c *Controller) SimulateCar(ctx context.Context, initial *carpb.CarEntity, carCh chan *carpb.Car) {
	car := initial
	c.Logger.Info("Simulating car", zap.String("id", car.Id))
	// exec actions based on car status
	for update := range carCh {
		if update.Status == carpb.CarStatus_LOCKING {
			updated, err := c.lockCar(ctx, car)
			if err != nil {
				c.Logger.Error("cannot unlock car.", zap.String("id", car.Id), zap.Error(err))
				break
			}
			car = updated
		} else if update.Status == carpb.CarStatus_UNLOCKING {
			updated, err := c.unlockCar(ctx, car)
			if err != nil {
				c.Logger.Error("cannot lock car.", zap.String("id", car.Id), zap.Error(err))
				break
			}
			car = updated
		}
	}
}

func (c *Controller) lockCar(ctx context.Context, car *carpb.CarEntity) (*carpb.CarEntity, error) {
	// update real car's status
	car.Car.Status = carpb.CarStatus_LOCKED
	// update car state in mongodb
	_, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
		Id:     car.Id,
		Status: carpb.CarStatus_LOCKED,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot update car state: %v", err)
	}

	return car, nil
}

func (c *Controller) unlockCar(ctx context.Context, car *carpb.CarEntity) (*carpb.CarEntity, error) {
	car.Car.Status = carpb.CarStatus_UNLOCKED
	_, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
		Id:     car.Id,
		Status: carpb.CarStatus_UNLOCKED,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot update car state: %v", err)
	}

	return car, nil
}
