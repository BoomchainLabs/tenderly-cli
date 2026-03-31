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
	"github.com/tenderly/tenderly-cli/userError"
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
	Run:   validateFunc,
}

// JSON output types

type validateOutput struct {
	Valid         bool                  `json:"valid"`
	SchemaErrors []string              `json:"schema_errors,omitempty"`
	TriggerErrors []triggerErrorOutput  `json:"trigger_errors,omitempty"`
}

type triggerErrorOutput struct {
	Project string   `json:"project"`
	Action  string   `json:"action"`
	Errors  []string `json:"errors"`
}

func validateFunc(cmd *cobra.Command, args []string) {
	if validateJSON {
		logrus.SetLevel(logrus.FatalLevel)
	}

	if !config.IsAnyActionsInit() {
		if validateJSON {
			printValidateJSON(&validateOutput{
				Valid:        false,
				SchemaErrors: []string{"actions not initialized: tenderly.yaml not found"},
			})
			os.Exit(1)
		}
		logrus.Error(commands.Colorizer.Sprintf(
			"Actions not initialized. Are you in the right directory? Run %s to initialize project.",
			commands.Colorizer.Bold(commands.Colorizer.Red("tenderly actions init")),
		))
		os.Exit(1)
	}

	content, err := config.ReadProjectConfig()
	if err != nil {
		if validateJSON {
			printValidateJSON(&validateOutput{
				Valid:        false,
				SchemaErrors: []string{fmt.Sprintf("failed reading tenderly.yaml: %s", err)},
			})
			os.Exit(1)
		}
		userError.LogErrorf("failed reading project config: %s",
			userError.NewUserError(err, "Failed reading project's tenderly.yaml config."),
		)
		os.Exit(1)
	}

	result := &validateOutput{Valid: true}

	// Phase 1: JSON Schema validation
	logrus.Info("\nValidating against JSON Schema...")
	schemaErrors, err := actionsModel.ValidateConfig(content)
	if err != nil {
		if validateJSON {
			printValidateJSON(&validateOutput{
				Valid:        false,
				SchemaErrors: []string{fmt.Sprintf("schema validation error: %s", err)},
			})
			os.Exit(1)
		}
		userError.LogErrorf("schema validation failed: %s",
			userError.NewUserError(err, "Failed to run schema validation."),
		)
		os.Exit(1)
	}
	if len(schemaErrors) > 0 {
		result.Valid = false
		result.SchemaErrors = schemaErrors
		for _, e := range schemaErrors {
			logrus.Info(commands.Colorizer.Red("  " + e))
		}
	} else {
		logrus.Info(commands.Colorizer.Green("  Schema validation passed."))
	}

	// Phase 2: Go-level validation (parse + validate triggers)
	logrus.Info("\nValidating triggers configuration...")
	allActions := MustGetActions()

	projectsToValidate := allActions
	if actionsProjectName != "" {
		projectsToValidate = make(map[string]actionsModel.ProjectActions)
		for slug, pa := range allActions {
			if strings.EqualFold(slug, actionsProjectName) {
				projectsToValidate[slug] = pa
			}
		}
		if len(projectsToValidate) == 0 {
			if validateJSON {
				printValidateJSON(&validateOutput{
					Valid:        false,
					SchemaErrors: []string{fmt.Sprintf("project %s not found in tenderly.yaml", actionsProjectName)},
				})
				os.Exit(1)
			}
			logrus.Error(commands.Colorizer.Sprintf(
				"Project %s not found in tenderly.yaml.",
				commands.Colorizer.Bold(commands.Colorizer.Red(actionsProjectName)),
			))
			os.Exit(1)
		}
	}

	for slug, pa := range projectsToValidate {
		logrus.Info(commands.Colorizer.Sprintf("\n  Project: %s",
			commands.Colorizer.Bold(commands.Colorizer.Blue(slug)),
		))

		if !actionsModel.IsRuntimeSupported(pa.Runtime) {
			result.Valid = false
			result.TriggerErrors = append(result.TriggerErrors, triggerErrorOutput{
				Project: slug,
				Action:  "",
				Errors:  []string{fmt.Sprintf("invalid runtime %s", pa.Runtime)},
			})
			logrus.Info(commands.Colorizer.Sprintf(
				"    %s Invalid runtime %s. Supported: {%s}",
				commands.Colorizer.Red("x"),
				commands.Colorizer.Bold(commands.Colorizer.Red(pa.Runtime)),
				commands.Colorizer.Green(strings.Join(actionsModel.SupportedRuntimes, ", ")),
			))
		}

		for name, spec := range pa.Specs {
			var actionErrors []string

			if spec.ExecutionType != actionsModel.ParallelExecutionType &&
				spec.ExecutionType != actionsModel.SequentialExecutionType &&
				spec.ExecutionType != "" {
				actionErrors = append(actionErrors, fmt.Sprintf("invalid execution_type %s", spec.ExecutionType))
				logrus.Info(commands.Colorizer.Sprintf(
					"    %s %s: invalid execution_type %s",
					commands.Colorizer.Red("x"),
					commands.Colorizer.Bold(name),
					commands.Colorizer.Red(spec.ExecutionType),
				))
			}

			if err := spec.Parse(); err != nil {
				actionErrors = append(actionErrors, fmt.Sprintf("failed parsing trigger: %s", err))
				logrus.Info(commands.Colorizer.Sprintf(
					"    %s %s: failed parsing trigger: %s",
					commands.Colorizer.Red("x"),
					commands.Colorizer.Bold(name),
					err,
				))
			} else {
				response := spec.TriggerParsed.Validate(actionsModel.ValidatorContext(name + ".trigger"))
				for _, i := range response.Infos {
					logrus.Info(commands.Colorizer.Sprintf("    %s %s",
						commands.Colorizer.Blue("i"),
						commands.Colorizer.Blue(i),
					))
				}
				if len(response.Errors) > 0 {
					actionErrors = append(actionErrors, response.Errors...)
					for _, e := range response.Errors {
						logrus.Info(commands.Colorizer.Sprintf("    %s %s",
							commands.Colorizer.Red("x"),
							commands.Colorizer.Red(e),
						))
					}
				} else {
					logrus.Info(commands.Colorizer.Sprintf("    %s %s",
						commands.Colorizer.Green("ok"),
						commands.Colorizer.Green(name),
					))
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

	if validateJSON {
		printValidateJSON(result)
		if !result.Valid {
			os.Exit(1)
		}
		return
	}

	if !result.Valid {
		logrus.Info("")
		logrus.Error(commands.Colorizer.Bold(commands.Colorizer.Red("Validation failed.")))
		os.Exit(1)
	}

	logrus.Info(commands.Colorizer.Sprintf("\n%s",
		commands.Colorizer.Bold(commands.Colorizer.Green("Validation passed.")),
	))
}

func printValidateJSON(result *validateOutput) {
	bytes, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(bytes))
}
