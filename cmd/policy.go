package cmd

import (
	"errors"
	"net"

	"github.com/kubearmor/kubearmor-client/vm"
	"github.com/spf13/cobra"
)

var policyOptions vm.PolicyOptions

// vmPolicyCmd represents the vm command for policy enforcement
var vmPolicyCmd = &cobra.Command{
	Use:   "policy",
	Short: "policy handling for non-k8s",
	Long:  "policy handling for non-k8s",
}

// vmPolicyAddCmd represents the vm add policy command for policy enforcement
var vmPolicyAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add policy for non-k8s",
	Long:  "add policy for non-k8s",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a path to valid policy YAML as argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create http address
		httpAddress := "http://" + net.JoinHostPort(HTTPIP, HTTPPort)

		if err := vm.PolicyHandling("ADDED", args[0], policyOptions, httpAddress, false); err != nil {
			return err
		}
		return nil
	},
}

// vmPolicyDeleteCmd represents the vm delete policy command for policy enforcement
var vmPolicyDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete policy for non-k8s",
	Long:  "delete policy for non-k8s",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a path to valid policy YAML as argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		httpAddress := "http://" + net.JoinHostPort(HTTPIP, HTTPPort)

		if err := vm.PolicyHandling("DELETED", args[0], policyOptions, httpAddress, false); err != nil {
			return err
		}
		return nil
	},
}

// ========== //
// == Init == //
// ========== //

func init() {
	vmCmd.AddCommand(vmPolicyCmd)

	// Subcommand for policy command
	vmPolicyCmd.AddCommand(vmPolicyAddCmd)
	vmPolicyCmd.AddCommand(vmPolicyDeleteCmd)

	// gRPC endpoint flag to communicate with KubeArmor. Available across subcommands.
	vmPolicyCmd.PersistentFlags().StringVar(&policyOptions.GRPC, "gRPC", "", "gRPC server information")
}
