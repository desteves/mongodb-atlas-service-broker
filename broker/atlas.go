package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// TODO -- Move this to CredHub/Env vars, don't leave this exposed here #shame
const (
	Host  = "https://cloud.mongodb.com"
	URI   = "/api/atlas/v1.0"
	User  = "diana.esteves"
	Pass  = "787b61f2-9476-4ed5-963c-5570e13720bc"
	Group = "5b75e84b3b34b9469d01b20e"

	StateIDLE      = "IDLE"
	StateCREATING  = "CREATING"
	StateUPDATING  = "UPDATING"
	StateDELETING  = "DELETING"
	StateDELETED   = "DELETED"
	StateREPAIRING = "REPAIRING"
)

// AutoScaling - Provision Setting
type AutoScaling struct {
	DiskGBEnabled bool `json:"diskGBEnabled"`
}

// ProviderSettings - Provision Setting
type ProviderSettings struct {
	DiskIOPS         int    `json:"diskIOPS"`
	EncryptEBSVolume bool   `json:"encryptEBSVolume"`
	InstanceSizeName string `json:"instanceSizeName"`
	ProviderName     string `json:"providerName"`
	RegionName       string `json:"regionName"`
}

//Provision Setting
type Provision struct {
	AutoScaling              AutoScaling      `json:"autoScaling"`
	BackupEnabled            bool             `json:"backupEnabled"`
	DiskSizeGB               float32          `json:"diskSizeGB"`
	EncryptionAtRestProvider string           `json:"encryptionAtRestProvider"`
	Name                     string           `json:"name"`
	NumShards                int              `json:"numShards"`
	ReplicationFactor        int              `json:"replicationFactor"`
	ProviderSettings         ProviderSettings `json:"providerSettings"`
}

//BiConnector -
type BiConnector struct {
	Enabled        bool   `json:"enabled"`
	ReadPreference string `json:"readPreference"`
}

//Links --
type Links struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// ReplicationSpec       struct {
// 	USEAST1 struct { // Horrible to have the region as the field name, ughr!!!! skipping the marshaling of this sub-doc altogether
// 		ElectableNodes int `json:"electableNodes"`
// 		Priority       int `json:"priority"`
// 		ReadOnlyNodes  int `json:"readOnlyNodes"`
// 	} `json:"US_EAST_1"`
// } `json:"replicationSpec"`

//ProvisionResponse struct
type ProvisionResponse struct {
	AutoScaling              AutoScaling      `json:"autoScaling"`
	BackupEnabled            bool             `json:"backupEnabled"`
	BiConnector              BiConnector      `json:"biConnector"`
	ClusterType              string           `json:"clusterType"`
	DiskSizeGB               float32          `json:"diskSizeGB"`
	EncryptionAtRestProvider string           `json:"encryptionAtRestProvider"`
	GroupID                  string           `json:"groupId"`
	ID                       string           `json:"id"`
	Links                    []Links          `json:"links"`
	MongoDBMajorVersion      string           `json:"mongoDBMajorVersion"`
	MongoDBVersion           string           `json:"mongoDBVersion"`
	MongoURI                 string           `json:"mongoURI"`
	MongoURIUpdated          time.Time        `json:"mongoURIUpdated"`
	MongoURIWithOptions      string           `json:"mongoURIWithOptions"`
	Name                     string           `json:"name"`
	NumShards                int              `json:"numShards"`
	Paused                   bool             `json:"paused"`
	ProviderBackupEnabled    bool             `json:"providerBackupEnabled"`
	ProviderSettings         ProviderSettings `json:"providerSettings"`
	ReplicationFactor        int              `json:"replicationFactor"`
	StateName                string           `json:"stateName"`
}

type DeprovisionResponse struct{

}
//LastOperationResponse struct
type LastOperationResponse struct {
	AutoScaling              AutoScaling      `json:"autoScaling"`
	BackupEnabled            bool             `json:"backupEnabled"`
	BiConnector              BiConnector      `json:"biConnector"`
	ClusterType              string           `json:"clusterType"`
	DiskSizeGB               int              `json:"diskSizeGB"`
	EncryptionAtRestProvider string           `json:"encryptionAtRestProvider"`
	GroupID                  string           `json:"groupId"`
	MongoDBVersion           string           `json:"mongoDBVersion"`
	MongoURI                 string           `json:"mongoURI"`
	MongoURIUpdated          string           `json:"mongoURIUpdated"`
	MongoURIWithOptions      string           `json:"mongoURIWithOptions"`
	Name                     string           `json:"name"`
	NumShards                int              `json:"numShards"`
	Paused                   bool             `json:"paused"`
	ProviderSettings         ProviderSettings `json:"providerSettings"`
	ReplicationFactor        int              `json:"replicationFactor"`
	StateName                string           `json:"stateName"`
}

