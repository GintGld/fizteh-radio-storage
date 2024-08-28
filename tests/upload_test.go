package tests

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"radio-storage/tests/suite"

	storagev1 "github.com/GintGld/fizteh-radio-proto/gen/go/storage"
	"github.com/stretchr/testify/require"
)

func TestUpload(t *testing.T) {
	ctx, st := suite.New(t)

	// Generate data.
	var size int64 = rand.Int63n(maxDataSize)
	data := make([]byte, 0, size)

	var bufferSize int64 = rand.Int63n(maxBufferLen)

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

	// Get filename of file.
	dir, err := st.GetCorrespondingDir(id)
	require.NoError(t, err)
	filename := fmt.Sprintf("../%s/%d.mp3", dir, id)

	// Extract actual data.
	dataActual, err := os.ReadFile(filename)
	require.NoError(t, err)

	// Check if data equal.
	require.Equal(t, data, dataActual)
}
