package grpc

import (
	"fmt"

	ssov1 "github.com/GintGld/fizteh-radio-proto/gen/go/storage"
	"google.golang.org/grpc"
)

type DownloadStreamWrapper struct {
	Stream grpc.ServerStreamingServer[ssov1.DownloadResponse]
}

func (w *DownloadStreamWrapper) Write(p []byte) error {
	const op = "DownloadStreamWrapper.Write"

	if err := w.Stream.Send(&ssov1.DownloadResponse{Chunk: p}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
