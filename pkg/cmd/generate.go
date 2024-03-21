package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/neticdk/go-openapi/pkg/generator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/tools/go/packages"
)

const (
	output = "output"
)

var (
	generateCmd = &cobra.Command{Use: "generate [packages]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &packages.Config{
				Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
			}
			pkgs, err := packages.Load(cfg, args...)
			if err != nil {
				return fmt.Errorf("unable to load packages: %w", err)
			}

			spec := generator.GenerateSpec(pkgs)

			var file *os.File
			if viper.GetString(output) == "-" {
				file = os.Stdout
			} else {
				file, err = os.Create(viper.GetString(output))
				if err != nil {
					return fmt.Errorf("unable to create output file %s: %w", viper.GetString(output), err)
				}
				defer file.Close()
			}

			err = json.NewEncoder(file).Encode(spec)
			if err != nil {
				return fmt.Errorf("unable to encode openapi specification: %w", err)
			}

			return nil
		},
	}
)

func init() {
	generateCmd.Flags().StringP(output, "o", "openapi.json", "Output file for openapi specification document")
	viper.BindPFlag(output, generateCmd.Flags().Lookup(output))

	rootCmd.AddCommand(generateCmd)
}
