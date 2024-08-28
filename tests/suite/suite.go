package suite

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"

	"radio-storage/internal/config"

	ssov1 "github.com/GintGld/fizteh-radio-proto/gen/go/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	Cfg    *config.Config
	Client ssov1.FileServiceClient
}

const (
	grpcHost = "localhost"
)

// New creates new test suite.
//
// TODO: for pipeline tests we need to wait for app is ready
func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadPath(configPath())

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	cc, err := grpc.NewClient(
		grpcAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	return ctx, &Suite{
		T:      t,
		Cfg:    cfg,
		Client: ssov1.NewFileServiceClient(cc),
	}
}

// getCorrespondingDir returns path,
// where source with given id should be placed.
func (s *Suite) GetCorrespondingDir(id int) (string, error) {
	const op = "Storage.getCorrespondingDir"

	if id < 0 {
		return "", fmt.Errorf("%s: invalid media storage id", op)
	}

	str := strconv.Itoa(int(id))

	if len(str) > s.Cfg.Source.IdLength {
		return "", fmt.Errorf("%s: invalid media storage id", op)
	}

	splitted := make([]string, s.Cfg.Source.NestingDepth)

	for j := 0; j < s.Cfg.Source.IdLength-len(str); j++ {
		splitted[j] = "0"
	}
	for j := s.Cfg.Source.IdLength - len(str); j < s.Cfg.Source.NestingDepth; j++ {
		splitted[j] = string(str[j-s.Cfg.Source.IdLength+len(str)])
	}

	return s.Cfg.Source.SourcePath + "/" + strings.Join(splitted, "/"), nil
}

func configPath() string {
	const key = "CONFIG_PATH"

	if v := os.Getenv(key); v != "" {
		return v
	}

	return "../config/local-tests.yaml"
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}
