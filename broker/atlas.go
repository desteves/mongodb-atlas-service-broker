package broker

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// TODO -- Move this to CredHub/Env vars, don't leave this exposed here
// #shame
// /giphy shame
// :shame:

// cost - move to CredHUB
const (
	Host  = "https://cloud.mongodb.com"
	URI   = "/api/atlas/v1.0"
	User  = "diana.esteves"
	Pass  = "787b61f2-9476-4ed5-963c-5570e13720bc"
	Group = "5b75e84b3b34b9469d01b20e"

	OperationDeprovision = "deprovision"
	OperationProvision   = "provision"

	StateIDLE      = "IDLE"
	StateCREATING  = "CREATING"
	StateUPDATING  = "UPDATING"
	StateDELETING  = "DELETING"
	StateDELETED   = "DELETED"
	StateREPAIRING = "REPAIRING"

	ErrorCode404 = "CLUSTER_NOT_FOUND"
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

// Horrible to have the region as the field name, ughr!!!!
// Skipping the marshaling of this sub-doc altogether!
// ReplicationSpec       struct {
// 	USEAST1 struct { /// shoot me please
// 		ElectableNodes int `json:"electableNodes"`
// 		Priority       int `json:"priority"`
// 		ReadOnlyNodes  int `json:"readOnlyNodes"`
// 	} `json:"US_EAST_1"`
// } `json:"replicationSpec"`

//ProvisionResponse struct
type ProvisionResponse struct {
	AutoScaling              AutoScaling      `json:"autoScaling,omitempty"`
	BackupEnabled            bool             `json:"backupEnabled,omitempty"`
	BiConnector              BiConnector      `json:"biConnector,omitempty"`
	ClusterType              string           `json:"clusterType,omitempty"`
	DiskSizeGB               float32          `json:"diskSizeGB,omitempty"`
	EncryptionAtRestProvider string           `json:"encryptionAtRestProvider,omitempty"`
	GroupID                  string           `json:"groupId,omitempty"`
	ID                       string           `json:"id,omitempty"`
	Links                    []Links          `json:"links,omitempty"`
	MongoDBMajorVersion      string           `json:"mongoDBMajorVersion,omitempty"`
	MongoDBVersion           string           `json:"mongoDBVersion,omitempty"`
	MongoURI                 string           `json:"mongoURI,omitempty"`
	MongoURIUpdated          time.Time        `json:"mongoURIUpdated,omitempty"`
	MongoURIWithOptions      string           `json:"mongoURIWithOptions,omitempty"`
	Name                     string           `json:"name,omitempty"`
	NumShards                int              `json:"numShards,omitempty"`
	Paused                   bool             `json:"paused,omitempty"`
	ProviderBackupEnabled    bool             `json:"providerBackupEnabled,omitempty"`
	ProviderSettings         ProviderSettings `json:"providerSettings,omitempty"`
	ReplicationFactor        int              `json:"replicationFactor,omitempty"`
	StateName                string           `json:"stateName,omitempty"`

	// deal with error json response
	Detail     string   `json:"detail,omitempty"`
	Error      int      `json:"error,omitempty"`
	ErrorCode  string   `json:"errorCode"`
	Parameters []string `json:"parameters,omitempty"`
	Reason     string   `json:"reason,omitempty"`
}

//DeprovisionResponse struct
type DeprovisionResponse struct {
}

//LastOperationResponse struct
type LastOperationResponse struct {
	AutoScaling              AutoScaling      `json:"autoScaling,omitempty"`
	BackupEnabled            bool             `json:"backupEnabled,omitempty"`
	BiConnector              BiConnector      `json:"biConnector,omitempty"`
	ClusterType              string           `json:"clusterType,omitempty"`
	DiskSizeGB               float32          `json:"diskSizeGB,omitempty"`
	EncryptionAtRestProvider string           `json:"encryptionAtRestProvider,omitempty"`
	GroupID                  string           `json:"groupId,omitempty"`
	MongoDBVersion           string           `json:"mongoDBVersion,omitempty"`
	MongoURI                 string           `json:"mongoURI,omitempty,omitempty"`
	MongoURIUpdated          string           `json:"mongoURIUpdated,omitempty,omitempty"`
	MongoURIWithOptions      string           `json:"mongoURIWithOptions,omitempty"`
	Name                     string           `json:"name,omitempty"`
	NumShards                int              `json:"numShards,omitempty"`
	Paused                   bool             `json:"paused,omitempty"`
	ProviderSettings         ProviderSettings `json:"providerSetting,omitemptys"`
	ReplicationFactor        int              `json:"replicationFactor,omitempty"`
	StateName                string           `json:"stateName"`

	// deal with error json response
	Detail     string   `json:"detail,omitempty"`
	Error      int      `json:"error,omitempty"`
	ErrorCode  string   `json:"errorCode"`
	Parameters []string `json:"parameters,omitempty"`
	Reason     string   `json:"reason,omitempty"`
}

// Does the digest handshake and assembles the actuall http call to make -- TODO move to client
func setupRequest(argMethod string, argURI string, argPostBody []byte) (*http.Request, error) {
	uri := URI + argURI
	url := Host + uri
	emptyRequest := http.Request{}
	req, err := http.NewRequest(argMethod, url, nil)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error - setupRequest - Failed http response. Resp: %+v, Err: %+v", resp, err)
		return &emptyRequest, err
	}
	defer resp.Body.Close()
	digestParts := digestParts(resp)
	digestParts["uri"] = uri
	digestParts["method"] = argMethod
	digestParts["username"] = User
	digestParts["password"] = Pass
	if argPostBody == nil {
		req, err = http.NewRequest(argMethod, url, nil)
	} else {
		req, err = http.NewRequest(argMethod, url, bytes.NewBuffer(argPostBody))
	}
	req.Header.Set("Authorization", getDigestAuthrization(digestParts))
	req.Header.Set("Content-Type", "application/json")
	return req, nil

}

