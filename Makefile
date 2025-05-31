.PHONY: generate run

install:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

generate-config:

	@test -n "$(package)" || (echo "package= parameter is required"; exit 1)
	@test -n "$(specpath)" || (echo "specpath= parameter is required to specify spec .yaml location to the codegen"; exit 1)

	# validate oas spec existence
	@test -f "$(specpath)/$(package).yaml" || (echo "error: open api spec file /$(specpath)/$(package).yaml does not exist"; exit 1)

	# generate or refresh oapi config file
	mkdir -p api/$(package)
	PACKAGE=$(package);\
	rm -f api/$(package)//server.cfg.yaml;\
	[ -f api/$(package)/server.cfg.yaml ] || echo "package: $(package)\ngenerate:\n  gin-server: true\n  strict-server: true\n  embedded-spec: true\noutput: $(package)-server.gen.go" > api/$(package)/server.cfg.yaml;\
	rm -f api/$(package)//types.cfg.yaml;\
	[ -f api/$(package)/types.cfg.yaml ] || echo "package: $(package)\ngenerate:\n  models: true\noutput: $(package)-types.gen.go" > api/$(package)/types.cfg.yaml;\
	rm -f api/$(package)//gen.go;\
	[ -f api/$(package)/gen.go ] || echo "//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=types.cfg.yaml ../../$(specpath)/$(package).yaml\n//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=server.cfg.yaml ../../$(specpath)/$(package).yaml\n\npackage $(package)\n" > api/$(package)/gen.go


generate-handler:
	@echo "generating code from OpenAPI spec..."
	go generate ./...

run:
	@echo "Starting server..."
	go run cmd/main.go