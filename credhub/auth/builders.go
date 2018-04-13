package auth

import (
	"net/http"

	"github.com/cloudfoundry-incubator/credhub-cli/credhub/auth/uaa"
)

// Config provides the CredHub configuration necessary to build an auth Strategy
//
// The credhub.CredHub struct conforms to this interface
type Config interface {
	AuthURL() (string, error)
	Client() *http.Client
}

// Builder constructs the auth type given a configuration
//
// A builder is required by the credhub.Auth() option for credhub.New()
type Builder func(config Config) (Strategy, error)

// Noop builds a NoopStrategy
var Noop Builder = func(config Config) (Strategy, error) {
	return &NoopStrategy{config.Client()}, nil
}

// Uaa builds an OauthStrategy for a UAA using existing tokens
func Uaa(clientId, clientSecret, username, password, accessToken, refreshToken string) Builder {
	return func(config Config) (Strategy, error) {
		httpClient := config.Client()
		authUrl, err := config.AuthURL()

		if err != nil {
			return nil, err
		}

		uaaClient := uaa.Client{
			AuthURL: authUrl,
			Client:  httpClient,
		}
		usingClientCredentials := clientSecret != ""
		oauth := &OAuthStrategy{
			Username:                username,
			Password:                password,
			ClientId:                clientId,
			ClientSecret:            clientSecret,
			ApiClient:               httpClient,
			OAuthClient:             &uaaClient,
			ClientCredentialRefresh: usingClientCredentials,
		}

		oauth.SetTokens(accessToken, refreshToken)

		return oauth, nil
	}
}
