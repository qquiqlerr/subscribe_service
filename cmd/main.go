package main

import (
	"flag"
	"fmt"
	"github.com/qquiqlerr/subpub"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"subscribe_service/internal/config"
	pb "subscribe_service/internal/gen/pb/proto"
	"subscribe_service/internal/server"
	logger "subscribe_service/pkg/logger"
	"syscall"
	"time"
)

func main() {
	// Parse command line flags
	pathToConfig := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	// Initialize config
	appConfig, err := config.NewConfig(*pathToConfig)
	if err != nil {
		panic("error initialization config: " + err.Error())
	}

	// Initialize logger
	appLogger, err := logger.NewLogger(appConfig.Logger.Level, appConfig.Logger.Format)
	if err != nil {
		panic("error initialization logger: " + err.Error())
	}
	appLogger.Info("Logger initialized", zap.String("level", appLogger.Level().String()), zap.String("format", appConfig.Logger.Format))
	appLogger.Debug("Config", zap.Any("config", appConfig))

	// Initialize PubSub system
	pubSub := subpub.NewSubPub()

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
	pubSubServer := server.NewPubSubServer(appLogger, pubSub)
	pb.RegisterPubSubServer(grpcServer, pubSubServer)

	// Start the gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", appConfig.GRPC.Host, appConfig.GRPC.Port))
	if err != nil {
		appLogger.Fatal("failed to listen", zap.Error(err))
	}

	// Handle termination signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Start the gRPC server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			appLogger.Fatal("failed to serve", zap.Error(err))
		}
	}()

	appLogger.Info("gRPC server started", zap.String("address", fmt.Sprintf("%s:%s", appConfig.GRPC.Host, appConfig.GRPC.Port)))

	// Wait for a termination signal
	<-sigCh
	appLogger.Info("gRPC server shutting down...")
	gracefulShutdown(grpcServer, appLogger, appConfig.GracefulShutdownTimeout)
	appLogger.Info("gRPC server stopped")
}

func gracefulShutdown(grpcServer *grpc.Server, log *zap.Logger, timeout time.Duration) {
	log.Info("Starting graceful shutdown")

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Info("gRPC server stopped gracefully")
	case <-time.After(timeout):
		log.Warn("Graceful shutdown timed out, forcing stop")
		grpcServer.Stop() // принудительно
	}
}
