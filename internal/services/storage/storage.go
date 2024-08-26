package storage

import (
	"context"
	"io"
	"log/slog"
)

type Storage struct {
	log *slog.Logger
}

func New(
	log *slog.Logger,
) *Storage {
	return &Storage{
		log: log,
	}
}

func (s *Storage) Upload(ctx context.Context, w io.Reader) (int, error) {
	panic("implement me")
}

func (s *Storage) Download(ctx context.Context, id int32, w io.Writer) error {
	panic("implement me")
}

func (s *Storage) Delete(ctx context.Context, fileId int32) error {
	panic("implement me")
}
