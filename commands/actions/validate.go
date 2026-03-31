package actions

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/config"
	actionsModel "github.com/tenderly/tenderly-cli/model/actions"
	"github.com/tenderly/tenderly-cli/userError"
)

func init() {
	actionsCmd.AddCommand(validateCmd)
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate tenderly.yaml actions configuration.",
	Long:  "Validates your tenderly.yaml against the actions JSON Schema and checks trigger configuration for errors. No login or API access required.",
	Run:   validateFunc,
}

func validateFunc(cmd *cobra.Command, args []string) {
	if !config.IsAnyActionsInit() {
		logrus.Error(commands.Colorizer.Sprintf(
			"Actions not initialized. Are you in the right directory? Run %s to initialize project.",
			commands.Colorizer.Bold(commands.Colorizer.Red("tenderly actions init")),
		))
		os.Exit(1)
	}

	content, err := config.ReadProjectConfig()
	if err != nil {
		userError.LogErrorf("failed reading project config: %s",
			userError.NewUserError(err, "Failed reading project's tenderly.yaml config."),
		)
		os.Exit(1)
	}

	hasErrors := false

	// Phase 1: JSON Schema validation
	logrus.Info("\nValidating against JSON Schema...")
	schemaErrors, err := actionsModel.ValidateConfig(content)
	if err != nil {
		userError.LogErrorf("schema validation failed: %s",
			userError.NewUserError(err, "Failed to run schema validation."),
		)
		os.Exit(1)
	}
	if len(schemaErrors) > 0 {
		hasErrors = true
		for _, e := range schemaErrors {
			logrus.Info(commands.Colorizer.Red("  " + e))
		}
	} else {
		logrus.Info(commands.Colorizer.Green("  Schema validation passed."))
	}

	// Phase 2: Go-level validation (parse + validate triggers)
	logrus.Info("\nValidating triggers configuration...")
	allActions := MustGetActions()

	// Filter to specific project if --project flag is set
	projectsToValidate := allActions
	if actionsProjectName != "" {
		projectsToValidate = make(map[string]actionsModel.ProjectActions)
		for slug, pa := range allActions {
			if strings.EqualFold(slug, actionsProjectName) {
				projectsToValidate[slug] = pa
			}
		}
		if len(projectsToValidate) == 0 {
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

		// Validate runtime
		if !actionsModel.IsRuntimeSupported(pa.Runtime) {
			hasErrors = true
			logrus.Info(commands.Colorizer.Sprintf(
				"    %s Invalid runtime %s. Supported: {%s}",
				commands.Colorizer.Red("x"),
				commands.Colorizer.Bold(commands.Colorizer.Red(pa.Runtime)),
				commands.Colorizer.Green(strings.Join(actionsModel.SupportedRuntimes, ", ")),
			))
		}

		for name, spec := range pa.Specs {
			// Validate execution type
			if spec.ExecutionType != actionsModel.ParallelExecutionType &&
				spec.ExecutionType != actionsModel.SequentialExecutionType &&
				spec.ExecutionType != "" {
				hasErrors = true
				logrus.Info(commands.Colorizer.Sprintf(
					"    %s %s: invalid execution_type %s",
					commands.Colorizer.Red("x"),
					commands.Colorizer.Bold(name),
					commands.Colorizer.Red(spec.ExecutionType),
				))
			}

			// Parse trigger
			if err := spec.Parse(); err != nil {
				hasErrors = true
				logrus.Info(commands.Colorizer.Sprintf(
					"    %s %s: failed parsing trigger: %s",
					commands.Colorizer.Red("x"),
					commands.Colorizer.Bold(name),
					err,
				))
				continue
			}

			// Validate trigger
			response := spec.TriggerParsed.Validate(actionsModel.ValidatorContext(name + ".trigger"))
			for _, i := range response.Infos {
				logrus.Info(commands.Colorizer.Sprintf("    %s %s",
					commands.Colorizer.Blue("i"),
					commands.Colorizer.Blue(i),
				))
			}
			if len(response.Errors) > 0 {
				hasErrors = true
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
	}

	if hasErrors {
		logrus.Info("")
		logrus.Error(commands.Colorizer.Bold(commands.Colorizer.Red("Validation failed.")))
		os.Exit(1)
	}

	logrus.Info(commands.Colorizer.Sprintf("\n%s",
		commands.Colorizer.Bold(commands.Colorizer.Green("Validation passed.")),
	))
}
