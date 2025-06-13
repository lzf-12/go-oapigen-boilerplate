package main

import (
	"log"
	"oapi-to-rest/api"
	"oapi-to-rest/pkg/env"
	"oapi-to-rest/pkg/middleware"
	"oapi-to-rest/specs/spec_validator"
	svcfg "oapi-to-rest/specs/spec_validator/config"
)

func main() {

	// load config
	config, err := env.LoadConfig(".env")

	// init server
	server := api.NewServer(config)

	// init spec validator to validate based on spec definition
	var svCfg []byte
	svCfg = svcfg.SpecValidationConfigDevFile
	if config.Env == env.Production.String() {
		svCfg = svcfg.SpecValidationConfigProdFile
	}
	msv := spec_validator.NewMultiSpecValidator(svCfg)

	err = msv.LoadValidationSpecsFromConfigFile()
	if err != nil {
		log.Fatalf("error: %s", err.Error())
		return
	}

	// apply spec validation middleware
	server.Router.Use(middleware.RequestValidationMiddleware(msv))
	server.Router.Use(middleware.ResponseValidationMiddleware(msv))

	server.RegisterRoutes()
	log.Fatal(server.Start(":8080"))
}
