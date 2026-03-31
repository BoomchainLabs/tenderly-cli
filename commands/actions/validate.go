package actions

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/config"
	actionsModel "github.com/tenderly/tenderly-cli/model/actions"
	"gopkg.in/yaml.v3"
)

var validateJSON bool

func init() {
	validateCmd.Flags().BoolVar(&validateJSON, "json", false, "Output validation results as JSON")
	actionsCmd.AddCommand(validateCmd)
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate tenderly.yaml actions configuration.",
	Long:  "Validates your tenderly.yaml against the actions JSON Schema and checks trigger configuration for errors. No login or API access required.",
	Run: func(cmd *cobra.Command, args []string) {
		result := runValidation()
		if validateJSON {
			renderValidateJSON(result)
		} else {
			renderValidateText(result)
		}
		if !result.Valid {
			os.Exit(1)
		}
	},
}

// --- Output types ---

type validateOutput struct {
	Valid         bool                 `json:"valid"`
	SchemaErrors []string             `json:"schema_errors,omitempty"`
	TriggerErrors []triggerErrorOutput `json:"trigger_errors,omitempty"`
}

type triggerErrorOutput struct {
	Project string   `json:"project"`
	Action  string   `json:"action,omitempty"`
	Errors  []string `json:"errors"`
}

// --- Core validation logic ---

func runValidation() *validateOutput {
	if !config.IsAnyActionsInit() {
		return &validateOutput{
			Valid:        false,
			SchemaErrors: []string{"actions not initialized: tenderly.yaml not found"},
		}
	}

	content, err := config.ReadProjectConfig()
	if err != nil {
		return &validateOutput{
			Valid:        false,
			SchemaErrors: []string{fmt.Sprintf("failed reading tenderly.yaml: %s", err)},
		}
	}

	result := &validateOutput{Valid: true}

	// Phase 1: JSON Schema validation
	schemaErrors, err := actionsModel.ValidateConfig(content)
	if err != nil {
		return &validateOutput{
			Valid:        false,
			SchemaErrors: []string{fmt.Sprintf("schema validation error: %s", err)},
		}
	}
	if len(schemaErrors) > 0 {
		result.Valid = false
		result.SchemaErrors = schemaErrors
	}

	// Phase 2: Go-level validation
	// Parse actions from YAML directly (avoid MustGetActions which calls os.Exit)
	var tenderlyYaml actionsTenderlyYaml
	if err := yaml.Unmarshal(content, &tenderlyYaml); err != nil {
		result.Valid = false
		result.SchemaErrors = append(result.SchemaErrors, fmt.Sprintf("failed parsing YAML: %s", err))
		return result
	}
	allActions := tenderlyYaml.Actions
	projectsToValidate := allActions
	if actionsProjectName != "" {
		projectsToValidate = make(map[string]actionsModel.ProjectActions)
		for slug, pa := range allActions {
			if strings.EqualFold(slug, actionsProjectName) {
				projectsToValidate[slug] = pa
			}
		}
		if len(projectsToValidate) == 0 {
			return &validateOutput{
				Valid:        false,
				SchemaErrors: []string{fmt.Sprintf("project %s not found in tenderly.yaml", actionsProjectName)},
			}
		}
	}

	for slug, pa := range projectsToValidate {
		if !actionsModel.IsRuntimeSupported(pa.Runtime) {
			result.Valid = false
			result.TriggerErrors = append(result.TriggerErrors, triggerErrorOutput{
				Project: slug,
				Errors:  []string{fmt.Sprintf("invalid runtime %s", pa.Runtime)},
			})
		}

		for name, spec := range pa.Specs {
			var actionErrors []string

			if spec.ExecutionType != actionsModel.ParallelExecutionType &&
				spec.ExecutionType != actionsModel.SequentialExecutionType &&
				spec.ExecutionType != "" {
				actionErrors = append(actionErrors, fmt.Sprintf("invalid execution_type %s", spec.ExecutionType))
			}

			if err := spec.Parse(); err != nil {
				actionErrors = append(actionErrors, fmt.Sprintf("failed parsing trigger: %s", err))
			} else {
				response := spec.TriggerParsed.Validate(actionsModel.ValidatorContext(name + ".trigger"))
				if len(response.Errors) > 0 {
					actionErrors = append(actionErrors, response.Errors...)
				}
			}

			if len(actionErrors) > 0 {
				result.Valid = false
				result.TriggerErrors = append(result.TriggerErrors, triggerErrorOutput{
					Project: slug,
					Action:  name,
					Errors:  actionErrors,
				})
			}
		}
	}

	return result
}

// --- Renderers ---

func renderValidateJSON(result *validateOutput) {
	bytes, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(bytes))
}

func renderValidateText(result *validateOutput) {
	logrus.Info("\nValidating against JSON Schema...")
	if len(result.SchemaErrors) > 0 {
		for _, e := range result.SchemaErrors {
			logrus.Info(commands.Colorizer.Red("  " + e))
		}
	} else {
		logrus.Info(commands.Colorizer.Green("  Schema validation passed."))
	}

	logrus.Info("\nValidating triggers configuration...")
	if len(result.TriggerErrors) > 0 {
		for _, te := range result.TriggerErrors {
			label := te.Project
			if te.Action != "" {
				label = te.Action
			}
			for _, e := range te.Errors {
				logrus.Info(commands.Colorizer.Sprintf("  %s %s: %s",
					commands.Colorizer.Red("x"),
					commands.Colorizer.Bold(label),
					commands.Colorizer.Red(e),
				))
			}
		}
	}

	if !result.Valid {
		logrus.Info("")
		logrus.Error(commands.Colorizer.Bold(commands.Colorizer.Red("Validation failed.")))
		return
	}

	logrus.Info(commands.Colorizer.Sprintf("\n%s",
		commands.Colorizer.Bold(commands.Colorizer.Green("Validation passed.")),
	))
}
