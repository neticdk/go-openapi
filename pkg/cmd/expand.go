package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-openapi/loads"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	expandOutput = "expand.output"
)

var (
	expandCmd = &cobra.Command{
		Use:   "expand [openapi-file]",
		Short: "Expand OpenAPI specification by inlining schema definitions",
		Long:  "Running this command will inline all schema references such that the path and operation definitions are self-contained. This can be useful for tooling rendering OpenAPI spcification documents to other formats.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := loads.Spec(args[0])
			if err != nil {
				return fmt.Errorf("unable to load openapi spec: %w", err)
			}

			exp, err := doc.Expanded()
			if err != nil {
				return fmt.Errorf("unable to expand openapi spec: %w", err)
			}

			var file *os.File
			if viper.GetString(expandOutput) == "-" {
				file = os.Stdout
			} else {
				file, err = os.Create(viper.GetString(expandOutput))
				if err != nil {
					return fmt.Errorf("unable to create output file %s: %w", viper.GetString(expandOutput), err)
				}
				defer file.Close()
			}

			err = json.NewEncoder(file).Encode(exp.Spec())
			if err != nil {
				return fmt.Errorf("unable to encode openapi specification: %w", err)
			}

			return nil
		},
	}
)

func init() {
	expandCmd.Flags().StringP("output", "o", "openapi.json", "Output file for openapi specification document")
	viper.BindPFlag(expandOutput, expandCmd.Flags().Lookup("output"))

	rootCmd.AddCommand(expandCmd)
}
