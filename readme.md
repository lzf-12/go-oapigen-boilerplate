### go-oapigen-boilerplate

this repository is a boilerplate for open api schema first, generate basic rest api using oapi-codegen library.


### how to use

1. define open api spec in /specs directory

2. generate oapi-codegen config file with package name
```make generate-config package=billing specpath=specs/api/v1```

3. generate server and request/response struct based on oapi specification
```make generate-handler```

4. code implementation & wire the routes

5. (optional) define spec_validator configuration

6. (optional) create middleware and assign middleware to validate based on defined open api spec