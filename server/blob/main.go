package main

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	"coolcar/blob/blob"
	"coolcar/blob/cos"
	"coolcar/blob/dao"
	"coolcar/shared/server"
	"flag"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var addr = flag.String("addr", ":8083", "address to listen")
var mongoURI = flag.String("mongo_uri", "mongodb://localhost:27017", "mongo uri")

// var cosAddr = flag.String("cos_addr", "<URL>", "cos address")
// var cosSecID = flag.String("cos_sec_id", "<SEC_ID>", "cos secret id")
// var cosSecKey = flag.String("cos_sec_key", "<SEC_KEY>", "cos secret key")
var cosAddr = flag.String("cos_addr", "https://coolcar-1304312343.cos.ap-beijing.myqcloud.com", "cos address")
var cosSecID = flag.String("cos_sec_id", "AKIDoRp1zjGblgMo0tY0GI5csufKq8Wt41oR", "cos secret id")
var cosSecKey = flag.String("cos_sec_key", "Xq33qNfjaFpQr53qX6QVKsAByWoqlQr5", "cos secret key")

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

	storage, err := cos.NewService(*cosAddr, *cosSecID, *cosSecKey)
	if err != nil {
		logger.Fatal("cannot create cos service", zap.Error(err))
	}

	err = server.RunGRPCServer(&server.GRPCConfig{
		Name:   "coolcar/blob",
		Addr:   *addr,
		Logger: logger,
		RegisterFunc: func(s *grpc.Server) {
			blobpb.RegisterBlobServiceServer(s, &blob.Service{
				Storage: storage,
				Mongo:   dao.NewMongo(db),
				Logger:  logger,
			})
		},
	})
	if err != nil {
		logger.Fatal("cannot serve", zap.Error(err))
	}
}
