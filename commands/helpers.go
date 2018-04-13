package commands

import (
"encoding/json"
"fmt"
"net/http"




"github.com/cloudfoundry-incubator/credhub-cli/config"
"github.com/cloudfoundry-incubator/credhub-cli/credhub"
"gopkg.in/yaml.v2"

)

func printCredential(outputJson bool, v interface{}) {
	if outputJson {
		s, _ := json.MarshalIndent(v, "", "\t")
		fmt.Println(string(s))
	} else {
		s, _ := yaml.Marshal(v)
		fmt.Println(string(s))
	}
}

func verifyAuthServerConnection(cfg config.Config, skipTlsValidation bool) error {
	credhubClient, err := credhub.New(cfg.ApiURL, credhub.CaCerts(cfg.CaCerts...), credhub.SkipTLSValidation(skipTlsValidation))
	if err != nil {
		return err
	}
	if !skipTlsValidation {
		request, _ := http.NewRequest("GET", cfg.AuthURL+"/info", nil)
		request.Header.Add("Accept", "application/json")
		_, err = credhubClient.Client().Do(request)
	}

	return err
}
