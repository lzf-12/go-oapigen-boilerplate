# go-oapigen-boilerplate

This repository provides a boilerplate for OpenAPI schema-first development. It uses the `oapi-codegen` library to generate a basic REST API from OpenAPI specifications.

---

## How to Use

#### 1. Define OpenAPI Specification

Place your OpenAPI specs in the `/specs/api` directory. Reusable common definitions should go in `/specs/api/common`.

#### 2. (Optional) Generate Common Configs

If you have reusable specs like responses (as referenced in specs like `auth.yaml`), generate them with:
```
make common-config
```

#### 3. Generate oapi-codegen Config File

Create a config file with the desired package name and its specs location:

```
make package-config name=billing specpath=specs/api/v1
```

#### 4. Generate Server Code and Structs

Generate the server and request/response structs from the OpenAPI spec:
```
make generate
```

#### 5. Implement and Wire Routes

Proceed to implement your business logic and connect the routes.


#### 6. (Optional) Define Spec Validator Config

Place the validator `.yaml` config in `specs/spec_validator/config`. Ensure that it includes all relevant packages based on your OpenAPI definitions.


#### 7. (Optional) Add Middleware for Validation

You can create custom middleware to validate requests based on the defined OpenAPI specs, you can see the example of middleware validation on `pkg/middleware/specvalidator_middleware` the example use `https://github.com/pb33f/libopenapi-validaton` for its validation rule and logic.



## Running Locally

#### Set Up Environment Variables

```
cp .env.example .env
```

#### Initialize Default Sqlite Database

```
make default-db
```


#### Start the application

```
go run ./cmd
```

## Build and run with docker


``` 
docker build -t go-oasfirst-sqlite . 
docker run --env-file .env go-oasfirst-sqlite
```