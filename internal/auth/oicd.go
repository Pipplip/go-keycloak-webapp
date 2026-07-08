package auth

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Provider struct {
	OAuth2Config oauth2.Config
	Verifier     *oidc.IDTokenVerifier
}

func NewProvider(
	ctx context.Context,
	issuer string,
	clientID string,
	clientSecret string,
	redirectURL string,
) (*Provider, error) {

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes: []string{
			oidc.ScopeOpenID,
			"profile",
			"email",
		},
	}

	verifier := provider.Verifier(
		&oidc.Config{
			ClientID: clientID,
		},
	)

	return &Provider{
		OAuth2Config: config,
		Verifier:     verifier,
	}, nil
}
