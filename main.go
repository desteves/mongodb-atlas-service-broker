package main

import (
	"net/http"

	"code.cloudfoundry.org/lager"
	"github.com/desteves/mongodb-atlas-service-broker/broker"
	"github.com/pivotal-cf/brokerapi"
)

func main() {
	brokerCreds := brokerapi.BrokerCredentials{
		Username: "admin",
		Password: "admin",
	}

	logger := lager.NewLogger("broker")
	atlas := broker.AtlasBroker{}
	handler := brokerapi.New(atlas, logger, brokerCreds)

	err := http.ListenAndServe(":8080", handler)
	logger.Fatal("HTTP Service Failed", err, lager.Data{})

	//create broker implementation (broker package)
	//pass broker implementation to brokerapi.new(impl)
	//should revieve back a http handler
	//server handler as webservice

	// r := gin.Default()
	// r.GET("/v2/catalog", broker.catalog())
	// })
	// r.Run() // listen and serve on 0.0.0.0:8080
}
