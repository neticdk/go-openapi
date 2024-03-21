package generator

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/rs/zerolog/log"
	"golang.org/x/tools/go/packages"
)

const (
	parameterQuery = "query"
	paramaterPath  = "path"
)

var (
	openapiOperationExp = regexp.MustCompile(`^//openapi:operation (\S+) (get|GET|put|PUT|post|POST|delete|DELETE|options|OPTIONS|head|HEAD|patch|PATCH|trace|TRACE)$`)
)

var opTypes = map[string]func(*spec.PathItem, *spec.Operation){
	http.MethodGet:     func(pi *spec.PathItem, op *spec.Operation) { pi.Get = op },
	http.MethodPut:     func(pi *spec.PathItem, op *spec.Operation) { pi.Put = op },
	http.MethodPost:    func(pi *spec.PathItem, op *spec.Operation) { pi.Post = op },
	http.MethodDelete:  func(pi *spec.PathItem, op *spec.Operation) { pi.Delete = op },
	http.MethodOptions: func(pi *spec.PathItem, op *spec.Operation) { pi.Options = op },
	http.MethodHead:    func(pi *spec.PathItem, op *spec.Operation) { pi.Head = op },
	http.MethodPatch:   func(pi *spec.PathItem, op *spec.Operation) { pi.Patch = op },
}

var opDirectives = []*struct {
	expr *regexp.Regexp
	fn   func(*spec.Operation, string, []string)
}{
	{
		expr: regexp.MustCompile(`^//openapi:parameter (\w+) (path|query) (\w+)(/(\S+))?( "([^"]+)")?$`),
		fn:   func(op *spec.Operation, _ string, m []string) { handleParameter(op, m[1], m[2], m[3], m[5], m[7]) },
	},
	{
		expr: regexp.MustCompile(`^//openapi:tag (\w+)$`),
		fn:   func(op *spec.Operation, _ string, m []string) { op.WithTags(m[1]) },
	},
	{
		expr: regexp.MustCompile(`^//openapi:response (default|[0-9]{3})( "([^"]+)")?$`),
		fn:   func(op *spec.Operation, _ string, m []string) { handleResponseDescription(op, m[1], m[3]) },
	},
	{
		expr: regexp.MustCompile(`^//openapi:responseContent (default|[0-9]{3}) (\S+) (\w+)$`),
		fn:   func(op *spec.Operation, _ string, m []string) { handleResponseContent(op, m[1], m[2], m[3]) },
	},
	{
		expr: regexp.MustCompile(`^//openapi:responseHeader (default|[0-9]{3}) (\S+) (\w+)(/(\S+))?( "([^"]+)")?$`),
		fn:   func(op *spec.Operation, _ string, m []string) { handleResponseHeader(op, m[1], m[2], m[3], m[5], m[7]) },
	},
	{
		expr: regexp.MustCompile(`^//openapi:responseExample (default|[0-9]{3}) (\S+) (\S+)$`),
		fn:   func(op *spec.Operation, file string, m []string) { handleResponseExample(op, file, m[1], m[2], m[3]) },
	},
	{
		expr: regexp.MustCompile(`^//openapi:requestBody (\S+) (\w+)( (true|false))?( "([^"]+)")?$`),
		fn:   func(op *spec.Operation, _ string, m []string) { handleRequestBody(op, m[1], m[2], m[4], m[6]) },
	},
}

type operationGenerator struct {
	paths *spec.Paths
}

func GenerateOperations(pkgs []*packages.Package) *spec.Paths {
	og := &operationGenerator{
		paths: &spec.Paths{
			Paths: map[string]spec.PathItem{},
		},
	}
	return og.Generate(pkgs)
}

func (og *operationGenerator) Generate(pkgs []*packages.Package) *spec.Paths {
	for _, p := range pkgs {
		for i, f := range p.Syntax { // Entry for each file in package
			file := p.GoFiles[i]
			for _, d := range f.Decls {
				fd, ok := d.(*ast.FuncDecl)
				if !ok {
					continue
				}

				if fd.Doc != nil {
					for _, l := range fd.Doc.List {
						operationMatch := openapiOperationExp.FindStringSubmatch(l.Text)
						if operationMatch != nil {
							path, method := operationMatch[1], operationMatch[2]
							og.operation(fd.Name.String(), file, path, method, fd.Doc)
						}
					}
				}
			}
		}
	}
	return og.paths
}

