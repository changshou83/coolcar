package main

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/car/car"
	"coolcar/car/dao"
	"coolcar/car/mq/amqpclt"
	"coolcar/car/sim"

	// "coolcar/car/ws"
	"coolcar/shared/server"
	"log"

	// "net/http"

	"github.com/namsral/flag"

	// "github.com/gorilla/websocket"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	addr = flag.String("addr", ":8084", "address to listen")
	// wsAddr  = flag.String("ws_addr", ":9090", "websocket address to listen")
	carAddr = flag.String("car_addr", "localhost:8084", "address for car service")
	// tripAddr = flag.String("trip_addr", "locahost:8082", "address for trip service")
	// aiAddr   = flag.String("ai_addr", "localhost:18001", "address for ai service")

	mongoURI = flag.String("mongo_uri", "mongodb://localhost:27017", "mongo uri")
	amqpURL  = flag.String("amqp_url", "amqp://localhost:5672/", "amqp url")
)

func main() {
	flag.Parse()

	// create logger
	logger, err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}

	// connect mongodb
	c := context.Background()
	mc, err := mongo.Connect(c, options.Client().ApplyURI(*mongoURI))
	if err != nil {
		logger.Fatal("cannot connect mongodb", zap.Error(err))
	}
	db := mc.Database("coolcar")

	// connect rabbitmq
	amqpConn, err := amqp.Dial(*amqpURL)
	if err != nil {
		logger.Fatal("cannot dial amqp", zap.Error(err))
	}
	exchange := "coolcar"
	pub, err := amqpclt.NewPublisher(amqpConn, exchange)
	if err != nil {
		logger.Fatal("cannot create publisher", zap.Error(err))
	}

	// run car simulations
	carConn, err := grpc.Dial(*carAddr, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("cannot connect car service", zap.Error(err))
	}
	sub, err := amqpclt.NewSubscriber(amqpConn, exchange, logger)
	if err != nil {
		logger.Fatal("cannot create subscriber", zap.Error(err))
	}
	simController := &sim.Controller{
		Logger:        logger,
		CarService:    carpb.NewCarServiceClient(carConn),
		CarSubscriber: sub,
	}
	go simController.RunSimulations(context.Background())

	// start websocket handler
	// u := &websocket.Upgrader{
	// 	CheckOrigin: func(r *http.Request) bool {
	// 		return true
	// 	},
	// }
	// http.HandleFunc("/ws", ws.CreateHandler(u, sub, logger))
	// go func() {
	// 	addr := *wsAddr
	// 	logger.Info("HTTP server started.", zap.String("addr", addr))
	// 	logger.Sugar().Fatal(http.ListenAndServe(addr, nil))
	// }()

	// run grpc
	err = server.RunGRPCServer(&server.GRPCConfig{
		Name:   "coolcar/car",
		Addr:   *addr,
		Logger: logger,
		RegisterFunc: func(s *grpc.Server) {
			carpb.RegisterCarServiceServer(s, &car.Service{
				Logger:    logger,
				Mongo:     dao.NewMongo(db),
				Publisher: pub,
			})
		},
	})
	if err != nil {
		logger.Fatal("cannot serve", zap.Error(err))
	}
}
