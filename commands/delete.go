package commands

import (
	"code.cloudfoundry.org/credhub-cli/errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type DeleteCommand struct {
	CredentialIdentifier string `short:"n" long:"name" description:"Name of the credential to delete"`
	CredentialPath       string `short:"p" long:"path" description:"Path of the credentials to delete"`
	ClientCommand
}

func (c *DeleteCommand) Execute([]string) error {
	if c.CredentialIdentifier != "" {
		return c.handleDeleteByName()
	} else if c.CredentialPath != "" {
		return c.handleDeleteByPath()
	}

	return errors.NewMissingDeleteParametersError()
}

func (c *DeleteCommand) handleDeleteByName() error {
	err := c.client.DeleteByName(c.CredentialIdentifier)

	if err == nil {
		fmt.Println("Credential successfully deleted")
	}

	return err
}

func (c *DeleteCommand) handleDeleteByPath() error {
	failedCredentials, err := c.client.DeleteByPath(c.CredentialPath)

	if err != nil {
		return err
	}

	if len(failedCredentials) == 0 {
		fmt.Println("\nAll credentials successfully deleted.")
	} else {
		fmt.Fprintln(os.Stderr, "\nThe following credentials failed to delete:")

		s, _ := yaml.Marshal(failedCredentials)
		fmt.Fprint(os.Stderr, string(s))
	}

	return nil
}
