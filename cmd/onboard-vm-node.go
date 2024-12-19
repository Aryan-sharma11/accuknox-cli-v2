package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/accuknox/accuknox-cli-v2/pkg/common"
	"github.com/accuknox/accuknox-cli-v2/pkg/logger"
	"github.com/accuknox/accuknox-cli-v2/pkg/onboard"
	"github.com/spf13/cobra"
)

var (
	kubeArmorAddr   string
	relayServerAddr string
	siaAddr         string
	peaAddr         string
	hardenAddr      string

	deploySumegine bool
)

// joinNodeCmd represents the join command
var joinNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Join this worker node with the control plane node for onboarding onto SaaS",
	Long:  "Join this worker node with the control plane node for onboarding onto SaaS",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		// need at least either one of the below flags
		if nodeAddr == "" {
			if siaAddr == "" || relayServerAddr == "" || peaAddr == "" || hardenAddr == "" {
				logger.Error("cp-node-addr (control-plane address) or address of each agent must be specified")
				return err
			}

			if deploySumegine && rmqAddress == "" {
				logger.Error("cp-node-addr (control-plane address) or address of control plane RabbitMQ server must be specified")
				return err
			}
		}
		// validate environment for pre-requisites
		var cc onboard.ClusterConfig
		_, err = cc.ValidateEnv()
		if vmMode == "" {
			if err == nil {
				vmMode = onboard.VMMode_Docker
			} else {
				logger.Warn("Warning: Docker requirements did not match:\n%s.\nFalling back to systemd mode for installation.\n", err.Error())
				vmMode = onboard.VMMode_Systemd
			}
		} else if vmMode == onboard.VMMode_Docker && err != nil {
			// docker mode specified explicitly but requirements didn't match
			logger.Error("failed to validate environment: %s", err.Error())
			return err
		}

		switch vmMode {
		case onboard.VMMode_Systemd:
			if err := os.Mkdir(common.SystemdKnoxctlDir, 0755); err != nil && !os.IsExist(err) {
				return err
			}
			logger.SetOut(filepath.Join(common.SystemdKnoxctlDir, "knoxctl.log"))
			logger.Debug("===\n%s - Running %s", time.Now().Format(time.RFC3339), strings.Join(os.Args, " "))
		case onboard.VMMode_Docker:
			// TODO
		}

		vmConfigs, err := onboard.CreateClusterConfig(onboard.ClusterType_VM, userConfigPath, vmMode,
			vmAdapterTag, kubeArmorRelayServerTag, peaVersionTag, siaVersionTag,
			feederVersionTag, sumEngineVersionTag, discoverVersionTag, hardeningAgentVersionTag, kubearmorVersion, releaseVersion, kubeArmorImage,
			kubeArmorInitImage, kubeArmorVMAdapterImage, kubeArmorRelayServerImage, siaImage,
			peaImage, feederImage, rmqImage, sumEngineImage, hardeningAgentImage, spireAgentImage, waitForItImage, discoverImage, nodeAddr, dryRun,
			true, deployRMQ, imagePullPolicy, visibility, hostVisibility, sumengineVisibility, audit, block, hostAudit, hostBlock,
			alertThrottling, maxAlertPerSec, throttleSec,
			cidr, secureContainers, skipBTF, systemMonitorPath, rmqAddress, deploySumegine, registry, registryConfigPath, insecure, plainHTTP, preserveUpstream, topicPrefix, tls, enableHostPolicyDiscovery)
		if err != nil {
			logger.Error("failed to create VM config: %s", err.Error())
			return err
		}
		joinConfig := onboard.JoinClusterConfig(*vmConfigs, kubeArmorAddr, relayServerAddr, siaAddr, peaAddr, hardenAddr)

		err = joinConfig.CreateBaseNodeConfig()
		if err != nil {
			logger.Error("failed to create VM config: %s", err.Error())
			return err
		}

		switch vmMode {

		case onboard.VMMode_Systemd:
			if err := joinConfig.JoinSystemdNode(); err != nil {
				logger.Error("failed to join worker node: %s", err.Error())
				return err
			}

		case onboard.VMMode_Docker:
			err = joinConfig.JoinWorkerNode()
			if err != nil {
				logger.Error("failed to join worker node: %s", err.Error())
				return err
			}

		default:
			logger.Error("vm mode: %s invalid, accepted values (docker/systemd)", vmMode)
			return err
		}

		logger.Print("VM successfully joined with control-plane!===")
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

	joinNodeCmd.PersistentFlags().StringVar(&tls.CaCert, "ca-cert", "", "ca certificate in bas64 encoded format to validate tls connection")

	joinNodeCmd.PersistentFlags().BoolVar(&deploySumegine, "deploy-summary-engine", false, "to deploy summary engine in worker node")

	onboardVMCmd.AddCommand(joinNodeCmd)
}
