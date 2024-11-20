package onboard

import (
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/accuknox/accuknox-cli-v2/pkg/common"
	"golang.org/x/mod/semver"
)

func JoinClusterConfig(cc ClusterConfig, kubeArmorAddr, relayServerAddr, siaAddr, peaAddr, hardenAddr string) *JoinConfig {
	return &JoinConfig{
		ClusterConfig:   cc,
		KubeArmorAddr:   kubeArmorAddr,
		RelayServerAddr: relayServerAddr,
		SIAAddr:         siaAddr,
		PEAAddr:         peaAddr,
		HardenAddr:      hardenAddr,
	}
}
func (jc *JoinConfig) CreateBaseNodeConfig() error {

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	kubeArmorURL := "localhost:32767"
	kubeArmorAddr := ""
	kubeArmorPort := "32767"
	if jc.KubeArmorAddr != "" {
		kubeArmorURL = jc.KubeArmorAddr
	}
	kubeArmorAddr, kubeArmorPort, err = parseURL(kubeArmorURL)
	if err != nil {
		return err
	}

	// parse URL and assign default values as needed
	var relayHost, relayPort, relayAddr string
	if jc.RelayServerAddr != "" {
		relayAddr = jc.RelayServerAddr
		relayHost, relayPort, err = parseURL(jc.RelayServerAddr)
		if err != nil {
			return err
		}
	} else if jc.CPNodeAddr != "" {
		relayHost = jc.CPNodeAddr
		relayPort = "32768"
		relayAddr = jc.CPNodeAddr + ":" + relayPort
	} else {
		return fmt.Errorf("Relay server address cannot be empty")
	}

	var siaAddr string
	if jc.SIAAddr != "" {
		siaAddr = jc.SIAAddr
	} else if siaAddr == "" && jc.CPNodeAddr != "" {
		siaAddr = jc.CPNodeAddr + ":" + "32769"
	} else {
		return fmt.Errorf("SIA address cannot be empty")
	}

	var peaAddr string
	if jc.PEAAddr != "" {
		peaAddr = jc.PEAAddr
	} else if peaAddr == "" && jc.CPNodeAddr != "" {
		peaAddr = jc.CPNodeAddr + ":" + "32770"
	} else {
		return fmt.Errorf("PEA address cannot be empty")
	}

	var hardenAddr string
	if jc.HardenAddr != "" {
		hardenAddr = jc.HardenAddr
	} else if hardenAddr == "" && jc.CPNodeAddr != "" {
		hardenAddr = jc.CPNodeAddr + ":" + "32771"
	} else {
		return fmt.Errorf("Hardening Agent address cannot be empty")
	}

	// RMQServer that would be used by summary engine
	if jc.DeploySumengine {
		if jc.RMQServer == "" && jc.CPNodeAddr != "" {
			cpNodeServerAddr, cpNodePort, err := parseURL(jc.CPNodeAddr)
			if err != nil {
				return err
			}

			if cpNodePort != "" {
				jc.RMQServer = cpNodeServerAddr
			} else {
				jc.RMQServer = cpNodeServerAddr + ":" + "5672"
			}

		} else {
			return fmt.Errorf("RMQ address cannot be empty")
		}
	}

	jc.TCArgs = TemplateConfigArgs{
		Hostname: hostname,

		// for vm-adapter
		KubeArmorAddr: kubeArmorAddr,
		KubeArmorURL:  kubeArmorURL,
		KubeArmorPort: kubeArmorPort,

		RelayServerURL:  relayAddr,
		RelayServerAddr: relayHost,
		RelayServerPort: relayPort,

		SIAAddr:    siaAddr,
		PEAAddr:    peaAddr,
		HardenAddr: hardenAddr,

		WorkerNode: jc.WorkerNode,
		// kubearmor config
		KubeArmorVisibility:     jc.Visibility,
		KubeArmorHostVisibility: jc.HostVisibility,

		KubeArmorFilePosture:    jc.DefaultFilePosture,
		KubeArmorNetworkPosture: jc.DefaultNetworkPosture,
		KubeArmorCapPosture:     jc.DefaultCapPosture,

		KubeArmorAlertThrottling: jc.AlertThrottling,
		KubeArmorMaxAlertsPerSec: jc.MaxAlertsPerSec,
		KubeArmorThrottleSec:     jc.ThrottleSec,

		KubeArmorHostFilePosture:    jc.DefaultHostFilePosture,
		KubeArmorHostNetworkPosture: jc.DefaultHostNetworkPosture,
		KubeArmorHostCapPosture:     jc.DefaultHostCapPosture,
		NetworkCIDR:                 jc.CIDR,
		VmMode:                      jc.Mode,
		SecureContainers:            jc.SecureContainers,
		TlsEnabled:                  jc.Tls.Enabled,
		PoliciesTopic:               getTopicName(jc.RMQTopicPrefix, "policies"),
		LogsTopic:                   getTopicName(jc.RMQTopicPrefix, "logs"),
		AlertsTopic:                 getTopicName(jc.RMQTopicPrefix, "alerts"),
		StateEventTopic:             getTopicName(jc.RMQTopicPrefix, "state-event"),
		PolicyV1Topic:               getTopicName(jc.RMQTopicPrefix, "policy-v1"),
		SummaryV2Topic:              getTopicName(jc.RMQTopicPrefix, "summary-v2"),

		EnableHostPolicyDiscovery: jc.EnableHostPolicyDiscovery,
	}

	jc.TCArgs.PoliciesKmuxConfig = common.KmuxPoliciesFileName
	jc.TCArgs.StateKmuxConfig = common.KmuxStateEventFileName
	jc.TCArgs.AlertsKmuxConfig = common.KmuxAlertsFileName
	jc.TCArgs.LogsKmuxConfig = common.KmuxLogsFileName
	jc.TCArgs.SummaryKmuxConfig = common.KmuxSummaryFileName
	jc.TCArgs.PolicyKmuxConfig = common.KmuxPolicyFileName

	return nil
}

