package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestGenerateOperations(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, "./fixture/api/...")
	if assert.NoError(t, err) {
		paths := GenerateOperations(pkgs)
		assert.Len(t, paths.Paths, 2)
		assert.NotNil(t, paths.Paths["/entities"].Get)

		require.NotNil(t, paths.Paths["/entities/{id}"].Get)
		assert.Len(t, paths.Paths["/entities/{id}"].Get.Responses.Default.Examples["application/ld+json"], 2)

		require.NotNil(t, paths.Paths["/entities/{id}"].Put)
		assert.Len(t, paths.Paths["/entities/{id}"].Put.Parameters, 2)

		/*
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(paths)
			t.Fail()
		*/
	}
}
