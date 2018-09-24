package main

import (
	"fmt"
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/desteves/mongodb-atlas-service-broker/broker"
	"github.com/pivotal-cf/brokerapi"
)

func main() {

	logger := lager.NewLogger("broker")
	atlas := broker.AtlasBroker{}

	username := os.Getenv("SECURITY_USER_NAME")
	password := os.Getenv("SECURITY_USER_PASSWORD")

	if len(username) == 0 {
		logger.Fatal("SECURITY_USER_NAME is not set!", fmt.Errorf("SECURITY_USER_NAME is not set!"))
	}

	if len(password) == 0 {
		logger.Fatal("SECURITY_USER_PASSWORD is not set!", fmt.Errorf("SECURITY_USER_PASSWORD is not set!"))
	}
	brokerCreds := brokerapi.BrokerCredentials{
		Username: username,
		Password: password,
	}
	handler := brokerapi.New(atlas, logger, brokerCreds)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		logger.Fatal("PORT is not set!", fmt.Errorf("PORT is not set!"))
	}
	err := http.ListenAndServe(":"+port, handler)
	logger.Fatal("HTTP Service Failed", err, lager.Data{})
}
