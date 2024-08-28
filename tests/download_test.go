package tests

import (
	"errors"
	"io"
	"math/rand"
	"radio-storage/tests/suite"
	"testing"

	storagev1 "github.com/GintGld/fizteh-radio-proto/gen/go/storage"
	"github.com/stretchr/testify/require"
)

func TestDownload(t *testing.T) {
	ctx, st := suite.New(t)

	// Generate data.
	var size int64 = 10 //rand.Int63n(maxDataSize)
	data := make([]byte, 0, size)

	var bufferSize int64 = 5 //rand.Int63n(maxBufferLen)

	stream, err := st.Client.Upload(ctx)
	require.NoError(t, err)

	var buff []byte

	// Send data.
	var i int64
	for i = 0; i < size; i += bufferSize {
		l := min(i+bufferSize, size)
		buff = make([]byte, l-i)
		for k := range buff {
			buff[k] = byte(rand.Uint32())
		}

		err := stream.Send(&storagev1.UploadRequest{Chunk: buff})
		require.NoError(t, err)

		data = append(data, buff...)
	}

	// Extract file id.
	resp, err := stream.CloseAndRecv()
	require.NoError(t, err)
	id := int(resp.GetFileId())

	// Open stream for reading.
	streamDownload, err := st.Client.Download(ctx, &storagev1.DownloadRequest{FileId: int32(id)})
	require.NoError(t, err)

	// Read data.
	dataActual := make([]byte, 0, len(data))
	for {
		recv, err := streamDownload.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			require.NoError(t, err)
		}

		dataActual = append(dataActual, recv.GetChunk()...)
	}

	require.Equal(t, data, dataActual)
}
