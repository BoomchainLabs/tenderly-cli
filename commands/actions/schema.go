package actions

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	actionsModel "github.com/tenderly/tenderly-cli/model/actions"
	"github.com/tenderly/tenderly-cli/userError"
)

const SchemaFileName = "tenderly-schema.json"

var schemaOutputFile string

func init() {
	schemaCmd.Flags().StringVarP(&schemaOutputFile, "output", "o", "", "Output file path (default: stdout)")
	actionsCmd.AddCommand(schemaCmd)
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Generate JSON Schema for tenderly.yaml actions configuration.",
	Long:  "Outputs a JSON Schema that describes the structure of the actions section in tenderly.yaml. Use with VS Code (redhat.vscode-yaml) or JetBrains for autocomplete and validation.",
	Run: func(cmd *cobra.Command, args []string) {
		schemaJSON, err := actionsModel.GenerateJSONSchemaString()
		if err != nil {
			userError.LogErrorf("failed generating schema: %s",
				userError.NewUserError(err, "Failed to generate JSON Schema."),
			)
			os.Exit(1)
		}

		if schemaOutputFile != "" {
			err = os.WriteFile(schemaOutputFile, []byte(schemaJSON+"\n"), 0644)
			if err != nil {
				userError.LogErrorf("failed writing schema file: %s",
					userError.NewUserError(err, fmt.Sprintf("Failed to write schema to %s.", schemaOutputFile)),
				)
				os.Exit(1)
			}
			logrus.Infof("Schema written to %s", schemaOutputFile)
		} else {
			fmt.Println(schemaJSON)
		}
	},
}

// mustWriteSchemaFile generates the schema file alongside tenderly.yaml.
func mustWriteSchemaFile() {
	schemaJSON, err := actionsModel.GenerateJSONSchemaString()
	if err != nil {
		userError.LogErrorf("failed generating schema: %s",
			userError.NewUserError(err, "Failed to generate JSON Schema."),
		)
		os.Exit(1)
	}

	err = os.WriteFile(SchemaFileName, []byte(schemaJSON+"\n"), 0644)
	if err != nil {
		userError.LogErrorf("failed writing schema file: %s",
			userError.NewUserError(err, fmt.Sprintf("Failed to write schema to %s.", SchemaFileName)),
		)
		os.Exit(1)
	}
}
