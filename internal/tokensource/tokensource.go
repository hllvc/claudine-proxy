package tokensource

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

// TokenSourceOption configures a TokenSource.
type TokenSourceOption func(*tokenSourceConfig)

// tokenSourceConfig holds configuration for NewTokenSource.
type tokenSourceConfig struct {
	baseTransport http.RoundTripper
}

// WithTransport sets a custom base transport for token refresh requests.
// If not provided, http.DefaultTransport is used.
func WithTransport(transport http.RoundTripper) TokenSourceOption {
	return func(c *tokenSourceConfig) {
		c.baseTransport = transport
	}
}

// TokenSource provides automatic token refresh for Anthropic OAuth2 tokens.
// Wraps oauth2.TokenSource with custom transport for JSON-encoded refresh requests.
type TokenSource struct {
	tokenSource oauth2.TokenSource
}

// Compile-time check to ensure TokenSource implements oauth2.TokenSource
var _ oauth2.TokenSource = (*TokenSource)(nil)

// NewTokenSource creates a TokenSource that automatically refreshes access tokens
// using the provided refresh token.
func NewTokenSource(initialRefreshToken string, endpoint oauth2.Endpoint, opts ...TokenSourceOption) *TokenSource {
	cfg := &tokenSourceConfig{
		baseTransport: http.DefaultTransport,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	oauth2Config := &oauth2.Config{
		ClientID:     ClientID,
		ClientSecret: "", // Empty for PKCE flow (public client)
		Scopes:       scopes,
		Endpoint:     endpoint,
	}

	initialToken := &oauth2.Token{
		RefreshToken: initialRefreshToken,
		// AccessToken populated by first Token() call
	}

	// HTTP client with JSON transport for Anthropic (wraps provided or default transport for connection pooling)
	httpClient := &http.Client{
		Timeout: 30 * time.Second, // Bounds token refresh even during shutdown (oauth2 uses context.Background internally)
		Transport: &tokenRefreshTransport{
			base: cfg.baseTransport,
		},
	}
	// oauth2 package injects custom HTTP clients via context (oauth2.HTTPClient key).
	// Since TokenSource.Token() has no context parameter, we store the context
	// at construction time per oauth2's documented API.
	oauthCtx := context.WithValue(context.Background(), oauth2.HTTPClient, httpClient)
	tokenSource := oauth2Config.TokenSource(oauthCtx, initialToken)

	return &TokenSource{
		tokenSource: tokenSource,
	}
}

// Token returns a valid access token, automatically refreshing if expired.
func (ts *TokenSource) Token() (*oauth2.Token, error) {
	return ts.tokenSource.Token()
}

// tokenRefreshTransport converts oauth2's form-encoded token refresh requests
// to JSON format required by Anthropic's token endpoint.
// The oauth2 package guarantees this transport only receives token endpoint requests.
type tokenRefreshTransport struct {
	base http.RoundTripper
}

// Compile-time check that tokenRefreshTransport implements http.RoundTripper.
var _ http.RoundTripper = (*tokenRefreshTransport)(nil)

// RoundTrip intercepts token refresh requests and converts them from form-encoded to JSON.
func (t *tokenRefreshTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Defer close since we consume the body entirely and create a new body for the cloned request.
	// Unlike passthrough patterns, we don't forward the original body to the next RoundTripper.
	defer func() { _ = req.Body.Close() }()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("reading request body: %w", err)
	}

	formData, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("parsing form data: %w", err)
	}

	// Convert all form data to JSON format
	jsonData := make(map[string]string, len(formData))
	for key, values := range formData {
		jsonData[key] = values[0] // OAuth2 spec defines single-value parameters
	}

	jsonBody, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("marshaling JSON request: %w", err)
	}

	newReq := req.Clone(req.Context())
	newReq.Body = io.NopCloser(bytes.NewReader(jsonBody))
	newReq.ContentLength = int64(len(jsonBody))
	newReq.Header.Set("Content-Type", "application/json")

	return t.base.RoundTrip(newReq)
}
