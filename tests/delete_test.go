package tests

import (
	"math/rand"
	"radio-storage/tests/suite"
	"testing"

	storagev1 "github.com/GintGld/fizteh-radio-proto/gen/go/storage"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	ctx, st := suite.New(t)

	// Generate data.
	var size int64 = rand.Int63n(maxDataSize)
	// data := make([]byte, 0, size)

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
	}

	// Extract file id.
	resp, err := stream.CloseAndRecv()
	require.NoError(t, err)
	id := int(resp.GetFileId())

	res, err := st.Client.Delete(ctx, &storagev1.DeleteRequest{FileId: int32(id)})
	require.NoError(t, err)
	require.True(t, res.GetSuccess())

	streamDownload, err := st.Client.Download(ctx, &storagev1.DownloadRequest{FileId: int32(id)})
	require.NoError(t, err)

	_, err = streamDownload.Recv()
	require.EqualError(t, err, "rpc error: code = NotFound desc = file not exists")
}
