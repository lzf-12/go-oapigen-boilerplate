// Package user provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package user

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	externalRef0 "oapi-to-rest/api/common"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime"
	strictgin "github.com/oapi-codegen/runtime/strictmiddleware/gin"
)

const (
	BearerAuthScopes = "bearerAuth.Scopes"
)

// PaginatedUserResponse defines model for PaginatedUserResponse.
type PaginatedUserResponse struct {
	Data *[]User `json:"data,omitempty"`

	// Filters show applied filters parameter
	Filters    *interface{} `json:"filters,omitempty"`
	Pagination *struct {
		CurrentPage *int64 `json:"currentPage,omitempty"`
		PageSize    *int64 `json:"pageSize,omitempty"`
		TotalItems  *int64 `json:"totalItems,omitempty"`
		TotalPages  *int64 `json:"totalPages,omitempty"`
	} `json:"pagination,omitempty"`
}

// User defines model for User.
type User struct {
	Email     *string `db:"email" json:"email,omitempty"`
	FirstName *string `db:"first_name" json:"first_name,omitempty"`
	Id        *int64  `db:"id" json:"id,omitempty"`
	IsActive  *int32  `db:"is_active" json:"is_active,omitempty"`
	LastName  *string `db:"last_name" json:"last_name,omitempty"`
}

