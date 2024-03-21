package generator

import (
	"fmt"
	"go/ast"
	"go/types"
	"regexp"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/rs/zerolog/log"
	"golang.org/x/tools/go/packages"
)

const refPrefix = "/definitions"

var (
	openapiComponentExp = regexp.MustCompile(`^//openapi:component schema (\w+)$`)

	schemaExampleExp = regexp.MustCompile("^//(openapi|schema):example (.*)$")
	schemaFormatExp  = regexp.MustCompile("^//(openapi|schema):format (.*)$")
	schemaDefaultExp = regexp.MustCompile("^//(openapi|schema):default (.*)$")
)

var simpleTypeMap = map[types.BasicKind]func() *spec.Schema{
	types.Bool:    spec.BoolProperty,
	types.Int:     spec.Int32Property,
	types.Int8:    spec.Int8Property,
	types.Int16:   spec.Int16Property,
	types.Int32:   spec.Int32Property,
	types.Int64:   spec.Int64Property,
	types.Uint:    spec.Int32Property,
	types.Uint8:   spec.Int8Property,
	types.Uint16:  spec.Int16Property,
	types.Uint32:  spec.Int32Property,
	types.Uint64:  spec.Int64Property,
	types.Float32: spec.Float32Property,
	types.Float64: spec.Float64Property,
	types.String:  spec.StringProperty,
}

var jsonTag = regexp.MustCompile(`json:"([^"]*)"`)

func GenerateSchemas(pkgs []*packages.Package) map[string]*spec.Schema {
	sg := &schemaGenerator{
		schemas: map[string]*spec.Schema{},
	}
	return sg.Generate(pkgs)
}

type schemaGenerator struct {
	schemas map[string]*spec.Schema
}

func (sg *schemaGenerator) Generate(pkgs []*packages.Package) map[string]*spec.Schema {
	for _, p := range pkgs {
		for _, f := range p.Syntax { // Entry for each file in package
			for _, d := range f.Decls {
				gd, ok := d.(*ast.GenDecl)
				if !ok {
					continue
				}

				for _, s := range gd.Specs {
					ts, ok := s.(*ast.TypeSpec)
					if !ok {
						continue
					}

					componentID := ""
					if ts.Doc != nil {
						for _, cmt := range ts.Doc.List {
							m := openapiComponentExp.FindStringSubmatch(cmt.Text)
							if m != nil {
								componentID = m[1]
							}
						}
					}
					if componentID == "" && gd.Doc != nil {
						for _, cmt := range gd.Doc.List {
							m := openapiComponentExp.FindStringSubmatch(cmt.Text)
							if m != nil {
								componentID = m[1]
							}
						}
					}

					if componentID != "" { // Component was identified
						description := ""
						if ts.Doc != nil {
							description = ts.Doc.Text()
						} else if gd.Doc != nil {
							description = gd.Doc.Text()
						}

						schema := sg.schema(p, ts, description)
						// if schema == nil --> ERROR
						// TODO: Check for existing schema!!
						sg.schemas[componentID] = schema
					}

				}
			}
		}
	}

	return sg.schemas
}

func (sg *schemaGenerator) schema(p *packages.Package, ts *ast.TypeSpec, description string) *spec.Schema {
	def, ok := p.TypesInfo.Defs[ts.Name]
	if !ok {
		log.Warn().Str("type", ts.Name.String()).Msg("No type info found")
		return nil
	}

	properties := map[string]spec.Schema{}

	var schema *spec.Schema
	switch ut := def.Type().Underlying().(type) {
	case *types.Struct: // If this is not a struct? Only structs qualify for components??
		for i := 0; i < ut.NumFields(); i++ {
			field := ut.Field(i)
			if field.Exported() || field.Embedded() {
				astStruct, ok := ts.Type.(*ast.StructType)
				if !ok {
					// Only structs qualify as components - DO WARNING!!
					continue
				}
				astField := astStruct.Fields.List[i] // Expect the order to match the types fields

				propertyName := field.Name()
				tags := ut.Tag(i)
				m := jsonTag.FindStringSubmatch(tags)
				if m != nil {
					field, _, _ := strings.Cut(m[1], ",") // Consider if json tag options may ve relevant
					if field != "" {
						propertyName = field
					}
				}

				props := sg.handleField(p, field.Type(), propertyName, field.Embedded(), astField.Doc)
				for k, v := range props {
					properties[k] = v
				}
			}
		}
		schema = &spec.Schema{
			SchemaProps: spec.SchemaProps{Type: []string{"object"}},
		}

	default:
		props := sg.handleField(p, def.Type().Underlying(), def.Name(), false, ts.Doc)
		prop := props[def.Name()]
		schema = &prop
	}

	return schema.
		WithDescription(description).
		WithProperties(properties)
}

