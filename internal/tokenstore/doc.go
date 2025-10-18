// Package tokenstore provides persistent storage abstractions for authentication tokens.
//
// Supports storage backends with different security and deployment tradeoffs:
//   - File: Local filesystem storage with atomic writes and secure permissions
//   - Env: Read-only environment variable access (requires external secret management)
//
// OAuth authentication requires writable storage (file), while static
// token authentication can use any backend including read-only env storage.
package tokenstore
