package main

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/profile"
	profiledao "coolcar/rental/profile/dao"
	"coolcar/rental/trip"
	"coolcar/rental/trip/client/car"
	locdesc "coolcar/rental/trip/client/locDesc"
	profileClient "coolcar/rental/trip/client/profile"
	tripdao "coolcar/rental/trip/dao"
	"coolcar/shared/server"
	"flag"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var addr = flag.String("addr", ":8082", "address to listen")
var mongoURI = flag.String("mongo_uri", "mongodb://localhost:27017", "mongo uri")

// var blobAddr = flag.String("blob_addr", "localhost:8083", "address for blob service")
// var aiAddr = flag.String("ai_addr", "localhost:18001", "address for ai service")
// var carAddr = flag.String("car_addr", "localhost:8084", "address for car service")
var authPublicKeyFile = flag.String("auth_public_key_file", "shared/public.key", "public key file for auth")

func main() {
	flag.Parse()

	logger, err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}

	c := context.Background()
	mc, err := mongo.Connect(c, options.Client().ApplyURI(*mongoURI))
	if err != nil {
		logger.Fatal("cannot connect mongodb", zap.Error(err))
	}
	db := mc.Database("coolcar")

	// create profile server
	profService := &profile.Service{
		Mongo:  profiledao.NewMongo(db),
		Logger: logger,
	}

	// run grpc server
	err = server.RunGRPCServer(&server.GRPCConfig{
		Name:              "coolcar/rental",
		Addr:              *addr,
		AuthPublicKeyFile: *authPublicKeyFile,
		Logger:            logger,
		RegisterFunc: func(s *grpc.Server) {
			// register trip service
			rentalpb.RegisterTripServiceServer(s, &trip.Service{
				Logger:         logger,
				Mongo:          tripdao.NewMongo(db),
				LocDescManager: &locdesc.Manager{},
				CarManager:     &car.Manager{},
				ProfileManager: &profileClient.Manager{
					Fetcher: profService,
				},
			})
			// register profile service
			rentalpb.RegisterProfileServiceServer(s, profService)
		},
	})
	if err != nil {
		logger.Fatal("cannot serve", zap.Error(err))
	}
}
