package actions

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
	actionsModel "github.com/tenderly/tenderly-cli/model/actions"
	"github.com/tenderly/tenderly-cli/userError"
)

func init() {
	actionsCmd.AddCommand(capabilitiesCmd)
}

type capabilitiesOutput struct {
	Version                 string                 `json:"version"`
	Commands                []commandInfo          `json:"commands"`
	TriggerTypes            []string               `json:"trigger_types"`
	Runtimes                []string               `json:"runtimes"`
	ExecutionTypes          []string               `json:"execution_types"`
	Intervals               []string               `json:"intervals"`
	Invocations             []string               `json:"invocations"`
	StatusValues            []string               `json:"status_values"`
	TransactionStatusValues []string               `json:"transaction_status_values"`
	Schema                  map[string]interface{} `json:"schema"`
}

type commandInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

var capabilitiesCmd = &cobra.Command{
	Use:   "capabilities",
	Short: "Output JSON manifest of CLI capabilities for agent/tooling discovery.",
	Run: func(cmd *cobra.Command, args []string) {
		var cmds []commandInfo
		for _, sub := range actionsCmd.Commands() {
			cmds = append(cmds, commandInfo{
				Name:        sub.Name(),
				Description: sub.Short,
			})
		}

		out := capabilitiesOutput{
			Version:                 commands.CurrentCLIVersion,
			Commands:                cmds,
			TriggerTypes:            actionsModel.TriggerTypes,
			Runtimes:                actionsModel.SupportedRuntimes,
			ExecutionTypes:          []string{actionsModel.SequentialExecutionType, actionsModel.ParallelExecutionType},
			Intervals:               actionsModel.Intervals,
			Invocations:             actionsModel.Invocations,
			StatusValues:            []string{"success", "fail"},
			TransactionStatusValues: []string{"mined", "confirmed10"},
			Schema:                  actionsModel.GenerateJSONSchema(),
		}

		bytes, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			userError.LogErrorf("failed marshalling capabilities: %s",
				userError.NewUserError(err, "Failed to generate capabilities output."),
			)
			os.Exit(1)
		}

		fmt.Println(string(bytes))
	},
}
