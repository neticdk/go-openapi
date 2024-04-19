# go-openapi

Yet another take at handling [OpenAPI](https://www.openapis.org/) specifications from Go - specifically
this tooling is aimed at compile time generation of OpenAPI specification documents based on REST
endpoint written in Go.

The tool is based on a mix of godoc directives and information derived from go `structs` and simple
types. The godoc directives handles metadata that cannot be derived directly from the source code.

An example can be found in the [fixture](./pkg/generator/fixture) directory.

## Usage

Install the tool running the following command line:

```sh
go install github.com/neticdk/go-openapi/cmd/openapi@latest
```

The tool comes with two commands: `generate` and `expand`

### Generate OpenAPI Specification

To generate the OpenAPI Specification run `generate`. The below will generate an OpenAPI Specification
document to standard out based on the example in the [fixture](./pkg/generator/fixture) directory.

```sh
openapi generate -o- ./pkg/generator/fixture/...
```

### Expand OpenAPI Specification

The tool also supports inlining JSON schema definitions in the OpenAPI Specification. This can be
useful for other tools that iterate through the specification because they do not have to resolve
the references then. The below will inline the references in the give OpenAPI Specification document
and return the result on standard out.

```sh
openapi expand -o- openapi.json
```

## JSON Schema directives

The following directives can be used on fields in a Go `struct` types to indicate how these should be included
for JSON Schema rendering. The directives for JSON Schema can be used with both `schema:` or `openapi:`
prefixes although the below table shows the `schema:` prefix. All these directive take exactly one parameter.

| Directive        | Description                                                                                                                                                                                                  |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `schema:example` |  An example value for the annotated field                                                                                                                                                                    |
| `schema:format`  |  JSON Schema format of the annotated field. OpenAPI supports the [primitive data](https://datatracker.ietf.org/doc/html/draft-zyp-json-schema-04#section-3.5) types from JSON Scheme draft 04 specification. |
| `schema:default` |  Describes the default value of the annotated field.                                                                                                                                                         |

The below is an example of a annotated Go struct.

```go
// Model is a fixture for rendering models
//
//openapi:component schema Model
type Model struct {
  fixture.CommonType
  embeddedPrivate

  // Field1 description
  //schema:example mystring
  //schema:default def
  Field1 string `json:"field1"`

  //openapi:format uri
  Field2 int `json:"field2"`
  Field3 float32
  Field4 time.Time         `json:"timestamp"`
  Field5 *refPrivat        `json:"field5"`
  Field6 RefExported       `json:"field6"`
  Field7 []string          `json:"field7"`
  Field8 map[string]string `json:"field8"`
}
```

## OpenAPI directives

The following directives support providing metadata for specifically rendering the OpenAPI Specification document.

| Directive                 | Level           | Parameters                                            | Description                                                                                                                                                                                                                       |
| ------------------------- | --------------- | ----------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `openapi:info`            | Package Level   | `<version>`                                           | The directive indicates that package level godoc should be used for the general documentation in the generated specification. The `version` parameter will be used to fill out the version in the OpenAPI Specification document. |
| `openapi:component`       | Struct Level    | `<type>` `<name>`                                     | Indicates that the annotated struct should be included as a component in the specification. The only supported component type is currently `schema` the name is used for references from, e.g., a operation request or response.  |
| `openapi:operation`       | Function Level  |  `<path>` `<http-method>`                             | Marks the function as the implementation of a REST endpoint to be included in the specification with the given path and http method.                                                                                              |
| `openapi:parameter`       | Function Level  | `<name>` `<param-type>` `<type>` `[description]`      |  Describes a path or query parameter supported by the operation. `param-type` can be `path` or `query`. If describing a `path` parameter the `name` should match the placeholder in the given path. The description is optional.  |
| `openapi:tag`             | Function Level  | `<tag>`                                               | Adds tag to the operation in the specification                                                                                                                                                                                    |
| `openapi:requestBody`     |  Function Level | `<media-type>` `<model>` `[required]` `[description]` | Specifies a request body definition for the given media type. The `model` should reference a struct with the `openapi:component` directive. `required` is a boolean indicating whether the body is required to be present.        |
| `openapi:response`        |  Function Level |  `<code>` `[description]`                             | Add response definition to an operation. The `code` may be set to `default`. `description` is optional.                                                                                                                           |
| `openapi:responseContent` | Function Level  | `<code>` `<media-type>` `<model>`                     | Sets the content type and response schema for the given return code. The `code` may be set to `default`. The `model` should reference a struct with the `openapi:component` directive.                                            |
| `openapi:responseHeader`  |  Function Level | `<code>` `<media-type>` `<type>` `[description]`      | Specifies a response header for the given response code. The `code` may be set to `default`. The `type` is a JSON primitive type definition. The description is optional.                                                         |
| `openapi:responseExample` | Function Level  | `<code>` `<media-type>` `<file>`                      |  Specifies to include an example response for the given response code and media type from a file. The `code` may be set to `default`.                                                                                             |

The below is an exmple of specifying general information for the generated OpenAPI Specification document.

```go
// Package fixture Fixture Demo API
//
// The package (and subpackage) provides test fixture and illustrates the use of the godoc directives
//
//openapi:info 1.0.0
package fixture
```

The below is an example of using the directives for a number of REST endpoints.

```go
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
```
