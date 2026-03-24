package local

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage implements file storage using local filesystem (Phase 1)
type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	_ = os.MkdirAll(basePath, 0755)
	return &LocalStorage{basePath: basePath}
}

func (s *LocalStorage) Upload(_ context.Context, key string, data io.Reader, _ string) (string, error) {
	fullPath := filepath.Join(s.basePath, key)
	_ = os.MkdirAll(filepath.Dir(fullPath), 0755)

	f, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, data)
	return key, err
}

func (s *LocalStorage) Delete(_ context.Context, key string) error {
	return os.Remove(filepath.Join(s.basePath, key))
}
