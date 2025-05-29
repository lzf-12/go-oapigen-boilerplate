### go-oapigen-boilerplate

this repository is a boilerplate for open api schema first, to generate basic rest api using oapi-codegen library.


### how to use

1. define open api spec in /docs directory

2. generate oapi-codegen config file with package name
```  make generate-config package=billing```

3. generate server and request/response struct based on oapi specification
``` make generate ```

4. code implementation & wire the routes


### notes
1. for request/response validation need to add middleware based on this:
https://github.com/oapi-codegen/oapi-codegen?tab=readme-ov-file#requestresponse-validation-middleware