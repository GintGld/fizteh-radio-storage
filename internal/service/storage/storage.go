package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"strings"

	grpcModels "radio-storage/internal/domain/grpc"
	"radio-storage/internal/lib/logger/sl"
	"radio-storage/internal/service"
)

const (
	bufferLen = 1024 * 32
)

type Storage struct {
	log          *slog.Logger
	dir          string
	nestingDepth int

	idLength int
	maxId    int
}

func New(
	log *slog.Logger,
	dir string,
	nestingDepth int,
	idLength int,
) *Storage {
	N := 1
	for i := 0; i < idLength; i++ {
		N *= 10
	}

	storage := &Storage{
		log:          log,
		dir:          dir,
		nestingDepth: nestingDepth,
		idLength:     idLength,
		maxId:        N,
	}

	storage.mustInitFilesystem()

	return storage
}

// Upload writes content from io.Reader to generated file,
// returns its id.
func (s *Storage) Upload(ctx context.Context, r *grpcModels.UploadStreamWrapper) (int, error) {
	const op = "Storage.Upload"

	log := s.log.With(
		slog.String("op", op),
	)

	// Generate new id.
	id, err := s.generateNewID()
	if err != nil {
		log.Error("failed to generate new id", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Construct new file.
	dir, err := s.getCorrespondingDir(id)
	if err != nil {
		log.Error("failed to get corresponding dir", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	filename := dir + "/" + strconv.Itoa(id) + ".mp3"

	file, err := os.Create(filename)
	if err != nil {
		log.Error("failed to create new file", slog.String("file", filename), sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer file.Close()

	// Load data
	for {
		chunk, err := r.GetChunk()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return 0, fmt.Errorf("%s: %w", op, err)
		}

		if _, err := file.Write(chunk); err != nil {
			log.Error("failed to write chunk", sl.Err(err))
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	log.Debug("uploaded file", slog.Int("id", id))

	return id, nil
}

// Download writes file to io.Writer.
func (s *Storage) Download(ctx context.Context, id int, w *grpcModels.DownloadStreamWrapper) error {
	const op = "Storage.Download"

	log := s.log.With(
		slog.String("op", op),
		slog.Int("id", id),
	)

	// Check if file exists.
	ok, err := s.checkExistingID(id)
	if err != nil {
		log.Error("failed to check existing id", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	if !ok {
		log.Warn("file not exists")
		return service.ErrFileNotExist
	}

	// Construct path to the file.
	dir, err := s.getCorrespondingDir(id)
	if err != nil {
		log.Error("failed to get corresponding dir", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	filename := dir + "/" + strconv.Itoa(id) + ".mp3"

	log.Debug("download file", slog.Int("id", id))

	// Open file to read.
	file, err := os.Open(filename)
	if err != nil {
		log.Error("failed to open file", slog.String("file", filename), sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	defer file.Close()

	// Copy data.
	buffer := make([]byte, bufferLen)
	for {
		p, err := file.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("%s: %w", op, err)
		}
		fmt.Println(buffer[:p])
		if err := w.Write(buffer[:p]); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	log.Debug("downloaded file", slog.Int("id", id))

	return nil
}

// Delete deletes file by its id.
//
// If file not exists return error.
func (s *Storage) Delete(ctx context.Context, id int) error {
	const op = "Storage.Delete"

	log := s.log.With(
		slog.String("op", op),
		slog.Int("id", id),
	)

	// Check if file exists.
	ok, err := s.checkExistingID(id)
	if err != nil {
		log.Error("failed to check existing id", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	if !ok {
		log.Warn("file not exists")
		return service.ErrFileNotExist
	}

	log.Debug("deleting file")

	// Construct path to the file.
	dir, err := s.getCorrespondingDir(id)
	if err != nil {
		log.Error("failed to get corresponding dir", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	file := dir + "/" + strconv.Itoa(id) + ".mp3"

	// Delete file
	if err := os.Remove(file); err != nil {
		log.Error("failed to delete file", slog.String("file", file), sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("deleted file")

	return nil
}

// mustinitFileSystem inits file system.
// Creates necessary directories.
//
// Panics if occurs error.
func (s *Storage) mustInitFilesystem() {
	const op = "Storage.mustInitFileSystem"

	log := s.log.With(
		slog.String("op", op),
	)

	// indexing directories
	N := 1
	for i := 0; i < s.nestingDepth; i++ {
		N *= 10
	}

	splitted := make([]string, s.nestingDepth)
	for i := 0; i < N; i++ {
		str := strconv.Itoa(i)

		for j := 0; j < s.nestingDepth-len(str); j++ {
			splitted[j] = "0"
		}
		for j := s.nestingDepth - len(str); j < s.nestingDepth; j++ {
			splitted[j] = string(str[j-s.nestingDepth+len(str)])
		}

		dir := s.dir + "/" + strings.Join(splitted, "/")

		if err := os.MkdirAll(dir, 0777); err != nil {
			log.Error("failed to create dir", slog.String("dir", dir), sl.Err(err))
			panic("failed to create dir")
		}
	}
}

func (s *Storage) generateNewID() (int, error) {
	const op = "Storage.generateNewID"

	log := s.log.With(
		slog.String("op", op),
	)

	for {
		sourceID := int(rand.Int31n(int32(s.maxId)))

		exists, err := s.checkExistingID(sourceID)
		if err != nil {
			log.Error("failed to check id", slog.Int("id", sourceID), sl.Err(err))
			return 0, fmt.Errorf("%s: %w", op, err)
		}

		if !exists {
			return sourceID, nil
		}
	}
}

// getCorrespondingDir returns path,
// where source with given id should be placed.
func (s *Storage) getCorrespondingDir(id int) (string, error) {
	const op = "Storage.getCorrespondingDir"

	log := s.log.With(
		slog.String("op", op),
	)

	if id < 0 {
		log.Warn("invalid media source id", slog.Int("id", id))
		return "", fmt.Errorf("%s: invalid media storage id", op)
	}

	str := strconv.Itoa(int(id))

	if len(str) > s.idLength {
		log.Warn("invalid media source id", slog.Int("id", id))
		return "", fmt.Errorf("%s: invalid media storage id", op)
	}

	splitted := make([]string, s.nestingDepth)

	for j := 0; j < s.idLength-len(str); j++ {
		splitted[j] = "0"
	}
	for j := s.idLength - len(str); j < s.nestingDepth; j++ {
		splitted[j] = string(str[j-s.idLength+len(str)])
	}

	return s.dir + "/" + strings.Join(splitted, "/"), nil
}

func (s *Storage) checkExistingID(id int) (bool, error) {
	const op = "Storage.checkExistingID"

	log := s.log.With(
		slog.String("op", op),
	)

	dir, err := s.getCorrespondingDir(id)
	if err != nil {
		log.Error("failed to get dir", slog.Int("id", id), sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	file := dir + "/" + strconv.Itoa(id) + ".mp3"

	if _, err := os.Stat(file); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		log.Error("failed to probe file", slog.String("file", file), sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}
}
