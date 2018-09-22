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

func GenPassFromCredhub(instanceID string, bindingID string) (credentials.Password, error) {
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
		return credentials.Password{}, err
	}
	passwordOps := generate.Password{
		Length:         32,
		IncludeSpecial: false,
	}
	//each new password will end up in /mongodbatlas/servicebroker/<INSTANCEID>/<BINDINGID>
	credentialPath := "/mongodbatlas/servicebroker"
	credentialName := fmt.Sprintf("%s/%s/%s", credentialPath, instanceID, bindingID)
	password, err := credhubClient.GeneratePassword(credentialName, passwordOps, credhub.NoOverwrite)
	if err != nil {
		log.Printf("failed to generate password from credhub with path %s", credentialName)
		return credentials.Password{}, err
	}
	return password, nil
}

func GetPassFromCredhub(instanceID string, bindingID string) (credentials.Password, error) {
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
		return credentials.Password{}, err
	}

	//each password will end up in /mongodbatlas/servicebroker/<INSTANCEID>/<BINDINGID>
	credentialPath := "/mongodbatlas/servicebroker"
	credentialName := fmt.Sprintf("%s/%s/%s", credentialPath, instanceID, bindingID)
	password, err := credhubClient.GetLatestPassword(credentialName)
	if err != nil {
		log.Printf("failed to get password from credhub with path %s", credentialName)
		return credentials.Password{}, err
	}
	return password, nil
}

func DeletePassFromCredhub(instanceID string, bindingID string) error {
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
		return err
	}

	//each password will end up in /mongodbatlas/servicebroker/<INSTANCEID>/<BINDINGID>
	credentialPath := "/mongodbatlas/servicebroker"
	credentialName := fmt.Sprintf("%s/%s/%s", credentialPath, instanceID, bindingID)
	err = credhubClient.Delete(credentialName)
	if err != nil {
		log.Printf("failed to delete %s from credhub", credentialName)
		return err
	}
	return nil
}
