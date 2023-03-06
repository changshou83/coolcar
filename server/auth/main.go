package main

import (
	"context"
	authpb "coolcar/auth/api/gen/v1"
	"coolcar/auth/auth"
	"coolcar/auth/dao"
	"coolcar/auth/token"
	"coolcar/auth/wechat"
	"coolcar/shared/server"
	"flag"
	"io"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	addr            = flag.String("addr", ":8081", "address to listen")
	mongoURI        = flag.String("mongo_uri", "mongodb://localhost:27017", "mongo uri")
	privateKeyFile  = flag.String("private_key_file", "auth/private.key", "rsa private key file")
	wechatAppID     = flag.String("wechat_app_id", "wx779670c13fa0c873", "wechat app id")
	wechatAppSecret = flag.String("wechat_app_secret", "eec6b69ae027cb45c64a15598501ad19", "wechat app secret")
)

func main() {
	flag.Parse()

	logger, err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}

	c := context.Background()
	mc, err := mongo.Connect(c, options.Client().ApplyURI(*mongoURI))
	if err != nil {
		logger.Fatal("cannot connect mongodb: %v", zap.Error(err))
	}
	// 从private.key读取private key
	pkFile, err := os.Open(*privateKeyFile)
	if err != nil {
		logger.Fatal("cannot open private key", zap.Error(err))
	}
	pkBytes, err := io.ReadAll(pkFile)
	if err != nil {
		logger.Fatal("cannot read private key", zap.Error(err))
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pkBytes)
	if err != nil {
		logger.Fatal("cannot parse private key", zap.Error(err))
	}

	service := &auth.Service{
		Logger:         logger,
		Mongo:          dao.NewMongo(mc.Database("coolcar")),
		TokenGenerator: token.NewJWTTokenGen("coolcar/auth", privateKey),
		TokenExpire:    2 * time.Hour,
		OpenIDResolver: &wechat.Service{
			AppID:     *wechatAppID,
			AppSecret: *wechatAppSecret,
		},
	}

	err = server.RunGRPCServer(&server.GRPCConfig{
		Name:   "coolcar/auth",
		Addr:   *addr,
		Logger: logger,
		RegisterFunc: func(s *grpc.Server) {
			authpb.RegisterAuthServiceServer(s, service)
		},
	})
	logger.Fatal("cannot serve", zap.Error(err))
}
