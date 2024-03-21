package generator

import (
	"regexp"
	"strings"

	"github.com/go-openapi/spec"
	"golang.org/x/tools/go/packages"
)

var (
	stripPackageDecl = regexp.MustCompile(`(?ms:\A(Package \S+ )?([^\n]+)\n(.*)\z)`)
	openapiInfoExp   = regexp.MustCompile(`^//openapi:info (\S+)$`)
)

func GenerateSpec(pkgs []*packages.Package) *spec.Swagger {
	openapi := &spec.Swagger{}
	openapi.Swagger = "2.0"
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			if file.Doc != nil {
				description := file.Doc.Text()
				title := "not found"
				stripMatch := stripPackageDecl.FindStringSubmatch(file.Doc.Text())
				if stripMatch != nil {
					title = stripMatch[2]
					description = strings.TrimSpace(stripMatch[3])
				}

				for _, l := range file.Doc.List {
					infoMatch := openapiInfoExp.FindStringSubmatch(l.Text)
					if infoMatch != nil {
						openapi.Info = &spec.Info{InfoProps: spec.InfoProps{
							Title:       title,
							Description: description,
							Version:     infoMatch[1],
						}}
					}
				}

			}
		}
	}

	schemas := GenerateSchemas(pkgs)
	defs := spec.Definitions{}
	for id, schema := range schemas {
		defs[id] = *schema
	}
	openapi.Definitions = defs

	openapi.Paths = GenerateOperations(pkgs)

	return openapi
}
