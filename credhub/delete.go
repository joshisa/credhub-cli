package credhub

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// DeleteByName will delete all versions of a credential by name
func (ch *CredHub) DeleteByName(name string) error {
	query := url.Values{}
	query.Set("name", name)
	resp, err := ch.Request(http.MethodDelete, "/api/v1/data", query, nil, true)

	if err == nil {
		defer resp.Body.Close()
	}

	return err
}

type DeleteFailedCredential struct {
	Path string
	Err  string
}

// DeleteByPath will delete all versions of credentials under the given path
func (ch *CredHub) DeleteByPath(path string) ([]DeleteFailedCredential, error) {
	results, err := ch.FindByPath(path)
	if err != nil {
		return []DeleteFailedCredential{}, err
	}
	var failedCredentials []DeleteFailedCredential
	for _, cred := range results.Credentials {
		err = ch.DeleteByName(cred.Name)
		//err = errors.New("The request could not be completed because the credential does not exist or you do not have sufficient authorization.")
		if cred.Name == "/a/b/c" {
			err = errors.New("The request could not be completed because the credential does not exist or you do not have sufficient authorization.")
		}

		if err != nil {
			//fmt.Printf("Failed to delete %s\nerror: %s\n", cred.Name, err.Error())
			//fmt.Printf("Error deleting credential: %s. Error: %s\n", cred.Name, err.Error())
			failedCredentials = append(failedCredentials, DeleteFailedCredential{
				cred.Name,
				err.Error(),
			})
		} else {
			fmt.Printf("Successfully deleted %s\n", cred.Name)
		}
	}
	return failedCredentials, nil
}