func (og *operationGenerator) operation(id, file, path, method string, doc *ast.CommentGroup) {
	op := spec.NewOperation(id).WithDescription(doc.Text())

	for _, l := range doc.List {
		for _, dh := range opDirectives {
			m := dh.expr.FindStringSubmatch(l.Text)
			if m != nil {
				dh.fn(op, file, m)
			}
		}
	}

	if _, ok := og.paths.Paths[path]; !ok {
		og.paths.Paths[path] = spec.PathItem{}
	}
	p := og.paths.Paths[path]
	if ot, ok := opTypes[strings.ToUpper(method)]; ok {
		ot(&p, op)
	} else {
		log.Warn().Str("method", method).Msg("Unsupported method - this should not happen")
	}
	og.paths.Paths[path] = p // PathItem is value _not_ a ref reference so it has to be replaced
}

func handleParameter(op *spec.Operation, name, in, typ, format, description string) {
	var param *spec.Parameter
	if in == paramaterPath {
		param = spec.PathParam(name)
	} else if in == parameterQuery {
		param = spec.QueryParam(name)
	}
	param.
		WithDescription(description).
		Typed(typ, format)
	op.Parameters = append(op.Parameters, *param)
}

func handleRequestBody(op *spec.Operation, mediaType, schemaID, required, description string) {
	param := spec.BodyParam("body", spec.RefSchema(fmt.Sprintf("#%s/%s", refPrefix, schemaID))).
		WithDescription(description)
	if req, err := strconv.ParseBool(required); err != nil || req {
		param.AsRequired()
	}
	op.AddParam(param).
		WithConsumes(mediaType)
}

func handleResponseDescription(op *spec.Operation, code, descriptiopn string) {
	handleResponse(op, code, func(r *spec.Response) { r.WithDescription(descriptiopn) })
}

func handleResponseContent(op *spec.Operation, code, mediaType, schemaID string) {
	op.WithProduces(mediaType)
	handleResponse(op, code, func(r *spec.Response) {
		r.WithSchema(spec.RefSchema(fmt.Sprintf("#%s/%s", refPrefix, schemaID)))
	})
}

func handleResponseHeader(op *spec.Operation, code, name, typ, format, description string) {
	handleResponse(op, code, func(r *spec.Response) {
		h := spec.ResponseHeader().
			Typed(typ, format).
			WithDescription(description)
		r.AddHeader(name, h)
	})
}

func handleResponseExample(op *spec.Operation, sourceFile, code, mediaType, exampleFile string) {
	path := filepath.Dir(sourceFile)
	example, err := os.ReadFile(filepath.Join(path, exampleFile))
	if err != nil {
		log.Warn().Str("file", filepath.Join(path, exampleFile)).Msg("Unable to load example file")
		return
	}

	// See if example can be parsed as json
	var parsed interface{}
	err = json.Unmarshal(example, &parsed)
	if err != nil {
		parsed = string(example)
	}

	handleResponse(op, code, func(r *spec.Response) {
		r.AddExample(mediaType, parsed)
	})
}

func handleResponse(op *spec.Operation, code string, setter func(*spec.Response)) {
	if op.Responses == nil {
		op.Responses = &spec.Responses{}
	}

	var response *spec.Response
	if code == "default" && op.Responses.Default != nil {
		response = op.Responses.Default
	} else {
		c, _ := strconv.Atoi(code)
		if r, ok := op.Responses.StatusCodeResponses[c]; ok {
			response = &r
		} else {
			response = spec.NewResponse()
		}
	}
	setter(response)
	if code == "default" {
		op.Responses.Default = response
	} else {
		c, _ := strconv.Atoi(code)
		if op.Responses.StatusCodeResponses == nil {
			op.Responses.StatusCodeResponses = map[int]spec.Response{}
		}
		op.Responses.StatusCodeResponses[c] = *response
	}
}
