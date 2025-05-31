package main

import (
	"log"
	"oapi-to-rest/api"
	"oapi-to-rest/specs/spec_validator"
	svcfg "oapi-to-rest/specs/spec_validator/config"
)

func main() {

	server := api.NewServer()

	// init spec validator to validate based on spec definition
	svCfg := svcfg.SpecValidationConfigDevFile // change based on env
	msv := spec_validator.NewMultiSpecValidator(svCfg)

	err := msv.LoadValidationSpecsFromConfigFile()
	if err != nil {
		log.Fatalf("error: %s", err.Error())
		return
	}

	// apply spec validation middleware
	server.Router.Use(api.RequestValidationMiddleware(msv))
	server.Router.Use(api.ResponseValidationMiddleware(msv))

	server.RegisterRoutes()
	log.Fatal(server.Start(":8080"))
}
