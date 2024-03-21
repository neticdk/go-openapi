package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

func TestGenerateSpec(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, "./fixture/...")
	if assert.NoError(t, err) {
		spec := GenerateSpec(pkgs)
		assert.Equal(t, "1.0.0", spec.Info.Version)
		assert.Len(t, spec.Paths.Paths, 2)

		/*
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(spec)
			t.Fail()
		*/

	}
}
