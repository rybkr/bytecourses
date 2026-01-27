package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FileStorage interface {
	Save(ctx context.Context, filename string, content io.Reader) (storagePath string, err error)
	Delete(ctx context.Context, storagePath string) error
	GetPath(storagePath string) string
	Read(ctx context.Context, storagePath string) (io.ReadCloser, error)
}

type LocalFileStorage struct {
	baseDir string
	urlBase string
}

func NewLocalFileStorage(baseDir, urlBase string) (*LocalFileStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &LocalFileStorage{baseDir: baseDir, urlBase: urlBase}, nil
}

func (s *LocalFileStorage) Save(ctx context.Context, filename string, content io.Reader) (string, error) {
	storagePath := filepath.Join(s.baseDir, filename)

	dir := filepath.Dir(storagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(storagePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, content); err != nil {
		os.Remove(storagePath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filename, nil
}

func (s *LocalFileStorage) Delete(ctx context.Context, storagePath string) error {
	fullPath := filepath.Join(s.baseDir, storagePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *LocalFileStorage) GetPath(storagePath string) string {
	return s.urlBase + "/" + storagePath
}

func (s *LocalFileStorage) Read(ctx context.Context, storagePath string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.baseDir, storagePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}
