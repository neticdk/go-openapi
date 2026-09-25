package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

func TestGenerateSchemas(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, "./fixture/model/...")
	if assert.NoError(t, err) {
		schemas := GenerateSchemas(pkgs)
		assert.Len(t, schemas, 4)
		assert.Len(t, schemas["Model"].Properties, 12)
		assert.Len(t, schemas["Model"].Properties["field1"].Description, 18)
		assert.True(t, schemas["Model"].Properties["field9"].Type.Contains("object"))
		assert.True(t, schemas["Model"].Properties["field0"].Items.Schema.Type.Contains("object"))
		/*
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(schemas)
		*/
	}
}
