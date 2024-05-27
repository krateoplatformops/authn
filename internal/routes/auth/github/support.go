package github

import (
	"context"
	"fmt"

	"github.com/krateoplatformops/authn/internal/helpers/kube/resolvers"
	"github.com/krateoplatformops/authn/internal/helpers/kube/secrets"
	"golang.org/x/oauth2"
	"k8s.io/client-go/rest"
)

func getConfig(rc *rest.Config, name string) (*oauth2.Config, string, error) {
	ghc, err := resolvers.GetGithubConfig(rc, name)
	if err != nil {
		return nil, "", fmt.Errorf("unable to resolve Github configuration")
	}

	sec, err := secrets.Get(context.Background(), rc, ghc.Spec.ClientSecretRef)
	if err != nil {
		return nil, "", err
	}

	clientSecret, ok := sec.Data[ghc.Spec.ClientSecretRef.Key]
	if !ok {
		return nil, "", fmt.Errorf("client secret not found")
	}

	return &oauth2.Config{
		ClientID:     ghc.Spec.ClientID,
		ClientSecret: string(clientSecret),
		RedirectURL:  ghc.Spec.RedirectURL,
		Scopes:       ghc.Spec.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  ghc.Spec.AuthURL,
			TokenURL: ghc.Spec.TokenURL,
		},
	}, ghc.Spec.Organization, nil
}
