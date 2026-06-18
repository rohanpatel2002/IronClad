package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
)

// OIDCProvider handles authentication via an external OIDC provider.
type OIDCProvider struct {
	provider     *oidc.Provider
	verifier     *oidc.IDTokenVerifier
	clientID     string
}

// NewOIDCProvider creates a new OIDC provider.
func NewOIDCProvider(ctx context.Context) (*OIDCProvider, error) {
	issuer := os.Getenv("OIDC_ISSUER")
	clientID := os.Getenv("OIDC_CLIENT_ID")
	if issuer == "" || clientID == "" {
		return nil, nil // OIDC disabled
	}

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %v", err)
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &OIDCProvider{
		provider:     provider,
		verifier:     verifier,
		clientID:     clientID,
	}, nil
}

// VerifyToken validates an ID token.
func (p *OIDCProvider) VerifyToken(ctx context.Context, rawToken string) (*oidc.IDToken, error) {
	return p.verifier.Verify(ctx, rawToken)
}
