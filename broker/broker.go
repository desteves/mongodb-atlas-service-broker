package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mongodb-atlas-service-broker/broker/credhub"
	"strings"

	"github.com/pivotal-cf/brokerapi"
)

const (
	awsDev = "aws_dev"
	gcpDev = "gcp_dev"
	custom = "custom"
)

// AtlasBroker - MongoDB Atlas Service Broker
type AtlasBroker struct {
}

// Services - MongoDB Atlas Broker
func (a AtlasBroker) Services(ctx context.Context) ([]brokerapi.Service, error) {

	trueVar := true
	falseVar := false
	awsDev := brokerapi.ServicePlan{
		Description: "AWS Dev Tier, Shared Cluster",
		Free:        &falseVar,
		Name:        awsDev,
		Bindable:    &trueVar,
		ID:          awsDev,
	}
	gcpDev := brokerapi.ServicePlan{
		Description: "GCP Dev Tier, Shared Cluster",
		Free:        &falseVar,
		Name:        gcpDev,
		Bindable:    &trueVar,
		ID:          gcpDev,
	}
	custom := brokerapi.ServicePlan{
		Description: "Custom Cluster",
		Free:        &falseVar,
		Name:        custom,
		Bindable:    &trueVar,
		ID:          custom,
	}

	s := brokerapi.Service{
		ID:            "atlas",
		Name:          "atlas",
		Description:   "MongoDB Atlas Service Broker",
		Bindable:      true,
		PlanUpdatable: true,
		Tags:          []string{"mongodb", "atlas"},
		Plans:         []brokerapi.ServicePlan{awsDev, gcpDev, custom},
	}
	return []brokerapi.Service{s}, nil
}

