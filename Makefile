.PHONY: generate run

install:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest


.PHONY: generate-common generate-common-config

default-db:
	@echo "build scripts for default sqlite with predefined tables..."
	go build -o db-init ./scripts/default_sqlite.go
	@echo "running scripts..."
	./db-init
	@echo "done"

common-config:

	@test -n "$(specpath)" || (echo "specpath= parameter is required"; exit 1)
	mkdir -p api/common
	rm -f api/common/cfg.yaml;\
	[ -f api/common/cfg.yaml ] || echo "package: common\noutput: common.gen.go\ngenerate:\n  models: true\n  embedded-spec: true\noutput-options:\n  skip-prune: true" > api/common/cfg.yaml;\
	rm -f api/common/gen.go;\
	[ -f api/common/gen.go ] || echo "//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=cfg.yaml ../../$(specpath)/response.yaml\n\npackage common\n" > api/common/gen.go

package-config:

	@test -n "$(name)" || (echo "name= parameter is required"; exit 1)
	@test -n "$(specpath)" || (echo "specpath= parameter is required to specify spec .yaml location to the codegen"; exit 1)

	# validate oas spec existence
	@test -f "$(specpath)/$(name).yaml" || (echo "error: open api spec file /$(specpath)/$(name).yaml does not exist"; exit 1)
	
	# generate or refresh oapi config file
	mkdir -p api/$(name)
	PACKAGE=$(name);\
	rm -f api/$(name)//cfg.yaml;\
	[ -f api/$(name)/cfg.yaml ] || echo "package: $(name)\ngenerate:\n  gin-server: true\n  strict-server: true\n  embedded-spec: true\n  models: true\noutput: $(name).gen.go\nimport-mapping:\n  specs/api/common/response.yaml: \"oapi-to-rest/api/common\"" > api/$(name)/cfg.yaml;\
	rm -f api/$(name)//gen.go;\
	[ -f api/$(name)/gen.go ] || echo "//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=cfg.yaml ../../$(specpath)/$(name).yaml\n\npackage $(name)\n" > api/$(name)/gen.go


generate:
	@echo "generating code from OpenAPI spec..."
	go generate ./...

run:
	@echo "Starting server..."
	go run cmd/main.go