//DoPOST using the MongoDB Atlas REST API
func DoPOST(argURI string, argPostBody []byte) ([]byte, error) {
	req, err := setupRequest(http.MethodPost, argURI, argPostBody)
	if err != nil {
		log.Printf("Failed setting up POST request...%+v", req)
		return []byte{}, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed setting up POST client.Do(req)...%+v", resp)
		return []byte{}, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		return []byte{}, err
	}
	return data, nil
}

func setupRequest(argMethod string, argURI string, argPostBody []byte) (*http.Request, error) {
	uri := URI + argURI
	url := Host + uri
	emptyRequest := http.Request{}
	req, err := http.NewRequest(argMethod, url, nil)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &emptyRequest, err
	}
	defer resp.Body.Close()
	digestParts := digestParts(resp)
	digestParts["uri"] = uri
	digestParts["method"] = argMethod
	digestParts["username"] = User
	digestParts["password"] = Pass
	req, err = http.NewRequest(argMethod, url, bytes.NewBuffer(argPostBody))
	req.Header.Set("Authorization", getDigestAuthrization(digestParts))
	req.Header.Set("Content-Type", "application/json")
	return req, nil

}

//DoGET using the MongoDB Atlas REST API -- TODO TEST
func DoGET(argURI string) ([]byte, error) {
	req, err := setupRequest(http.MethodGet, argURI, []byte{})
	if err != nil {
		log.Printf("Failed setting up GET request...%+v", req)
		return []byte{}, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed setting up GET client.Do(req)...%+v", resp)
		return []byte{}, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

//DoDELETE using the MongoDB Atlas REST API
func DoDELETE(argURI string, argPostBody []byte) ([]byte, error) {
	req, err := setupRequest(http.MethodDelete, argURI, argPostBody)
	if err != nil {
		log.Printf("Failed setting up DELETE request...%+v", req)
		return []byte{}, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed setting up DELETE client.Do(req)...%+v", resp)
		return []byte{}, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

//NewCluster in MongoDB Atlas
func NewCluster(argPostBody []byte) (ProvisionResponse, error) {
	returnObject := ProvisionResponse{}
	uri := "/groups/" + Group + "/clusters"
	body, err := DoPOST(uri, argPostBody)
	fmt.Printf("Sending POST to %s...%+v", uri, string(argPostBody))
	if err != nil {
		fmt.Printf("Error Sending POST to %s...%+v", uri, string(argPostBody))
		return ProvisionResponse{}, err
	}
	err = json.Unmarshal(body, &returnObject)
	if err != nil {
		fmt.Printf("Error Unmarshalling NewCluster Response...%+v", string(body))
		return ProvisionResponse{}, err
	}
	log.Println("NewCluster body: ", string(body))
	return returnObject, err
}

//GetCluster in MongoDB Atlas   -- TEST TODO
func GetCluster(instanceID string) (LastOperationResponse, error) {
	returnObject := LastOperationResponse{}
	uri := "/groups/" + Group + "/clusters" + instanceID
	body, err := DoGET(uri)
	if err != nil {
		fmt.Printf("Error Sending GetCluster GET to %s..", uri)
		return returnObject, err
	}
	err = json.Unmarshal(body, &returnObject)
	if err != nil {
		fmt.Printf("Error Unmarshalling GetCluster Response...%+v", string(body))
		return returnObject, err
	}
	log.Println("GetCluster body: ", string(body))
	return returnObject, err
}

//TerminateCluster in MongoDB Atlas -- TEST TODO
func TerminateCluster(instanceID string) (DeprovisionResponse, error) {
	returnObject := DeprovisionResponse{}
	uri := "/groups/" + Group + "/clusters/" + instanceID
	body, err := DoDELETE(uri)
	fmt.Printf("Sending DELETE to %s..", uri)
	if err != nil {
		fmt.Printf("Error Sending DELETE to %s.", uri))
		return DeprovisionResponse{}, err
	}
	err = json.Unmarshal(body, &returnObject)
	if err != nil {
		return DeprovisionResponse{}, err
	}
	log.Println("TerminateCluster body: ", string(body))
	return returnObject, err
}

//
//
//
//
//
//
//

// 	stateBytes, err := ioutil.ReadFile(repo.statefilePath)
// 	if err != nil {
// 		repo.logger.Error(
// 			"failed to read statefile",
// 			err, lager.Data{"statefilePath": repo.statefilePath},
// 		)
// 		return statefileContents, err
// 	}

// 	err = json.Unmarshal(stateBytes, &statefileContents)
// 	if err != nil {
// 		repo.logger.Error(
// 			"failed to read statefile due to invalid JSON",
// 			err,
// 			lager.Data{
// 				"statefilePath":     repo.statefilePath,
// 				"stateFileContents": string(stateBytes),
// 			},
// 		)
// 		return statefileContents, err

// 	bindings, _ := repo.instanceBindings[instanceID]
// 	found := false
// 	for _, binding := range bindings {
// 		if binding != bindingID {
// 			newInstanceBindings = append(newInstanceBindings, binding)
// 		} else {
// 			found = true

// 	repo.instanceBindings[instanceID] = newInstanceBindings
