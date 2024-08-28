package app

import (
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	storageGRPC "radio-storage/internal/grpc"
	storage "radio-storage/internal/service/storage"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	port int,
	allowedIPs []string,
	storageDir string,
	nestingDepth int,
	idLength int,
) *App {
	gRPCServer := grpc.NewServer()

	storageSrv := storage.New(
		log,
		storageDir,
		nestingDepth,
		idLength,
	)

	storageGRPC.Register(
		gRPCServer,
		storageSrv,
		allowedIPs,
	)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

// MustRun runs gRPC server and panics if any error occurs.
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// Run starts gRPC server.
func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grpc server is running", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Stop stopa gRPC server.
func (a *App) Stop() {
	const op = "grpcapp.stop"

	a.log.With(slog.String("op", op)).Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
