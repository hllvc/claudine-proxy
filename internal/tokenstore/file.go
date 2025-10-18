package tokenstore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileStore provides atomic file-based token storage with secure permissions.
// Writes use temp file + rename for crash safety.
type FileStore struct {
	filePath string
}

// Compile-time check to ensure FileStore implements TokenStore
var _ TokenStore = (*FileStore)(nil)

// NewFileStore creates a FileStore for the given path, creating parent directories
// with 0700 permissions if they don't exist.
func NewFileStore(filePath string) (*FileStore, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	return &FileStore{
		filePath: filePath,
	}, nil
}

// Read returns the stored token after trimming whitespace. Returns error if file
// doesn't exist, is empty, or has insecure permissions.
func (f *FileStore) Read(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Check file permissions before reading
	info, err := os.Stat(f.filePath)
	if err != nil {
		return "", err
	}
	if info.Mode().Perm() != 0600 {
		return "", fmt.Errorf("insecure permissions on %s: %04o (expected 0600)", f.filePath, info.Mode().Perm())
	}

	data, err := os.ReadFile(f.filePath)
	if err != nil {
		return "", err
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("empty token file %s", f.filePath)
	}
	return token, nil
}

// Write atomically saves the token using temp file + rename for crash safety.
// Sets file permissions to 0600 (owner read/write only).
func (f *FileStore) Write(ctx context.Context, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Create secure temp file in same directory for atomic rename
	dir := filepath.Dir(f.filePath)
	tempFile, err := os.CreateTemp(dir, "*.tmp")
	if err != nil {
		return err
	}
	tempName := tempFile.Name()
	// Cleanup deferred for all exit paths
	defer func() { _ = os.Remove(tempName) }()
	defer func() { _ = tempFile.Close() }()

	if _, err := tempFile.Write([]byte(strings.TrimSpace(token + "\n"))); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	// Atomic rename to final location
	if err := os.Rename(tempName, f.filePath); err != nil {
		return err
	}

	// Set secure file permissions (0600 = rw-------)
	if err := os.Chmod(f.filePath, 0600); err != nil {
		return err
	}

	return nil
}
