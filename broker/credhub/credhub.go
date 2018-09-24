package credhub

import (
	"fmt"
	"log"
	"os"

	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/auth"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials/generate"
)

func getClient() (*credhub.CredHub, error) {
	credhubEndpoint := os.Getenv("CREDHUB_URL")
	credhubClientUser := os.Getenv("CREDHUB_ADMIN_CLIENT")
	credhubClientSecret := os.Getenv("CREDHUB_CLIENT_SECRET")
	credhubClient, err := credhub.New(
		credhubEndpoint,
		credhub.SkipTLSValidation(true),
		credhub.Auth(auth.UaaClientCredentials(credhubClientUser, credhubClientSecret)),
	)
	if err != nil {
		log.Printf("Failed to create credhub client! %v", err)
		return nil, err
	}
	return credhubClient, nil
}

func EnableAppAccess(appV4UUID string, credentialName string) error {
	credhubClient, err := getClient()
	if err != nil {
		log.Printf("failed to connect to credhub %s", err)
		return err
	}
	actor := fmt.Sprintf("mtls-app:%s", appV4UUID)
	_, err = credhubClient.AddPermission(credentialName, actor, []string{"read", "read_acl"})
	if err != nil {
		log.Printf("failed to add permission to %s %s", actor, err)
		return err
	}
	return nil
}

func GetPath(instanceID string, bindingID string) string {
	//each new password will end up in /mongodbatlas/servicebroker/<INSTANCEID>/<BINDINGID>
	credentialPath := "/mongodbatlas/servicebroker"
	credentialName := fmt.Sprintf("%s/%s/%s", credentialPath, instanceID, bindingID)

	return credentialName
}

func GenPassFromCredhub(instanceID string, bindingID string) (credentials.Password, error) {

	passwordOps := generate.Password{
		Length:         32,
		IncludeSpecial: false,
	}
	credentialName := GetPath(instanceID, bindingID)
	credhubClient, err := getClient()
	if err != nil {
		log.Printf("failed to connect to credhub %s", err)
		return credentials.Password{}, err
	}
	password, err := credhubClient.GeneratePassword(credentialName, passwordOps, credhub.NoOverwrite)
	if err != nil {
		log.Printf("failed to generate password from credhub with path %s", credentialName)
		return credentials.Password{}, err
	}
	return password, nil
}

func GetPassFromCredhub(instanceID string, bindingID string) (credentials.Password, error) {

	credhubClient, err := getClient()
	if err != nil {
		log.Printf("failed to connect to credhub %s", err)
		return credentials.Password{}, err
	}

	credentialName := GetPath(instanceID, bindingID)
	password, err := credhubClient.GetLatestPassword(credentialName)
	if err != nil {
		log.Printf("failed to get password from credhub with path %s", credentialName)
		return credentials.Password{}, err
	}
	return password, nil
}

func DeletePassFromCredhub(instanceID string, bindingID string) error {

	credhubClient, err := getClient()
	if err != nil {
		log.Printf("failed to connect to credhub %s", err)
		return err
	}
	credentialName := GetPath(instanceID, bindingID)
	err = credhubClient.Delete(credentialName)
	if err != nil {
		log.Printf("failed to delete %s from credhub", credentialName)
		return err
	}
	return nil
}
