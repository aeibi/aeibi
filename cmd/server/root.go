package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"aeibi/api"
	"aeibi/cmd/env"
	"aeibi/internal/config"
	"aeibi/internal/controller"
	"aeibi/internal/service"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Run boots the application with the provided configuration.
func Run(ctx context.Context, cfg *config.Config) error {
	dbConn, err := env.InitDB(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	ossClient, err := env.InitOSS(ctx, cfg.OSS)
	if err != nil {
		return err
	}

	// Initialize service registrars

	gatewayEndpoint := cfg.Server.GRPCAddr
	gatewayDialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// User service
	userSvc := service.NewUserService(dbConn, ossClient, cfg)
	userHandler := controller.NewUserHandler(userSvc)
	userRegistrar := ServiceRegistrar{
		Name: "user",
		RegisterGRPC: func(s *grpc.Server) {
			api.RegisterUserServiceServer(s, userHandler)
		},
		RegisterGateway: func(ctx context.Context, mux *runtime.ServeMux) error {
			return api.RegisterUserServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts)
		},
	}

	// Post service
	postSvc := service.NewPostService(dbConn, ossClient)
	postHandler := controller.NewPostHandler(postSvc)
	postRegistrar := ServiceRegistrar{
		Name: "post",
		RegisterGRPC: func(s *grpc.Server) {
			api.RegisterPostServiceServer(s, postHandler)
		},
		RegisterGateway: func(ctx context.Context, mux *runtime.ServeMux) error {
			return api.RegisterPostServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts)
		},
	}

	// File service
	fileSvc := service.NewFileService(dbConn, ossClient)
	fileHandler := controller.NewFileHandler(fileSvc)
	fileRegistrar := ServiceRegistrar{
		Name: "file",
		RegisterGRPC: func(s *grpc.Server) {
			api.RegisterFileServiceServer(s, fileHandler)
		},
		RegisterGateway: func(ctx context.Context, mux *runtime.ServeMux) error {
			return api.RegisterFileServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts)
		},
	}

	registrars := []ServiceRegistrar{
		userRegistrar,
		postRegistrar,
		fileRegistrar,
	}

	// Start gRPC server
	grpcServer, grpcErrCh, err := StartGRPCServer(cfg, registrars)
	if err != nil {
		return err
	}

	// Start gRPC-Gateway HTTP server
	httpServer, httpErrCh, err := StartGateway(ctx, cfg, registrars)
	if err != nil {
		grpcServer.GracefulStop()
		return err
	}

	slog.Info("gRPC server listening", "addr", cfg.Server.GRPCAddr)
	slog.Info("HTTP gateway listening", "addr", cfg.Server.HTTPAddr)

	// Wait for termination
	select {
	case err := <-grpcErrCh:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Warn("HTTP shutdown after gRPC failure", "error", err)
		}
		return fmt.Errorf("gRPC server: %w", err)
	case err := <-httpErrCh:
		grpcServer.GracefulStop()
		return fmt.Errorf("HTTP server: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcServer.GracefulStop()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Warn("HTTP shutdown", "error", err)
		}
	}

	return nil
}
