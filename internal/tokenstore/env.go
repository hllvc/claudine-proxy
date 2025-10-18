package tokenstore

import (
	"context"
	"fmt"
	"os"
)

// EnvStore provides read-only access to tokens stored in environment variables.
// Suitable for static token authentication but not OAuth (requires writable storage).
type EnvStore struct {
	envKey string
}

// Compile-time check to ensure EnvStore implements TokenStore
var _ TokenStore = (*EnvStore)(nil)

// NewEnvStore creates an EnvStore for the given environment variable.
// Returns error if the variable name is empty or not set in the environment.
func NewEnvStore(envKey string) (*EnvStore, error) {
	if envKey == "" {
		return nil, fmt.Errorf("environment key cannot be empty")
	}

	if _, exists := os.LookupEnv(envKey); !exists {
		return nil, fmt.Errorf("environment variable %s not set", envKey)
	}

	return &EnvStore{
		envKey: envKey,
	}, nil
}

// Read returns the token from the environment variable. Returns error if empty.
func (e *EnvStore) Read(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	token := os.Getenv(e.envKey)
	if token == "" {
		return "", fmt.Errorf("environment variable %s is empty", e.envKey)
	}
	return token, nil
}

// Write is not supported for environment variables (they are read-only).
func (e *EnvStore) Write(ctx context.Context, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return fmt.Errorf("environment variable storage is read-only")
}
