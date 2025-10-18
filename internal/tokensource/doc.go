// Package tokensource provides OAuth2 token acquisition and automatic refresh
// for Anthropic Claude API.
//
// Anthropic's OAuth2 implementation deviates from the standard in several critical ways
// that require custom handling:
//   - Token exchange and refresh use JSON-encoded requests (standard OAuth2 uses form-encoding)
//
// # Token Sources
//
// Use NewTokenSource for OAuth2 refresh tokens:
//
//	ts := tokensource.NewTokenSource(refreshToken, tokensource.Endpoint)
//	// TokenSource implements oauth2.TokenSource and can be used with oauth2.Transport
//
// # Custom Base Transport
//
// Configure a custom base transport for token refresh requests (e.g., for proxies or custom timeouts):
//
//	ts := tokensource.NewTokenSource(
//		refreshToken,
//		tokensource.Endpoint,
//		tokensource.WithTransport(customTransport),
//	)
package tokensource
