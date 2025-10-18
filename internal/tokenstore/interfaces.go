package tokenstore

import "context"

// TokenStore reads and writes tokens to persistent storage.
//
// OAuth authentication requires writable storage.
type TokenStore interface {
	// Read returns the stored token. Returns error if token is missing or empty.
	Read(ctx context.Context) (string, error)

	// Write persists the token to storage. Returns error if storage backend
	// is read-only (e.g., environment variables) or if write operation fails.
	Write(ctx context.Context, token string) error
}
