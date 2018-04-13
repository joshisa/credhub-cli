package commands

import (
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/credhub-cli/config"
	"github.com/cloudfoundry-incubator/credhub-cli/credhub/auth"
)

func init() {
	CredHub.Token = func() {
		cfg := config.ReadConfig()

		if cfg.AccessToken != "" && cfg.AccessToken != "revoked" {
			cfg = refreshConfiguration(cfg)
			config.WriteConfig(cfg)
			fmt.Println("Bearer " + cfg.AccessToken)
		} else if os.Getenv("CREDHUB_CLIENT") != "" && os.Getenv("CREDHUB_SECRET") != "" {
			cfg = refreshConfiguration(cfg)
			fmt.Println("Bearer " + cfg.AccessToken)
		} else {
			fmt.Fprint(os.Stderr, "You are not currently authenticated. Please log in to continue.")
		}
		os.Exit(0)
	}
}

func refreshConfiguration(cfg config.Config) config.Config {
	credhubClient, err := BuildClient()
	if err != nil {
		return cfg
	}
	authObject := credhubClient.Auth
	oauth := authObject.(*auth.OAuthStrategy)
	if err := oauth.Refresh(); err != nil {
		fmt.Println("ERRRRRRRR", err)
		fmt.Println("Bearer " + cfg.AccessToken)
	}

	cfg.AccessToken = oauth.AccessToken()
	cfg.RefreshToken = oauth.RefreshToken()
	return cfg
}
