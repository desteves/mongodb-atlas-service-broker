package credhub

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/auth"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials/generate"
	"code.cloudfoundry.org/credhub-cli/credhub/permissions"
)

func getClient() (*credhub.CredHub, error) {
	vcapPlatformOps := os.Getenv("VCAP_PLATFORM_OPTIONS")
	credhubEndpoint := ""
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(vcapPlatformOps), &jsonMap)
	if err != nil {
		credhubEndpoint = os.Getenv("CREDHUB_URL")
		if len(credhubEndpoint) == 0 {
			log.Printf("Failed to find credhub url! %v", err)
			return nil, err
		}
	} else {
		temp, ok := jsonMap["credhub-uri"].(string)
		if ok {
			credhubEndpoint = temp
		} else {
			log.Printf("Failed to find credhub in VCAP_PLATFORM_OPTIONS! %v", err)
			return nil, err
		}
	}

	credhubUser := ""
	credhubPass := ""

	credhubUser = os.Getenv("CF_ADMIN_USER")
	if len(credhubUser) == 0 {
		credhubUser = os.Getenv("CREDHUB_ADMIN_USERNAME")
		if len(credhubUser) == 0 {
			log.Printf("Failed to get credhub admin username,CF_ADMIN_USER, and CREDHUB_ADMIN_USERNAME not set !")
			return nil, err
		}
	}

	credhubPass = os.Getenv("CF_ADMIN_PASSWORD")
	if len(credhubPass) == 0 {
		credhubPass = os.Getenv("CREDHUB_ADMIN_PASSWORD")
		if len(credhubPass) == 0 {
			log.Printf("Failed to get credhub admin password, CF_ADMIN_PASSWORD, and CREDHUB_ADMIN_PASSWORD not set !")
			return nil, err
		}
	}

	credhubClient, err := credhub.New(
		credhubEndpoint,
		credhub.SkipTLSValidation(true),
		credhub.Auth(auth.UaaPassword("admin", "", credhubUser, credhubPass)),
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
	_, err = credhubClient.AddPermissions(
		credentialName,
		[]permissions.Permission{
			permissions.Permission{
				Actor:      actor,
				Operations: []string{"read", "read_acl"},
				Path:       credentialName,
			},
		},
	)
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
