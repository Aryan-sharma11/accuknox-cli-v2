package cmd

import (
	"fmt"
	"os"

	"github.com/accuknox/accuknox-cli-v2/pkg/deboard"
	"github.com/accuknox/accuknox-cli-v2/pkg/onboard"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// deboardNodeCmd represents the deboardNode command
var deboardNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Deboard a worker node",
	Long:  "Deboard a worker node",
	RunE: func(cmd *cobra.Command, args []string) error {

		if vmMode == "" {
			systemdInstallation, err := onboard.CheckSystemdInstallation()
			if err != nil {
				return fmt.Errorf(color.RedString("error checking systemd files: %s", err.Error()))
			}

			if systemdInstallation {
				vmMode = onboard.VMMode_Systemd
			} else {
				vmMode = onboard.VMMode_Docker
			}
		}

		switch vmMode {
		case onboard.VMMode_Systemd:
			_, err := deboard.Deboard(onboard.NodeType_WorkerNode, vmMode, dryRun)
			if err != nil {
				return fmt.Errorf(color.RedString("Failed to deboard worker node: %s", err.Error()))
			}
		case onboard.VMMode_Docker:
			configPath, err := deboard.Deboard(onboard.NodeType_WorkerNode, vmMode, dryRun)
			if err != nil && os.IsPermission(err) {
				fmt.Println(color.YellowString("Please remove any remaining resources at %s", configPath))
			} else if err != nil {
				return fmt.Errorf(color.RedString("Failed to deboard worker node: %s", err.Error()))
			}
		default:
			return fmt.Errorf(color.RedString("vm mode: %s invalid, accepted values (docker/systemd)", vmMode))
		}

		fmt.Println(color.GreenString("Worker node deboarded successfully."))
		return nil
	},
}

func init() {
	deboardVMCmd.AddCommand(deboardNodeCmd)
}
