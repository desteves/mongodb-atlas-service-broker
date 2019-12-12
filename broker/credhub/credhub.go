package credhub

import (
	"fmt"
	"log"
	"os"
	"strings"

	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/auth"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials/generate"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials/values"
	"code.cloudfoundry.org/credhub-cli/credhub/permissions"
)

type CredhubJSON struct {
	Username string `json:"username"`
	Password string `json:"password"`
	URI      string `json:"uri"`
}

type CFVCAPOps struct {
	CFCredhubURI string `json:"credhub-uri"`
}

func getClient() (*credhub.CredHub, error) {
	isCF := os.Getenv("VCAP_APPLICATION")
	credhubEndpoint := ""

	if len(isCF) == 0 {
		credhubEndpoint = os.Getenv("CREDHUB_URL")
		if len(credhubEndpoint) == 0 {
			err := fmt.Errorf("Failed to find credhub url in CREDHUB_URL: %s", credhubEndpoint)
			return nil, err
		}
	} else {
		credhubEndpoint = "https://credhub.service.cf.internal:8844"
	}

	credhubPass := os.Getenv("UAA_ADMIN_CLIENT_SECRET")
	if len(credhubPass) == 0 {
		err := fmt.Errorf("Failed to get credhub admin password, UAA_ADMIN_CLIENT_SECRET not set !")
		return nil, err
	}

	credhubClient, err := credhub.New(
		credhubEndpoint,
		credhub.SkipTLSValidation(true),
		credhub.Auth(auth.UaaPassword("credhub_cli", "", "admin", credhubPass)),
	)
	if err != nil {
		log.Printf("Failed to create credhub client! %v", err)
		return nil, err
	}
	return credhubClient, nil
}

func EnableAppAccess(appV4UUID string, credentialName string) error {
	log.Printf("appguid: %+v, credentialname: %+v", appV4UUID, credentialName)
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
			},
		},
	)
	if err != nil {
		log.Printf("failed to add permission to %s %s", actor, err)
		return err
	}
	return nil
}

func GetPath(instanceID string, bindingID string, variable string) string {
	//each new password will end up in /c/mongo_atlas_service_broker/[instance-id]/[binding-id]/credentials (or password)
	credentialPath := "/c/mongo_atlas_service_broker"
	credentialName := fmt.Sprintf("%s/%s/%s/%s", credentialPath, instanceID, bindingID, variable)

	return credentialName
}

func GenPassFromCredhub(instanceID string, bindingID string) (credentials.Password, error) {

	passwordOps := generate.Password{
		Length:         32,
		IncludeSpecial: false,
	}
	credentialName := GetPath(instanceID, bindingID, "password")
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

func StoreJSON(instanceID string, bindingID string, mongoURI string) (credentials.JSON, error) {
	password, err := GenPassFromCredhub(instanceID, bindingID)
	if err != nil {
		log.Printf("failed to generate password from credhub %+v", err)
		return credentials.JSON{}, err
	}
	credhubClient, err := getClient()
	if err != nil {
		log.Printf("failed to connect to credhub %s", err)
		return credentials.JSON{}, err
	}
	credentialName := GetPath(instanceID, bindingID, "credential")
	pathSeg := strings.Split(mongoURI, "//")

	parsedMongoURI := fmt.Sprintf("%s//%s:%s@%s", pathSeg[0], bindingID, string(password.Value), pathSeg[1])

	pathSeg = strings.Split(parsedMongoURI, "?")
	parsedMongoURIDB := fmt.Sprintf("%stest?%s", pathSeg[0], pathSeg[1])
	m := values.JSON{}
	m["username"] = bindingID
	m["password"] = string(password.Value)
	m["uri"] = parsedMongoURIDB

	json, err := credhubClient.SetJSON(credentialName, m, credhub.Overwrite)
	if err != nil {
		log.Printf("failed to setJSON in credhub %+v %s", m, err)
		return credentials.JSON{}, err
	}
	return json, err
}

func GetPassFromCredhub(instanceID string, bindingID string) (credentials.Password, error) {

	credhubClient, err := getClient()
	if err != nil {
		log.Printf("failed to connect to credhub %s", err)
		return credentials.Password{}, err
	}

	credentialName := GetPath(instanceID, bindingID, "password")
	password, err := credhubClient.GetLatestPassword(credentialName)
	if err != nil {
		log.Printf("failed to get password from credhub with path %s", credentialName)
		return credentials.Password{}, err
	}
	return password, nil
}

func DeleteJSONFromCredhub(instanceID string, bindingID string) error {

	credhubClient, err := getClient()
	if err != nil {
		log.Printf("failed to connect to credhub %s", err)
		return err
	}
	credentialName := GetPath(instanceID, bindingID, "credential")
	err = credhubClient.Delete(credentialName)
	if err != nil {
		log.Printf("failed to delete %s from credhub", credentialName)
		return err
	}
	err = DeletePassFromCredhub(instanceID, bindingID)
	if err != nil {
		log.Printf("failed to DeletePassFromCredhub from credhub")
		return err
	}
	return nil
}

func DeletePassFromCredhub(instanceID string, bindingID string) error {

	credhubClient, err := getClient()
	if err != nil {
		log.Printf("failed to connect to credhub %s", err)
		return err
	}
	credentialName := GetPath(instanceID, bindingID, "password")
	err = credhubClient.Delete(credentialName)
	if err != nil {
		log.Printf("failed to delete %s from credhub", credentialName)
		return err
	}
	return nil
}
