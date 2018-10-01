package credhub

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"encoding/json"

	"code.cloudfoundry.org/credhub-cli/credhub/permissions"
	"github.com/hashicorp/go-version"
)

type permissionsResponse struct {
	CredentialName string                   `json:"credential_name"`
	Permissions    []permissions.Permission `json:"permissions"`
}

func (ch *CredHub) getV1Permission(name string) ([]permissions.Permission, error) {
	query := url.Values{}
	query.Set("credential_name", name)

	resp, err := ch.Request(http.MethodGet, "/api/v1/permissions", query, nil, true)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)
	var response permissionsResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Permissions, nil
}

func (ch *CredHub) getV2Permission(uuid string) (*permissions.Permission, error) {
	path := "/api/v2/permissions/" + uuid

	resp, err := ch.Request(http.MethodGet, path, nil, nil, true)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)
	var response permissions.Permission

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (ch *CredHub) GetPermission(param string) (*permissions.Permission, error) {
	if ch.cachedServerVersion == "" {
		serverVersion, err := ch.ServerVersion()
		if err != nil {
			return nil, err
		}
		ch.cachedServerVersion = serverVersion.String()
	}

	serverVersion, err := version.NewVersion(ch.cachedServerVersion)
	if err != nil {
		return nil, err
	}

	if serverVersion.Segments()[0] < 2 {
		ch.getV1Permission(param)
	} else {
		ch.getV2Permission(param)
	}

	return nil, nil
}

func (ch *CredHub) AddV1Permissions(credName string, perms []permissions.Permission) ([]permissions.Permission, error) {
	requestBody := map[string]interface{}{}
	requestBody["credential_name"] = credName
	requestBody["permissions"] = perms

	_, err := ch.Request(http.MethodPost, "/api/v1/permissions", nil, requestBody, true)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (ch *CredHub) AddV2Permission(path string, actor string, ops []string) (*permissions.Permission, error) {
	requestBody := map[string]interface{}{}
	requestBody["path"] = path
	requestBody["actor"] = actor
	requestBody["operations"] = ops

	resp, err := ch.Request(http.MethodPost, "/api/v2/permissions", nil, requestBody, true)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)
	var response permissions.Permission

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (ch *CredHub) AddPermission(path string, actor string, ops []string) (*permissions.Permission, error) {
	if ch.cachedServerVersion == "" {
		serverVersion, err := ch.ServerVersion()
		if err != nil {
			return nil, err
		}
		ch.cachedServerVersion = serverVersion.String()
	}

	serverVersion, err := version.NewVersion(ch.cachedServerVersion)
	if err != nil {
		return nil, err
	}

	if serverVersion.Segments()[0] < 2 {
		ch.AddV1Permissions(path, []permissions.Permission{{Actor: actor, Operations: ops}})
	} else {
		ch.AddV2Permission(path, actor, ops)
	}

	return nil, nil
}
