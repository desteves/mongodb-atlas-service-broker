package broker

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// const
const (
	OperationDeprovision = "deprovision"
	OperationProvision   = "provision"
	OperationBind        = "bind"
	OperationUnbind      = "unbind"

	StateIDLE      = "IDLE"
	StateCREATING  = "CREATING"
	StateUPDATING  = "UPDATING"
	StateDELETING  = "DELETING"
	StateDELETED   = "DELETED"
	StateREPAIRING = "REPAIRING"

	ErrorCode404 = "CLUSTER_NOT_FOUND"

	UserDatabaseStore = "admin"
	UserRoleDatabase  = "admin"
	UserRoleName      = "readWriteAnyDatabase"
)

type atlasCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	URI      string `json:"url"`
}

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

// Roles --
type Roles struct {
	CollectionName string `json:"collectionName,omitempty"`
	DatabaseName   string `json:"databaseName"`
	RoleName       string `json:"roleName"`
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

//UnbindResponse struct
type UnbindResponse struct {
}

//LastBindingOperationResponse struct -- Gets a single databse user
type LastBindingOperationResponse struct {
	DatabaseName    string  `json:"databaseName"`
	DeleteAfterDate string  `json:"deleteAfterDate,omitempty"`
	GroupID         string  `json:"groupId"`
	Links           []Links `json:"links"`
	Roles           []Roles `json:"roles"`
	Username        string  `json:"username"`

	//TODO may have to add fields to handle error json response -- CHECK
}

// TODO see if it contains the srv entry

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

// BindRequest struct - Bind Settings
type BindRequest struct {
	DatabaseName    string  `json:"databaseName"`
	Password        string  `json:"password"`
	Roles           []Roles `json:"roles"`
	Username        string  `json:"username"`
	DeleteAfterDate string  `json:"deleteAfterDate,omitempty"`
	GroupID         string  `json:"groupId"`
}

//BindResponse struct - Bind Settings
type BindResponse struct {
	DatabaseName    string  `json:"databaseName"`
	DeleteAfterDate string  `json:"deleteAfterDate,omitempty"`
	GroupID         string  `json:"groupId"`
	Links           []Links `json:"links"`
	Roles           []Roles `json:"roles"`
	Username        string  `json:"username"`

	//TODO may have to add fields to handle error json response -- CHECK
}

// Does the digest handshake and assembles the actuall http call to make -- TODO move to client
func setupRequest(argMethod string, argURI string, argPostBody []byte) (*http.Request, error) {
	uri := os.Getenv("ATLAS_URI") + argURI
	url := os.Getenv("ATLAS_HOST") + uri
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
	digestParts["username"] = os.Getenv("ATLAS_USER")
	digestParts["password"] = os.Getenv("ATLAS_PASS")
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
	uri := "/groups/" + os.Getenv("ATLAS_GROUP") + "/clusters"
	returnObject := ProvisionResponse{}
	body, err := DoPOST(uri, argPostBody)
	if err != nil {
		log.Printf("Error - NewCluster - Failed DoPOST call.  Body: %+v, Err: %+v", body, err)
		return ProvisionResponse{}, err
	}
	err = json.Unmarshal(body, &returnObject)
	if err != nil {
		log.Printf("Error - NewCluster - Failed Unmarshal. Response: %+v, Err: %+v", returnObject, err)
		return ProvisionResponse{}, err
	}
	return returnObject, err
}

//NewUser in MongoDB Atlas
func NewUser(argPostBody []byte) (BindResponse, error) {
	// https://docs.atlas.mongodb.com/reference/api/database-users-create-a-user/
	uri := "/groups/" + os.Getenv("ATLAS_GROUP") + "/databaseUsers"
	returnObject := BindResponse{}
	body, err := DoPOST(uri, argPostBody)
	if err != nil {
		log.Printf("Error - NewUser - Failed DoPOST call.  Body: %+v, Err: %+v", body, err)
		return returnObject, err
	}
	err = json.Unmarshal(body, &returnObject)
	if err != nil {
		log.Printf("Error - NewUser - Failed Unmarshal. Response: %+v, Err: %+v", returnObject, err)
		return returnObject, err
	}
	return returnObject, err
}

//GetCluster in MongoDB Atlas
func GetCluster(instanceID string) (LastOperationResponse, error) {
	returnObject := LastOperationResponse{}
	// https://docs.atlas.mongodb.com/reference/api/clusters-get-one/
	uri := "/groups/" + os.Getenv("ATLAS_GROUP") + "/clusters/" + instanceID
	body, err := DoGET(uri)
	if err != nil {
		log.Printf("Error - GetCluster - Failed DoGET call.  Body: %+v, Err: %+v", body, err)
		return returnObject, err
	}
	err = json.Unmarshal(body, &returnObject)
	if err != nil {
		log.Printf("Error - GetCluster - Failed Unmarshal. Response: %+v, Err: %+v", returnObject, err)
		return returnObject, err
	}
	return returnObject, nil
}

//GetUser in MongoDB Atlas
func GetUser(instanceID string, bindingID string) (LastBindingOperationResponse, error) {
	returnObject := LastBindingOperationResponse{}
	//https://docs.atlas.mongodb.com/reference/api/database-users-get-single-user/
	uri := "/groups/" + os.Getenv("ATLAS_GROUP") + "/databaseUsers/admin/" + bindingID
	body, err := DoGET(uri)
	if err != nil {
		log.Printf("Error - GetUser - Failed DoGET call.  Body: %+v, Err: %+v", body, err)
		return returnObject, err
	}
	err = json.Unmarshal(body, &returnObject)
	if err != nil {
		log.Printf("Error - GetUser - Failed Unmarshal. Response: %+v, Err: %+v", returnObject, err)
		return returnObject, err
	}
	return returnObject, nil
}

//TerminateCluster in MongoDB Atlas
func TerminateCluster(instanceID string) (DeprovisionResponse, error) {
	uri := "/groups/" + os.Getenv("ATLAS_GROUP") + "/clusters/" + instanceID
	_, err := DoDELETE(uri)
	if err != nil {
		log.Printf("Error - TerminateCluster - Failed DoDELETE call.  Err: %+v", err)
	}
	return DeprovisionResponse{}, err
}

//DeleteUser in MongoDB Atlas
func DeleteUser(instanceID string, bindingID string) (UnbindResponse, error) {
	//https://docs.atlas.mongodb.com/reference/api/database-users-delete-a-user/
	//DELETE /api/atlas/v1.0/groups/{GROUP-ID}/databaseUsers/admin/{USERNAME}
	uri := "/groups/" + os.Getenv("ATLAS_GROUP") + "/databaseUsers/admin/" + bindingID
	_, err := DoDELETE(uri)
	if err != nil {
		log.Printf("Error - DeleteUser - Failed DoDELETE call.  Err: %+v", err)
	}
	return UnbindResponse{}, err
}
