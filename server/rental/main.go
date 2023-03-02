package main

import (
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/trip"
	"coolcar/shared/server"
	"log"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger, err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}

	err = server.RunGRPCServer(&server.GRPCConfig{
		Name:              "coolcar/rental",
		Addr:              ":8082",
		AuthPublicKeyFile: "shared/public.key",
		Logger:            logger,
		RegisterFunc: func(s *grpc.Server) {
			rentalpb.RegisterTripServiceServer(s, &trip.Service{
				Logger: logger,
			})
		},
	})
	if err != nil {
		logger.Fatal("cannot serve", zap.Error(err))
	}
}