func (jc *JoinConfig) JoinWorkerNode() error {
	// validate this environment
	dockerStatus, err := jc.ValidateEnv()
	if err != nil {
		return err
	}
	fmt.Println(dockerStatus)

	configPath, err := createDefaultConfigPath()
	if err != nil {
		return err
	}
	// configs specific to docker mode of installation

	jc.TCArgs.KubeArmorImage = jc.KubeArmorImage
	jc.TCArgs.KubeArmorInitImage = jc.KubeArmorInitImage
	jc.TCArgs.KubeArmorVMAdapterImage = jc.KubeArmorVMAdapterImage
	jc.TCArgs.ImagePullPolicy = string(jc.ImagePullPolicy)
	jc.TCArgs.ConfigPath = configPath
	jc.TCArgs.SumEngineImage = jc.SumEngineImage
	jc.TCArgs.TlsEnabled = jc.Tls.Enabled

	if jc.Tls.RMQCredentials != "" {
		rmqData := strings.Split(Decode(jc.Tls.RMQCredentials), ":")
		if len(rmqData) != 2 {
			return fmt.Errorf("invalid RMQ credentials")
		}
		jc.TCArgs.RMQUsername = rmqData[0]
		jc.TCArgs.RMQPassword = rmqData[1]
	}

	// initialize sprig for templating
	sprigFuncs := sprig.GenericFuncMap()

	if jc.Tls.Enabled {
		jc.TCArgs.TlsCertFile = fmt.Sprintf("%s%s%s/%s", jc.UserConfigPath, configPath, common.DefaultCACertDir, common.DefaultEncodedFileName)
		caPath := configPath + "/cert/encoded.pem"
		if err := StoreCert(map[string]string{
			caPath: jc.Tls.CaCert,
		}); err != nil {
			return err
		}
	}

	// write compose file
	composeFilePath, err := copyOrGenerateFile(jc.UserConfigPath, configPath, "docker-compose.yaml", sprigFuncs, workerNodeComposeFileTemplate, jc.TCArgs)
	if err != nil {
		return err
	}

	kmuxConfigArgs := KmuxConfigTemplateArgs{
		ReleaseVersion: jc.AgentsVersion,
		RMQServer:      jc.RMQServer,
		RMQUsername:    jc.TCArgs.RMQUsername,
		RMQPassword:    jc.TCArgs.RMQPassword,
		TlsEnabled:     jc.TCArgs.TlsEnabled,
		TlsCertFile:    jc.TCArgs.TlsCertFile,
	}

	populateAgentArgs(&jc.TCArgs, "sumengine")
	if _, err := copyOrGenerateFile(jc.UserConfigPath, configPath, "sumengine/config.yaml", sprigFuncs, sumEngineConfig, jc.TCArgs); err != nil {
		return err
	}

	kmuxConfigFileTemplateMap := map[string]string{
		"sumengine/" + common.KmuxConfigFileName:      kmuxPublisherConfig,
		"sumengine/" + common.KmuxSummaryFileName:     kmuxPublisherConfig,
		"vm-adapter/" + common.KmuxStateEventFileName: kmuxPublisherConfig,
		"vm-adapter/" + common.KmuxAlertsFileName:     kmuxPublisherConfig,
		"vm-adapter/" + common.KmuxLogsFileName:       kmuxPublisherConfig,
		"vm-adapter/" + common.KmuxPoliciesFileName:   kmuxConsumerConfig,
	}
	// Generate or copy kmux config files
	for filePath, templateString := range kmuxConfigFileTemplateMap {
		agentName, file := strings.Split(filePath, "/")[0], strings.Split(filePath, "/")[1]
		populateKmuxArgs(&kmuxConfigArgs, agentName, file, jc.RMQTopicPrefix, jc.TCArgs.Hostname)
		if _, err := copyOrGenerateFile(jc.UserConfigPath, configPath, filePath, sprigFuncs, templateString, kmuxConfigArgs); err != nil {
			return err
		}
	}

	diagnosis := true

	args := []string{"-f", composeFilePath, "--profile", "kubearmor-only"}
	if jc.DeploySumengine {
		args = append(args, "--profile", "accuknox-agents")
	}
	args = append(args, "up", "-d")

	// need these flags for diagnosis
	if semver.Compare(jc.composeVersion, common.MinDockerComposeWithWaitSupported) >= 0 {
		args = append(args, "--wait", "--wait-timeout", "60")
	} else {
		diagnosis = false
	}

	// run compose command
	_, err = ExecComposeCommand(true, jc.DryRun, jc.composeCmd, args...)
	if err != nil {
		// cleanup volumes
		_, volDelErr := ExecDockerCommand(true, false, "docker", "volume", "rm", "kubearmor-init-vol")
		if volDelErr != nil {
			fmt.Println("Error while removing volumes:", volDelErr.Error())
		}

		if diagnosis {
			diagnosisResult, diagErr := diagnose(NodeType_WorkerNode)
			if diagErr != nil {
				diagnosisResult = diagErr.Error()
			}
			return fmt.Errorf("Error: %s.\n\nDIAGNOSIS:\n%s", err.Error(), diagnosisResult)
		}

		return err
	}

	return nil
}