func (sg *schemaGenerator) handleField(p *packages.Package, t types.Type, name string, embedded bool, doc *ast.CommentGroup) map[string]spec.Schema {
	var prop *spec.Schema
	switch fieldType := t.(type) {
	case *types.Basic:
		if fn, ok := simpleTypeMap[fieldType.Kind()]; ok {
			prop = fn()
		} else {
			log.Warn().Int("kind", int(fieldType.Kind())).Msg("Unable to map go simple type to json schema type")
			prop = spec.StringProperty()
		}

	case *types.Named:
		prop = checkKnownTypes(fieldType.Obj())
		if prop != nil {
			break // break from the switch
		}

		var pkg *packages.Package
		if fieldType.Obj().Pkg() == p.Types {
			pkg = p
		} else {
			pkg = p.Imports[fieldType.Obj().Pkg().Path()]
		}

		typeSpec := findTypeSpec(pkg, fieldType.Obj().Name())
		if typeSpec != nil {
			sch := sg.schema(pkg, typeSpec, doc.Text())

			if embedded {
				return sch.Properties
			} else {
				// TODO: Check it already exists...
				sg.schemas[fieldType.Obj().Name()] = sch
				prop = spec.RefSchema(fmt.Sprintf("#%s/%s", refPrefix, fieldType.Obj().Name()))
			}

		} else {
			log.Warn().Str("package", pkg.String()).Str("type", fieldType.Obj().Name()).Msg("Unable to find ast type specification")
		}

	case *types.Slice:
		elementProps := sg.handleField(p, fieldType.Elem(), "item", false, nil)
		elSchema := elementProps["item"]
		prop = spec.ArrayProperty(&elSchema)

	case *types.Map:
		elementProps := sg.handleField(p, fieldType.Elem(), "item", false, nil)
		elSchema := elementProps["item"]
		prop = spec.MapProperty(&elSchema)

	case *types.Pointer:
		return sg.handleField(p, fieldType.Elem(), name, embedded, doc)

	case *types.Interface:
		prop = &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"object"}}}

	default:
		log.Warn().Str("field", name).Type("type", t).Msg("Unsupported type for field")
		return map[string]spec.Schema{}
	}

	if doc != nil {
		handleGodoc(prop, doc)
	}

	properties := map[string]spec.Schema{}
	properties[name] = *prop
	return properties
}

func handleGodoc(prop *spec.Schema, doc *ast.CommentGroup) *spec.Schema {
	prop.Description = doc.Text()

	for _, c := range doc.List {
		exampleMatch := schemaExampleExp.FindStringSubmatch(c.Text)
		if exampleMatch != nil {
			prop = prop.WithExample(exampleMatch[2])
		}

		formatMatch := schemaFormatExp.FindStringSubmatch(c.Text)
		if formatMatch != nil {
			prop.Format = formatMatch[2]
		}

		defaultMatch := schemaDefaultExp.FindStringSubmatch(c.Text)
		if defaultMatch != nil {
			prop = prop.WithDefault(defaultMatch[2])
		}
	}

	return prop
}

func findTypeSpec(p *packages.Package, fieldName string) *ast.TypeSpec {
	for _, af := range p.Syntax {
		for _, de := range af.Decls {
			if gd, ok := de.(*ast.GenDecl); ok {
				for _, spec := range gd.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if typeSpec.Name.Name == fieldName {
							return typeSpec
						}
					}
				}
			}
		}
	}
	return nil
}

func checkKnownTypes(t *types.TypeName) *spec.Schema {
	if t.Pkg().Path() == "time" && t.Name() == "Time" {
		return spec.DateTimeProperty()
	}
	return nil
}
