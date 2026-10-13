package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Provider struct {
	OAuth2Config oauth2.Config
	Verifier     *oidc.IDTokenVerifier
	Issuer       string
}

func NewProvider(
	ctx context.Context,
	discoveryURL string,
	expectedIssuer string,
	clientID string,
	clientSecret string,
	redirectURL string,
	authorizationURL string,
) (*Provider, error) {

	provider, err := oidc.NewProvider(ctx, discoveryURL)
	if err != nil {
		return nil, err
	}

	endpoint := provider.Endpoint()
	if authorizationURL != "" {
		endpoint.AuthURL = authorizationURL
	}

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     endpoint,
		RedirectURL:  redirectURL,
		Scopes: []string{
			oidc.ScopeOpenID,
			"profile",
			"email",
		},
	}

	verifierConfig := &oidc.Config{ClientID: clientID}
	if discoveryURL != expectedIssuer {
		verifierConfig.SkipIssuerCheck = true
	}
	verifier := provider.Verifier(verifierConfig)

	return &Provider{
		OAuth2Config: config,
		Verifier:     verifier,
		Issuer:       expectedIssuer,
	}, nil
}

func (p *Provider) ValidateIssuer(issuer string) error {
	if issuer != p.Issuer {
		return fmt.Errorf("OIDC token issuer %q does not match configured issuer %q", issuer, p.Issuer)
	}
	return nil
}
