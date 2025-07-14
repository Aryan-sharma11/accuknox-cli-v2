package cmd

// This package will provide the kubearmor and its agents compatibility with the host vm.

import (
	"github.com/accuknox/accuknox-cli-v2/pkg/logger"
	"github.com/accuknox/accuknox-cli-v2/pkg/vm"
	"github.com/spf13/cobra"
)

var nodeType string

// vmInspectCmd represents the vm command for inspect
var vmInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect VM for compatibility",
	Long:  "Inspect VM for compatibility with KubeArmor and its agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := vm.InspectVM()
		if err != nil {
			logger.Error(err.Error())
			return err
		}
		return nil
	},
}

// ========== //
// == Init == //
// ========== //

func init() {
	vmCmd.AddCommand(vmInspectCmd)
}