// GetUserParams defines parameters for GetUser.
type GetUserParams struct {
	// Email email to filter by
	Email *string `form:"email,omitempty" json:"email,omitempty"`

	// IsActive is_active to filter by
	IsActive *string `form:"is_active,omitempty" json:"is_active,omitempty"`

	// Page filter by page
	Page *int64 `form:"page,omitempty" json:"page,omitempty"`

	// PageSize size of each page
	PageSize *int64 `form:"pageSize,omitempty" json:"pageSize,omitempty"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get users with filters
	// (GET /user)
	GetUser(c *gin.Context, params GetUserParams)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandler       func(*gin.Context, error, int)
}

type MiddlewareFunc func(c *gin.Context)

// GetUser operation middleware
func (siw *ServerInterfaceWrapper) GetUser(c *gin.Context) {

	var err error

	c.Set(BearerAuthScopes, []string{})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetUserParams

	// ------------- Optional query parameter "email" -------------

	err = runtime.BindQueryParameter("form", true, false, "email", c.Request.URL.Query(), &params.Email)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter email: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "is_active" -------------

	err = runtime.BindQueryParameter("form", true, false, "is_active", c.Request.URL.Query(), &params.IsActive)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter is_active: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "page" -------------

	err = runtime.BindQueryParameter("form", true, false, "page", c.Request.URL.Query(), &params.Page)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter page: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "pageSize" -------------

	err = runtime.BindQueryParameter("form", true, false, "pageSize", c.Request.URL.Query(), &params.PageSize)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter pageSize: %w", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.GetUser(c, params)
}

// GinServerOptions provides options for the Gin server.
type GinServerOptions struct {
	BaseURL      string
	Middlewares  []MiddlewareFunc
	ErrorHandler func(*gin.Context, error, int)
}

// RegisterHandlers creates http.Handler with routing matching OpenAPI spec.
func RegisterHandlers(router gin.IRouter, si ServerInterface) {
	RegisterHandlersWithOptions(router, si, GinServerOptions{})
}

// RegisterHandlersWithOptions creates http.Handler with additional options
func RegisterHandlersWithOptions(router gin.IRouter, si ServerInterface, options GinServerOptions) {
	errorHandler := options.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(c *gin.Context, err error, statusCode int) {
			c.JSON(statusCode, gin.H{"msg": err.Error()})
		}
	}

	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandler:       errorHandler,
	}

	router.GET(options.BaseURL+"/user", wrapper.GetUser)
}

type GetUserRequestObject struct {
	Params GetUserParams
}

type GetUserResponseObject interface {
	VisitGetUserResponse(w http.ResponseWriter) error
}

type GetUser200JSONResponse PaginatedUserResponse

func (response GetUser200JSONResponse) VisitGetUserResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetUser400JSONResponse externalRef0.StandardErrorResponse

func (response GetUser400JSONResponse) VisitGetUserResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)

	return json.NewEncoder(w).Encode(response)
}

type GetUser500JSONResponse externalRef0.StandardErrorResponse

func (response GetUser500JSONResponse) VisitGetUserResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {
	// Get users with filters
	// (GET /user)
	GetUser(ctx context.Context, request GetUserRequestObject) (GetUserResponseObject, error)
}

type StrictHandlerFunc = strictgin.StrictGinHandlerFunc
type StrictMiddlewareFunc = strictgin.StrictGinMiddlewareFunc

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
}

// GetUser operation middleware
func (sh *strictHandler) GetUser(ctx *gin.Context, params GetUserParams) {
	var request GetUserRequestObject

	request.Params = params

	handler := func(ctx *gin.Context, request interface{}) (interface{}, error) {
		return sh.ssi.GetUser(ctx, request.(GetUserRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetUser")
	}

	response, err := handler(ctx, request)

	if err != nil {
		ctx.Error(err)
		ctx.Status(http.StatusInternalServerError)
	} else if validResponse, ok := response.(GetUserResponseObject); ok {
		if err := validResponse.VisitGetUserResponse(ctx.Writer); err != nil {
			ctx.Error(err)
		}
	} else if response != nil {
		ctx.Error(fmt.Errorf("unexpected response type: %T", response))
	}
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/8xWTW/zNgz+Kwa3o9OkH9vBtwL7QHcK1g07FEXA2IyjwpZUik6bFv7vgyQnThO389r3",
	"8J5sS+JDUuTz0K+Qm9oaTVocZK/g8jXVGF7nWCqNQsXfjvhPctZoR37DsrHEoigcK1DQP5VQHRZ+ZFpB",
	"Bj9Me+RpBzv1UNCmIFtLkAEy49Z/r1QlxBGPXM7KijIaMnBr85SgtZWiIulOJRYZaxJi2COZ5QPlAik8",
	"T0oz6RaVFuIV5vTaeic2JhSAj7PIG2bSMscypLgyXKNEiJ+vejcesYwpWCzpVr2MPS5GsLrZ3dFYAx/P",
	"OIP2+CoGVlII93+SPNWoKv/SnXfCSpfhMg1aNclNQSXpCT0L40SwjIVaQtaZtqGE7GShsabPIR3YezhV",
	"jEl7DLAqIqBbYC5qc1Kwy4tP4u4BPXyFX8q+N28HC+cs5W6BVi1yU9dGL7gj5OJWUBfIxa/M5iOakqCq",
	"3Gl0bQrkTcMWFoXyBMFqfmg9FJHSTlDnNIjoBKU5dHbY2KomJ1jbQUtRUg1jCmNOi9gXp5th4WRjkBWO",
	"8oaVbG+9KMXbWRIy8XUj6/7rt12P/PHPX5BGZfRIcbfvmbWIjVVTemVCEDGHwLbkF1Oj0sn1/AZS2BC7",
	"KGznZ7OzmY/cWNJoFWRwGZa8ssg6RDVtOrqWJP7hKxL066aADH4nCXT2Bp0eOsjujhU0MDQR06lnstyC",
	"jxQyeGyI/Ufs2o7KXaL4Tlm3ITFPH2jTY1d7Roxx19PnKy73XhKvx++46rZ6LwWtsKkEsvN0jLKeDCX1",
	"QolZJYT5+r/8hhEx7Hs2xvl9Cjuqh564mM3CvDJaSIeuCNMxD30xfXBxuPXuPprGwxM+tPLbhK+TSjnx",
	"KfuOdL4uV98wjv8rbgMRLrFImB4bckGcfvq+ogs/IhqrJCjtGxEKjD2Un7t7X3PX1DXyNtI83nrypGS9",
	"+weKXhzxZsf6hivIYIpWTTfn0N63/wYAAP//9vUGXdsJAAA=",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	for rawPath, rawFunc := range externalRef0.PathToRawSpec(path.Join(path.Dir(pathToFile), "../common/response.yaml")) {
		if _, ok := res[rawPath]; ok {
			// it is not possible to compare functions in golang, so always overwrite the old value
		}
		res[rawPath] = rawFunc
	}
	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
