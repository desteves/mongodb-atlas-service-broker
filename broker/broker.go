package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	returnObject.DashboardURL = Host + "/v2/groups/" + Group + "#clusters/detail/" + instanceID32
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

// Bind - MongoDB Atlas Broker
func (a AtlasBroker) Bind(ctx context.Context, instanceID, bindingID string, details brokerapi.BindDetails, asyncAllowed bool) (brokerapi.Binding, error) {

	// TODO
	// instanceID32 := strings.Replace(instanceID, "-", "", -1)

	b := brokerapi.Binding{}

	return b, nil
}

// GetBinding - MongoDB Atlas Broker
func (a AtlasBroker) GetBinding(ctx context.Context, instanceID, bindingID string) (brokerapi.GetBindingSpec, error) {
	return brokerapi.GetBindingSpec{}, nil
}

// LastBindingOperation - MongoDB Atlas Broker
func (a AtlasBroker) LastBindingOperation(ctx context.Context, instanceID, bindingID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error) {
	return brokerapi.LastOperation{}, nil
}

// Unbind - MongoDB Atlas Broker ---- TODO
func (a AtlasBroker) Unbind(ctx context.Context, instanceID, bindingID string, details brokerapi.UnbindDetails, asyncAllowed bool) (brokerapi.UnbindSpec, error) {
	return brokerapi.UnbindSpec{}, nil
}

// Update - MongoDB Atlas Broker --- TODO
func (a AtlasBroker) Update(ctx context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {
	return brokerapi.UpdateServiceSpec{}, nil
}

// func (atlasServiceBroker *AtlasServiceBroker) Bind(instanceID, bindingID string, details
// 			credentialsMap := map[string]interface{}{
// 				"host":     instanceCredentials.Host,
// 				"port":     instanceCredentials.Port,
// 				"password": instanceCredentials.Password,
// 			binding.Credentials = credentialsMap
// 			return binding, nil
