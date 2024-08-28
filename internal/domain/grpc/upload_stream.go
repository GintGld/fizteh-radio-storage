package grpc

import (
	"errors"
	"fmt"
	"io"

	ssov1 "github.com/GintGld/fizteh-radio-proto/gen/go/storage"
	"google.golang.org/grpc"
)

type UploadStreamWrapper struct {
	Stream grpc.ClientStreamingServer[ssov1.UploadRequest, ssov1.UploadResponse]
}

func (w *UploadStreamWrapper) GetChunk() ([]byte, error) {
	const op = "UploadStreamWrapper.GetChunk"

	req, err := w.Stream.Recv()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return req.GetChunk(), nil
}
