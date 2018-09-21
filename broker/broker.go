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
	_ = ProvisionResponse{}

	if details.PlanID == custom {
		//Just send it along.....
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
		//TODO - test this!!!
		provisionReq.ProviderSettings.ProviderName = "GCP"
		provider.RegionName = "EASTERN_US"
		provider.InstanceSizeName = "M10"
		provider.DiskIOPS = 100
	default:
		// error
		// TODO
		return returnObject, fmt.Errorf("life sucks")
	}

	provisionReq.ProviderSettings = provider
	json, err := json.Marshal(provisionReq)
	if err != nil {
		log.Panic(err)
	}

	_, err = NewCluster(json)
	if err != nil {
		returnObject.IsAsync = true
		returnObject.OperationData = ""
		returnObject.DashboardURL = "https://cloud.mongodb.com/api/atlas/v1.0/groups/" + Group + "/clusters/" + instanceID32
	}

	return returnObject, err
}

// LastOperation - MongoDB Atlas Broker
// If the broker provisions asynchronously, the Cloud Controller will poll this endpoint
// for the status of the provisioning operation.
func (a AtlasBroker) LastOperation(ctx context.Context, instanceID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error) {
	instanceID32 := strings.Replace(instanceID, "-", "", -1)
	returnObject := brokerapi.LastOperation{}
	response := LastOperationResponse{}
	response, err := GetCluster(instanceID32)
	lastOperationState := brokerapi.Failed
	if err != nil {
		returnObject.Description = response.StateName
		switch response.StateName {
		case StateIDLE:
			lastOperationState = brokerapi.Succeeded
		default:
			lastOperationState = brokerapi.InProgress
		}
	}
	returnObject.State = lastOperationState
	return returnObject, err
}

// Bind - MongoDB Atlas Broker
func (a AtlasBroker) Bind(ctx context.Context, instanceID, bindingID string, details brokerapi.BindDetails, asyncAllowed bool) (brokerapi.Binding, error) {

	instanceID32 := strings.Replace(instanceID, "-", "", -1)

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

// Deprovision - MongoDB Atlas Broker -- TODO
func (a AtlasBroker) Deprovision(ctx context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {

	instanceID32 := strings.Replace(instanceID, "-", "", -1)

	return brokerapi.DeprovisionServiceSpec{}, nil
}

// func (atlasServiceBroker *AtlasServiceBroker) Bind(instanceID, bindingID string, details
// 			credentialsMap := map[string]interface{}{
// 				"host":     instanceCredentials.Host,
// 				"port":     instanceCredentials.Port,
// 				"password": instanceCredentials.Password,
// 			binding.Credentials = credentialsMap
// 			return binding, nil
