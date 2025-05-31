package spec_validator

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
	validatorError "github.com/pb33f/libopenapi-validator/errors"
	"gopkg.in/yaml.v2"
)

type SpecValidator struct {
	Validator   validator.Validator
	Name        string
	BasePath    string
	Version     string
	Description string
}

type Config struct {
	Validation ValidationConfig `yaml:"validation"`
	Specs      []SpecConfig     `yaml:"specs"`
}

type ValidationConfig struct {
	Enabled           bool            `yaml:"enabled"`
	ValidateResponses bool            `yaml:"validate_responses"`
	SkipPaths         map[string]bool `yaml:"skip_paths"`
	// Specs             []SpecConfig    `yaml:"specs"`
}

type SpecConfig struct {
	Name        string `yaml:"name" json:"name"`
	FilePath    string `yaml:"file_path" json:"file_path"`
	BasePath    string `yaml:"base_path" json:"base_path"`
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	Description string `yaml:"description" json:"description"`
}

type MultiSpecValidator struct {
	Config       Config
	validators   map[string]*SpecValidator
	routeMapping map[string]string
	fallbackSpec string
	mu           sync.RWMutex
}

func NewMultiSpecValidator(configFile []byte) *MultiSpecValidator {

	var cfg Config
	err := yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		log.Fatal("failed to parse validation configuration yaml")
	}

	return &MultiSpecValidator{
		Config:       cfg,
		validators:   make(map[string]*SpecValidator),
		routeMapping: make(map[string]string),
	}
}

func (msv *MultiSpecValidator) LoadSpecsFromDirectory(directory string) error {
	files, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", directory, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
			specPath := filepath.Join(directory, file.Name())
			specName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

			if err := msv.LoadSpec(specName, specPath, ""); err != nil {
				return fmt.Errorf("failed to load spec %s: %w", specName, err)
			}
		}
	}

	return nil
}

func (msv *MultiSpecValidator) LoadSpec(name, filePath, basePath string) error {
	msv.mu.Lock()
	defer msv.mu.Unlock()

	specData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read spec file: %w", err)
	}

	// parse the OpenAPI document
	doc, err := libopenapi.NewDocument(specData)
	if err != nil {
		return fmt.Errorf("failed to parse OpenAPI document: %w", err)
	}

	// build the document model
	model, errs := doc.BuildV3Model()
	if len(errs) > 0 {
		return fmt.Errorf("failed to build document model: %v", errs)
	}

	v, errs := validator.NewValidator(doc, nil)
	if len(errs) > 0 {

		// log each error when failed to create validator
		for _, e := range errs {
			fmt.Printf("error when creating validator, err: %v", e)
		}
		return errors.New("failed to create validator")
	}

	// extract info from the spec
	info := model.Model.Info
	description := ""
	if info != nil && info.Description != "" {
		description = info.Description
	}

	// store the validator
	msv.validators[name] = &SpecValidator{
		Validator:   v,
		Name:        name,
		BasePath:    basePath,
		Description: description,
	}

	// auto-map routes based on basePath if provided
	if basePath != "" {
		msv.routeMapping[basePath] = name
	}

	return nil
}

func (msv *MultiSpecValidator) LoadValidationSpecsFromConfigFile() error {

	if len(msv.Config.Specs) == 0 {
		return fmt.Errorf("specs is not configured in spec config file")
	}

	specs := msv.Config.Specs

	// load specs config
	for _, spec := range specs {
		if !spec.Enabled {
			continue
		}

		if err := msv.LoadSpec(spec.Name, spec.FilePath, spec.BasePath); err != nil {
			return fmt.Errorf("failed to load spec %s: %w", spec.Name, err)
		}
	}

	return nil
}

func (msv *MultiSpecValidator) AddRouteMapping(routePattern, specName string) {
	msv.mu.Lock()
	defer msv.mu.Unlock()
	msv.routeMapping[routePattern] = specName
}

func (msv *MultiSpecValidator) GetValidatorForRequest(r *http.Request) (*SpecValidator, error) {
	msv.mu.RLock()
	defer msv.mu.RUnlock()

	path := r.URL.Path

	// try exact match first
	for route, specName := range msv.routeMapping {
		if strings.HasPrefix(path, route) {
			if validator, exists := msv.validators[specName]; exists {
				return validator, nil
			}
		}
	}
	if msv.fallbackSpec != "" {
		if validator, exists := msv.validators[msv.fallbackSpec]; exists {
			return validator, nil
		}
	}

	return nil, fmt.Errorf("no suitable validator found for path: %s", path)
}

func (msv *MultiSpecValidator) ValidateRequest(r *http.Request) ([]*validatorError.ValidationError, error) {

	specValidator, err := msv.GetValidatorForRequest(r)
	if err != nil {
		return nil, err
	}

	valid, errs := specValidator.Validator.ValidateHttpRequest(r)
	if !valid && len(errs) > 0 {
		return errs, nil
	}

	return nil, nil
}

func (msv *MultiSpecValidator) ValidateResponse(r *http.Request, resp *http.Response) ([]*validatorError.ValidationError, error) {
	specValidator, err := msv.GetValidatorForRequest(r)
	if err != nil {
		return nil, err
	}

	valid, errs := specValidator.Validator.ValidateHttpResponse(r, resp)
	if !valid && len(errs) > 0 {
		return errs, nil
	}

	return nil, nil
}

// returns information about all loaded validators
func (msv *MultiSpecValidator) ListValidators() map[string]*SpecValidator {
	msv.mu.RLock()
	defer msv.mu.RUnlock()

	result := make(map[string]*SpecValidator)
	for name, validator := range msv.validators {
		result[name] = validator
	}
	return result
}

// could be used to add exception
func (msv *MultiSpecValidator) RemoveValidator(name string) {
	msv.mu.Lock()
	defer msv.mu.Unlock()

	delete(msv.validators, name)

	// remove associated route mappings
	for route, specName := range msv.routeMapping {
		if specName == name {
			delete(msv.routeMapping, route)
		}
	}
}