// Provision - MongoDB Atlas Broker
func (a AtlasBroker) Provision(ctx context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {

	instanceID32 := strings.Replace(instanceID, "-", "", -1)
	returnObject := brokerapi.ProvisionedServiceSpec{}
	provider := ProviderSettings{}

	if details.PlanID == custom {
		// TODO
		return returnObject, fmt.Errorf("life sucks")
	}

	// defaults
	provisionReq := Provision{
		Name:                     instanceID32,
		BackupEnabled:            false,
		ReplicationFactor:        3,
		EncryptionAtRestProvider: "NONE",
		AutoScaling:              AutoScaling{DiskGBEnabled: false},
		NumShards:                1,
		DiskSizeGB:               100,
	}

	switch details.PlanID {
	case awsDev:
		provider.ProviderName = "AWS"
		provider.RegionName = "US_EAST_1"
		provider.InstanceSizeName = "M10"
		provider.DiskIOPS = 300
		provider.EncryptEBSVolume = false
	case gcpDev:
		//TODO
	default:
		// TODO
		return returnObject, fmt.Errorf("life sucks")
	}

	provisionReq.ProviderSettings = provider
	json, err := json.Marshal(provisionReq)
	if err != nil {
		log.Printf("Error - Provision - Failed Marshal. JSON: %+v, Err: %+v", json, err)
		return brokerapi.ProvisionedServiceSpec{}, err
	}

	_, err = NewCluster(json)
	if err != nil {
		log.Printf("Error - Provision - Failed NewCluster. Err: %+v", err)
		return returnObject, err
	}

	returnObject.IsAsync = true
	returnObject.OperationData = OperationProvision
	returnObject.DashboardURL = Host + "/v2/" + Group + "#clusters/detail/" + instanceID32
	return returnObject, err
}

// DELETE localhost:8080/v2/service_instances/instanceID32?service_id=atlas&plan_id=aws_dev

// Deprovision - MongoDB Atlas Broker
func (a AtlasBroker) Deprovision(ctx context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	returnObject := brokerapi.DeprovisionServiceSpec{}
	instanceID32 := strings.Replace(instanceID, "-", "", -1)
	_, err := TerminateCluster(instanceID32)
	if err != nil {
		log.Printf("Error - Deprovision - Failed TerminateCluster. Err: %+v", err)
		return returnObject, err
	}
	returnObject.IsAsync = true
	returnObject.OperationData = OperationDeprovision
	return returnObject, err
}

// GET localhost:8080/v2/service_instances/abc123abc123abc123/last_operation?service_id=atlas&plan_id=aws_dev&operation=provision
// GET localhost:8080/v2/service_instances/abc123abc123abc123/last_operation?service_id=atlas&plan_id=aws_dev&operation=deprovision

// LastOperation - MongoDB Atlas Broker
func (a AtlasBroker) LastOperation(ctx context.Context, instanceID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error) {

	instanceID32 := strings.Replace(instanceID, "-", "", -1)
	returnObject := brokerapi.LastOperation{}
	response, err := GetCluster(instanceID32)
	if err != nil {
		log.Printf("Error - LastOperation - Failed GetCluster. Response: %+v, Err: %+v", response, err)
		return returnObject, err
	}

	lastOperationState := brokerapi.LastOperationState(brokerapi.Failed)
	switch details.OperationData {
	case OperationDeprovision:
		if response.StateName == StateDELETED || response.ErrorCode == ErrorCode404 {
			// TODO ErrorCode404 in OperationDeprovision should really return a 410 Gone response
			lastOperationState = brokerapi.Succeeded
		} else if response.StateName == StateDELETING {
			lastOperationState = brokerapi.InProgress
		} else {
			lastOperationState = brokerapi.Failed
		}
	case OperationProvision:
		switch response.StateName {
		case StateIDLE:
			lastOperationState = brokerapi.Succeeded
		case StateCREATING:
			lastOperationState = brokerapi.InProgress
		default:
			lastOperationState = brokerapi.Failed
		}
	default:
		log.Printf("LastOperation OperationData Unknown %+v", details.OperationData)
		lastOperationState = brokerapi.Failed
	}

	returnObject.State = lastOperationState
	returnObject.Description = "Atlas Cluster State is " + response.StateName + response.ErrorCode
	log.Printf("\nLastOperation (details): %+v\nLastOperation (response): %+v\nLastOperation (returnObject): %+v\n", details, response, returnObject)
	return returnObject, nil
}

// Update - MongoDB Atlas Broker --- TODO
func (a AtlasBroker) Update(ctx context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {
	// instanceID32 := strings.Replace(instanceID, "-", "", -1)
	returnObject := brokerapi.UpdateServiceSpec{}
	return returnObject, nil
}

// PUT localhost:8080/v2/service_instances/abc123abc123abc123/service_bindings/bbbbbbbb

// Bind - MongoDB Atlas Broker - Creates a readWriteAnyDatabse user in the Atlas project
func (a AtlasBroker) Bind(ctx context.Context, instanceID, bindingID string, details brokerapi.BindDetails, asyncAllowed bool) (brokerapi.Binding, error) {

	instanceID32 := strings.Replace(instanceID, "-", "", -1)
	bindingID32 := strings.Replace(bindingID, "-", "", -1)
	returnObject := brokerapi.Binding{}
	request := BindRequest{}

	request.DatabaseName = UserDatabaseStore
	request.GroupID = Group
	role := Roles{DatabaseName: UserRoleDatabase, RoleName: UserRoleName}
	request.Roles = []Roles{role}

	// TODO - Generate and fetch from CredHub
	credhubPass, err := credhub.GenPassFromCredhub(instanceID32, bindingID32)
	if err != nil {
		log.Printf("Error - Bind - Failed genPassFromCredhub. Err: %+v", err)
		return returnObject, err
	}
	request.Username = bindingID32
	request.Password = string(credhubPass.Value)

	log.Printf("\nBind (request): %+v", request)

	json, err := json.Marshal(request)
	if err != nil {
		log.Printf("Error - Bind - Failed Marshal. JSON: %+v, Err: %+v", json, err)
		return returnObject, err
	}
	response, err := NewUser(json) // BindResponse
	if err != nil {
		log.Printf("Error - Bind - Failed NewUser. Response: %+v, Err: %+v", response, err)
		return returnObject, err
	}

	// fill out response
	returnObject.IsAsync = true
	returnObject.OperationData = OperationBind ///TODO handle in LastOperation

	// TODO
	// TODO - per spec, shouldn't return this when async and 202 Accepted is returned, but oh well! (for now)
	// returnObject.Credentials = {
	// 	user: response.Username,
	// 	pass: request.Password,
	// 	uri: ""
	// }

	return returnObject, nil
}

// GetBinding - MongoDB Atlas Broker -- TODO
func (a AtlasBroker) GetBinding(ctx context.Context, instanceID, bindingID string) (brokerapi.GetBindingSpec, error) {
	instanceID32 := strings.Replace(instanceID, "-", "", -1)
	bindingID32 := strings.Replace(bindingID, "-", "", -1)
	returnObject := brokerapi.GetBindingSpec{}

	credhubPass, err := credhub.GetPassFromCredhub(instanceID32, bindingID32)
	if err != nil {
		log.Printf("Error - Bind - Failed genPassFromCredhub. Err: %+v", err)
		return returnObject, err
	}

	returnObject.Credentials = atlasCredentials{
		username: bindingID32,
		password: string(credhubPass.Value),
		// diana for you ! url:
	}

	return returnObject, nil
}

// GET localhost:8080/v2/service_instances/abc123abc123abc123/service_bindings/bbbbbbbb/last_operation?service_id=atlas&plan_id=aws_dev&operation=bind

// LastBindingOperation - MongoDB Atlas Broker
func (a AtlasBroker) LastBindingOperation(ctx context.Context, instanceID, bindingID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error) {

	// no state associated with a user, so as long as we get the user back it succeeded.
	// do need to check for weird err msgs...
	// is this call sync? TODO Check

	instanceID32 := strings.Replace(instanceID, "-", "", -1)
	bindingID32 := strings.Replace(bindingID, "-", "", -1)
	lastOperationState := brokerapi.LastOperationState(brokerapi.Failed)
	returnObject := brokerapi.LastOperation{}

	response, err := GetUser(instanceID32, bindingID32)
	if err != nil {
		log.Printf("Error - LastBindingOperation - Failed GetUser. Response: %+v, Err: %+v", response, err)
		// TODO -- take a closer looks at the error messages to see if the user is in progress maybe
	} else {
		switch details.OperationData {
		case OperationBind:
			if response.Username == bindingID32 {
				lastOperationState = brokerapi.Succeeded
			} else {
				lastOperationState = brokerapi.Failed // how to tell if failed?
			}
		case OperationUnbind:
			if response.Username == bindingID32 {
				lastOperationState = brokerapi.InProgress
			} else {
				lastOperationState = brokerapi.Succeeded
			}
		default:
			log.Printf("LastBindingOperation OperationData Unknown %+v", details.OperationData)
			// lastOperationState = brokerapi.Failed
		}
	}

	returnObject.State = lastOperationState
	returnObject.Description = "Atlas Responded with User: " + response.Username
	log.Printf("\nLastBindingOperation (details): %+v\nLastBindingOperation (response): %+v\nLastBindingOperation (returnObject): %+v\n", details, response, returnObject)
	return returnObject, nil
}

// Unbind - MongoDB Atlas Broker
func (a AtlasBroker) Unbind(ctx context.Context, instanceID, bindingID string, details brokerapi.UnbindDetails, asyncAllowed bool) (brokerapi.UnbindSpec, error) {

	returnObject := brokerapi.UnbindSpec{}
	instanceID32 := strings.Replace(instanceID, "-", "", -1)
	bindingID32 := strings.Replace(bindingID, "-", "", -1)

	_, err := DeleteUser(instanceID32, bindingID32)
	if err != nil {
		log.Printf("Error - Unbind - Failed DeleteUser. Err: %+v", err)
		return returnObject, err
	}

	err = credhub.DeletePassFromCredhub(instanceID32, bindingID32)
	if err != nil {
		log.Printf("Error - Unbind - Failed DeletePassFromCredhub to delete user from credhub. Err: %+v", err)
		return returnObject, err
	}

	returnObject.IsAsync = true
	returnObject.OperationData = OperationUnbind
	return returnObject, nil

}

// func (atlasServiceBroker *AtlasServiceBroker) Bind(instanceID, bindingID string, details
// 			credentialsMap := map[string]interface{}{
// 				"host":     instanceCredentials.Host,
// 				"port":     instanceCredentials.Port,
// 				"password": instanceCredentials.Password,
// 			binding.Credentials = credentialsMap
// 			return binding, nil
