package grpc

import (
	"errors"
	"fmt"
	"io"

	ssov1 "github.com/GintGld/fizteh-radio-proto/gen/go/storage"
	"google.golang.org/grpc"
)

// UploadStreamWrapper is wrapper for io.Reader interface
type UploadStreamWrapper struct {
	Stream grpc.ClientStreamingServer[ssov1.UploadRequest, ssov1.UploadResponse]
}

func (w *UploadStreamWrapper) Read(p []byte) (int, error) {
	const op = "UploadStreamWrapper.GetChunk"

	req, err := w.Stream.Recv()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return 0, io.EOF
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	p = req.Chunk

	return len(p), nil
}
