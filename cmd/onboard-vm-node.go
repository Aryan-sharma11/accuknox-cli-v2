package cmd

import (
	"fmt"

	"github.com/accuknox/accuknox-cli-v2/pkg/onboard"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	kubeArmorAddr   string
	relayServerAddr string
	siaAddr         string
	peaAddr         string
	hardenAddr      string
)

// joinNodeCmd represents the join command
var joinNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Join this worker node with the control plane node for onboarding onto SaaS",
	Long:  "Join this worker node with the control plane node for onboarding onto SaaS",
	RunE: func(cmd *cobra.Command, args []string) error {
		// need at least either one of the below flags
		if nodeAddr == "" && (siaAddr == "" || relayServerAddr == "" || peaAddr == "" || hardenAddr == "") {
			return fmt.Errorf(color.RedString("cp-node-addr (control-plane address) or address of each agent must be specified"))
		}
		// validate environment for pre-requisites
		var cc onboard.ClusterConfig
		_, err := cc.ValidateEnv()
		if vmMode == "" {
			if err == nil {
				vmMode = onboard.VMMode_Docker
			} else {
				fmt.Printf(
					color.YellowString("Warning: Docker requirements did not match:\n%s.\nFalling back to systemd mode for installation.\n", err.Error()))
				vmMode = onboard.VMMode_Systemd
			}
		} else if vmMode == onboard.VMMode_Docker && err != nil {
			// docker mode specified explicitly but requirements didn't match
			return fmt.Errorf(color.RedString("failed to validate environment: %s", err.Error()))
		}

		vmConfigs, err := onboard.CreateClusterConfig(onboard.ClusterType_VM, userConfigPath, vmMode,
			vmAdapterTag, kubeArmorRelayServerTag, peaVersionTag, siaVersionTag,
			feederVersionTag, sumEngineVersionTag, discoverVersionTag, hardeningAgentVersionTag, kubearmorVersion, releaseVersion, kubeArmorImage,
			kubeArmorInitImage, kubeArmorVMAdapterImage, kubeArmorRelayServerImage, siaImage,
			peaImage, feederImage, sumEngineImage, hardeningAgentImage, spireAgentImage, discoverImage, nodeAddr, dryRun,
			true, imagePullPolicy, visibility, hostVisibility, audit, block, hostAudit, hostBlock,
			cidr, secureContainers, skipBTF, systemMonitorPath)
		if err != nil {
			return fmt.Errorf(color.RedString("failed to create VM config: %s", err.Error()))
		}
		joinConfig := onboard.JoinClusterConfig(*vmConfigs, kubeArmorAddr, relayServerAddr, siaAddr, peaAddr, hardenAddr)

		err = joinConfig.CreateBaseNodeConfig()
		if err != nil {
			return fmt.Errorf(color.RedString("failed to create VM config: %s", err.Error()))
		}

		switch vmMode {

		case onboard.VMMode_Systemd:
			if err := joinConfig.JoinSystemdNode(); err != nil {
				return fmt.Errorf(color.RedString("failed to join worker node: %s", err.Error()))
			}

		case onboard.VMMode_Docker:
			err = joinConfig.JoinWorkerNode()
			if err != nil {
				return fmt.Errorf(color.RedString("failed to join worker node: %s", err.Error()))
			}

		default:
			return fmt.Errorf(color.RedString("vm mode: %s invalid, accepted values (docker/systemd)", vmMode))
		}

		fmt.Println(color.GreenString("VM successfully joined with control-plane!"))
		return nil
	},
}

func init() {
	// configuration for connecting with KubeArmor and control plane
	joinNodeCmd.PersistentFlags().StringVar(&kubeArmorAddr, "kubearmor-addr", "", "address of kubearmor on this node")
	joinNodeCmd.PersistentFlags().StringVar(&relayServerAddr, "relay-server-addr", "", "address of relay-server on control plane to connect with for pushing telemetry events")
	joinNodeCmd.PersistentFlags().StringVar(&siaAddr, "sia-addr", "", "address of shared-informer-agent on control plane to push state events")
	joinNodeCmd.PersistentFlags().StringVar(&peaAddr, "pea-addr", "", "address of policy-enforcement-agent on control plane for receiving policy events")
	joinNodeCmd.PersistentFlags().StringVar(&hardenAddr, "harden-addr", "", "address of hardening-agent on control plane for receiving state events")
	joinNodeCmd.PersistentFlags().StringVar(&nodeAddr, "cp-node-addr", "", "address of control plane")

	joinNodeCmd.PersistentFlags().StringVarP(&releaseVersion, "version", "v", "", "version to use - recommended to keep same as control plane node version")

	onboardVMCmd.AddCommand(joinNodeCmd)
}
