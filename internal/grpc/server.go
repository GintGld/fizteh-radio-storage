package server

import (
	"context"
	"errors"
	"io"
	"slices"

	ssov1 "github.com/GintGld/fizteh-radio-proto/gen/go/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	grpcModels "github.com/GintGld/fizteh-radio-storage/internal/domain/grpc"
	"github.com/GintGld/fizteh-radio-storage/internal/service"
)

type Storage interface {
	Upload(ctx context.Context, w io.Reader) (int, error)
	Download(ctx context.Context, id int, w io.Writer) error
	Delete(ctx context.Context, fileId int) error
}

type serverAPI struct {
	ssov1.UnimplementedFileServiceServer

	storage    Storage
	allowedIps []string
}

func Register(
	gRPC *grpc.Server,
	storage Storage,
	allowedIps []string,
) {
	ssov1.RegisterFileServiceServer(gRPC, &serverAPI{
		storage:    storage,
		allowedIps: allowedIps,
	})
}

func (s *serverAPI) Upload(
	stream grpc.ClientStreamingServer[ssov1.UploadRequest, ssov1.UploadResponse],
) error {
	ctx := stream.Context()
	p, _ := peer.FromContext(ctx)
	if !slices.Contains(s.allowedIps, p.Addr.String()) {
		return status.Error(codes.PermissionDenied, "ip is not allowed")
	}

	uploadStream := &grpcModels.UploadStreamWrapper{Stream: stream}

	id, err := s.storage.Upload(ctx, uploadStream)
	if err != nil {
		return status.Error(codes.Internal, "internal server error")
	}

	if err := stream.SendAndClose(&ssov1.UploadResponse{FileId: int32(id)}); err != nil {
		return status.Error(codes.Internal, "internal server error")
	}

	return nil
}

func (s *serverAPI) Download(
	req *ssov1.DownloadRequest,
	stream grpc.ServerStreamingServer[ssov1.DownloadResponse],
) error {
	ctx := stream.Context()
	p, _ := peer.FromContext(ctx)
	if !slices.Contains(s.allowedIps, p.Addr.String()) {
		return status.Error(codes.PermissionDenied, "ip is not allowed")
	}

	downloadStream := &grpcModels.DownloadStreamWrapper{Stream: stream}

	if err := s.storage.Download(ctx, int(req.GetFileId()), downloadStream); err != nil {
		if errors.Is(err, service.ErrFileNotExist) {
			return status.Error(codes.NotFound, "file not exists")
		}
		return status.Error(codes.Internal, "internal server error")
	}

	return nil
}

func (s *serverAPI) Delete(
	ctx context.Context,
	req *ssov1.DeleteRequest,
) (*ssov1.DeleteResponse, error) {
	p, _ := peer.FromContext(ctx)
	if !slices.Contains(s.allowedIps, p.Addr.String()) {
		return nil, status.Error(codes.PermissionDenied, "ip is not allowed")
	}

	if err := s.storage.Delete(ctx, int(req.GetFileId())); err != nil {
		if errors.Is(err, service.ErrFileNotExist) {
			return nil, status.Error(codes.NotFound, "file not exists")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &ssov1.DeleteResponse{Success: true}, nil
}
