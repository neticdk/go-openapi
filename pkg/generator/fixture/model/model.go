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
