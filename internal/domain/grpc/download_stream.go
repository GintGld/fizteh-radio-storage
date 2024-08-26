package grpc

import (
	"fmt"

	ssov1 "github.com/GintGld/fizteh-radio-proto/gen/go/storage"
	"google.golang.org/grpc"
)

// DownloadStreamWrapper is wrapper for io.Writer interface
type DownloadStreamWrapper struct {
	Stream grpc.ServerStreamingServer[ssov1.DownloadResponse]
}

func (w *DownloadStreamWrapper) Write(p []byte) (int, error) {
	const op = "DownloadStreamWrapper.Write"

	if err := w.Stream.Send(&ssov1.DownloadResponse{Chunk: p}); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return len(p), nil
}
