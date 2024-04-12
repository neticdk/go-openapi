package model

import (
	"time"

	"github.com/neticdk/go-openapi/pkg/generator/fixture"
)

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

// embeddedPrivate is a struct which will be embedded but not exported
type embeddedPrivate struct {
	EmbeddedField string `json:"eField"`
	privateField  string
}

type refPrivat struct {
	RefField string `json:"refField"`
}

type RefExported struct {
	RE string `jsong:"re"`
}

// Problem is simple implementation of [RFC9457]
//
// [RFC9457]: https://datatracker.ietf.org/doc/html/rfc9457
//
//openapi:component schema Problem
type Problem struct {
	// Type identify problem type RFC-9457#3.1.1
	//schema:format uri
	Type string `json:"type,omitempty"`

	// Status is the http status code and must be consistent with the server status code RFC-9457#3.1.2
	Status *int `json:"status,omitempty"`

	// Title is short humanreadable summary RFC-9457#3.1.3
	Title string `json:"title,omitempty"`

	// Detail is humanreadable explanation of the specific occurence of the problem RFC-9457#3.1.4
	Detail string `json:"detail,omitempty"`

	// Instance identifies the specific instance of the problem RFC-9457#3.1.5
	Instance string `json:"instance,omitempty"`

	// Err is containing wrapped error and will not be serialized to JSON
	Err error `json:"-"`
}
