package api

// ListOperation lists the entities
//
//openapi:operation /entities GET
//openapi:tag tag1
//openapi:tag tag2
func ListOperation() {}

// GetOperation gets a specific entity
//
//openapi:operation /entities/{id} GET
//openapi:parameter id path string "the id of the entity"
//openapi:response default "this is a description"
//openapi:responseContent default application/json Model
//openapi:responseHeader default My-Custom-Header string "this header will tell you..."
//openapi:responseHeader default My-Other-Custom-Header string/date-time "this header will tell you..."
//openapi:responseExample default application/ld+json examples/get_operation_default.json
//openapi:response 400 "client did something wrong"
//openapi:responseContent 400 application/problem+json Problem
//openapi:responseExample 400 application/problem+json examples/get_operation_error.json
//openapi:response 404 "something was not found"
//openapi:responseContent 404 application/problem+json Problem
func GetOperation() {}

// ReplaceOperation will replace (or create) a specific entity
//
//openapi:operation /entities/{id} PUT
//openapi:parameter id path string "the id of the entity"
//openapi:requestBody application/json Model "The data to replace the current entity - if any"
func ReplaceOperation() {}

// NotApiOperation this is not an operation to be document in openapi specification
func NotApiOperation() {}

func NotDocumented() {}
