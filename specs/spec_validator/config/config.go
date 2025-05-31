package spec_validator

import (
	_ "embed"
)

//go:embed dev.config.yaml
var SpecValidationConfigDevFile []byte

//go:embed prod.config.yaml
var SpecValidationConfigProdFile []byte