//DoPOST using the MongoDB Atlas REST API -- TODO move to client
func DoPOST(argURI string, argPostBody []byte) ([]byte, error) {
	req, err := setupRequest(http.MethodPost, argURI, argPostBody)
	if err != nil {
		log.Printf("Error - DoPOST - Failed setupRequest call. Req: %+v, Err: %+v", req, err)
		return []byte{}, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error - DoPOST - Failed http response. Resp: %+v, Err: %+v", resp, err)
		return []byte{}, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error - DoPOST - Failed parsing body response. Data: %+v, Err: %+v", data, err)
		return []byte{}, err
	}
	return data, nil
}

//DoGET using the MongoDB Atlas REST API -- TODO move to client
func DoGET(argURI string) ([]byte, error) {
	req, err := setupRequest(http.MethodGet, argURI, []byte{})
	if err != nil {
		log.Printf("Error - DoGET - Failed setupRequest call. Req: %+v, Err: %+v", req, err)
		return []byte{}, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error - DoGET - Failed http response. Resp: %+v, Err: %+v", resp, err)
		return []byte{}, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error - DoGET - Failed parsing body response. Data: %+v, Err: %+v", data, err)
		return []byte{}, err
	}
	return data, nil
}

//DoDELETE using the MongoDB Atlas REST API -- TODO move to client
func DoDELETE(argURI string) ([]byte, error) {
	req, err := setupRequest(http.MethodDelete, argURI, nil)
	if err != nil {
		log.Printf("Error - DoDELETE - Failed setupRequest call. Req: %+v, Err: %+v", req, err)
		return []byte{}, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error - DoDELETE - Failed http response. Resp: %+v, Err: %+v", resp, err)
		return []byte{}, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error - DoDELETE - Failed parsing body response. Data: %+v, Err: %+v", data, err)
		return []byte{}, err
	}
	return data, nil
}

//NewCluster in MongoDB Atlas
func NewCluster(argPostBody []byte) (ProvisionResponse, error) {
	returnObject := ProvisionResponse{}
	uri := "/groups/" + Group + "/clusters"
	body, err := DoPOST(uri, argPostBody)
	if err != nil {
		log.Printf("Error - NewCluster - Failed DoPOST call.  Body: %+v, Err: %+v", body, err)
		return ProvisionResponse{}, err
	}
	err = json.Unmarshal(body, &returnObject)
	if err != nil {
		log.Printf("Error - NewCluster - Failed Unmarshal. ProvisionResponse: %+v, Err: %+v", returnObject, err)
		return ProvisionResponse{}, err
	}
	return returnObject, err
}

//GetCluster in MongoDB Atlas
func GetCluster(instanceID string) (LastOperationResponse, error) {
	returnObject := LastOperationResponse{}
	uri := "/groups/" + Group + "/clusters/" + instanceID
	body, err := DoGET(uri)
	if err != nil {
		log.Printf("Error - GetCluster - Failed DoGET call.  Body: %+v, Err: %+v", body, err)
		return returnObject, err
	}
	err = json.Unmarshal(body, &returnObject)
	if err != nil {
		log.Printf("Error - GetCluster - Failed Unmarshal. LastOperationResponse: %+v, Err: %+v", returnObject, err)
		return LastOperationResponse{}, err
	}
	return returnObject, err
}

//TerminateCluster in MongoDB Atlas
func TerminateCluster(instanceID string) (DeprovisionResponse, error) {
	uri := "/groups/" + Group + "/clusters/" + instanceID
	_, err := DoDELETE(uri)
	if err != nil {
		log.Printf("Error - TerminateCluster - Failed DoDELETE call.  Err: %+v", err)
	}
	return DeprovisionResponse{}, err
}

//
//  --- TODO ---
// 	bindings, _ := repo.instanceBindings[instanceID]
// 	found := false
// 	for _, binding := range bindings {
// 		if binding != bindingID {
// 			newInstanceBindings = append(newInstanceBindings, binding)
// 		} else {
// 			found = true
// 	repo.instanceBindings[instanceID] = newInstanceBindings
