package onboard

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/accuknox/accuknox-cli-v2/pkg/common"
)

func CreateClusterConfig(clusterType ClusterType, userConfigPath string, vmMode VMMode,
	vmAdapterTag, kubeArmorRelayServerTag, peaVersionTag, siaVersionTag, feederVersionTag, sumEngineTag, discoverVersionTag, hardeningAgentVersionTag string,
	kubearmorVersion, releaseVersion, kubearmorImage, kubearmorInitImage,
	vmAdapterImage, relayServerImage, siaImage, peaImage,
	feederImage, sumEngineImage, hardeningAgentImage, spireImage, discoverImage, nodeAddress string, dryRun, workerNode bool,
	imagePullPolicy, visibility, hostVisibility, audit,
	block, cidr string, secureContainers bool) (*ClusterConfig, error) {

	cc := new(ClusterConfig)

	// check if a config path is given by user
	if userConfigPath != "" {
		cleanUserConfigPath := filepath.Clean(userConfigPath)
		_, err := os.Stat(userConfigPath)
		if err != nil {
			return nil, err
		}

		defaultConfigPath, err := common.GetDefaultConfigPath()
		if err != nil {
			return nil, err
		}

		// userConfigPath can't be defaultConfigPath
		if strings.Compare(defaultConfigPath, cleanUserConfigPath) == 0 {
			return nil, fmt.Errorf("config path cannot be %s", defaultConfigPath)
		}

		cc.UserConfigPath = cleanUserConfigPath
	}
	// systemd or docker
	cc.Mode = vmMode

	cc.SecureContainers = secureContainers

	cc.ClusterType = clusterType

	cc.Visibility = visibility
	if cc.Visibility == "" {
		cc.Visibility = "process,network"
	}

	cc.HostVisibility = hostVisibility
	if cc.HostVisibility == "" {
		cc.HostVisibility = "process,network"
	}

	// if audit or no default posture specified
	cc.DefaultFilePosture = getDefaultPosture(audit, block, "file")
	cc.DefaultNetworkPosture = getDefaultPosture(audit, block, "network")
	cc.DefaultCapPosture = getDefaultPosture(audit, block, "capabilities")

	cc.DefaultHostFilePosture = getDefaultPosture(audit, block, "file")
	cc.DefaultHostNetworkPosture = getDefaultPosture(audit, block, "network")
	cc.DefaultHostCapPosture = getDefaultPosture(audit, block, "capabilities")

	var releaseInfo common.ReleaseMetadata
	if releaseVersion == "" {
		_, releaseInfo = common.GetLatestReleaseInfo()
	} else if releaseInfoTemp, ok := common.ReleaseInfo[releaseVersion]; ok {
		releaseInfo = releaseInfoTemp
	} else {
		// TODO: publish release JSON as OCI artifact to remove dependency
		// on needing to build knoxctl again and again
		return nil, fmt.Errorf("Unknown image tag %s", releaseVersion)
	}
	cc.AgentsVersion = releaseVersion
	switch cc.Mode {

	case VMMode_Docker:

		switch strings.ToLower(imagePullPolicy) {
		case string(ImagePullPolicy_Always):
			cc.ImagePullPolicy = ImagePullPolicy_Always
		case string(ImagePullPolicy_IfNotPresent):
			cc.ImagePullPolicy = ImagePullPolicy_IfNotPresent
		case string(ImagePullPolicy_Never):
			cc.ImagePullPolicy = ImagePullPolicy_Never
		default:
			return nil, fmt.Errorf("Image pull policy %s unrecognized", imagePullPolicy)
		}

		if kubearmorImage != "" {
			cc.KubeArmorImage = kubearmorImage
		} else if kubearmorVersion != "" {
			cc.KubeArmorImage = common.DefaultKubeArmorImage + kubearmorVersion
		} else {
			cc.KubeArmorImage = common.DefaultKubeArmorImage + releaseInfo.KubeArmorTag
		}

		if kubearmorInitImage != "" {
			cc.KubeArmorInitImage = kubearmorInitImage
		} else if kubearmorVersion != "" {
			cc.KubeArmorInitImage = common.DefaultKubeArmorInitImage + kubearmorVersion
		} else {
			cc.KubeArmorInitImage = common.DefaultKubeArmorInitImage + releaseInfo.KubeArmorTag
		}

		if relayServerImage != "" {
			cc.KubeArmorRelayServerImage = relayServerImage
		} else {
			cc.KubeArmorRelayServerImage = common.DefaultRelayServerImage + releaseInfo.KubeArmorRelayTag
		}

		if vmAdapterImage != "" {
			cc.KubeArmorVMAdapterImage = vmAdapterImage
		} else {
			cc.KubeArmorVMAdapterImage = common.DefaultVMAdapterImage + releaseInfo.KubeArmorVMAdapterTag
		}

	case VMMode_Systemd:
		if kubearmorVersion == "" {
			cc.KubeArmorTag = releaseInfo.KubeArmorTag
		} else if kubearmorVersion == "stable" || kubearmorVersion == "latest" {
			log.Printf("%s tag not available for systemd package using values from release chart", kubearmorVersion)
			cc.KubeArmorTag = releaseInfo.KubeArmorTag
		} else {
			if !strings.HasPrefix(kubearmorVersion, "v") {
				kubearmorVersion = "v" + kubearmorVersion
			}
			cc.KubeArmorTag = kubearmorVersion
		}
		cc.VmAdapterTag = GetSystemdTag(vmAdapterTag, releaseInfo.KubeArmorVMAdapterTag)
		cc.RelayServerTag = GetSystemdTag(kubeArmorRelayServerTag, releaseInfo.KubeArmorRelayTag)

	}

	cc.WorkerNode = workerNode
	cc.DryRun = dryRun

	cc.CPNodeAddr = nodeAddress
	if cc.CPNodeAddr == "" {
		cc.CPNodeAddr = "<address-of-this-node>"
	}

	if cidr != "" {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}

		cc.CIDR = network.String()
	}

	if workerNode {
		return cc, nil
	}
	switch cc.Mode {
	case VMMode_Docker:

		if siaImage != "" {
			cc.SIAImage = siaImage
		} else if releaseVersion != "" {
			cc.SIAImage = releaseInfo.SIAImage + ":" + releaseInfo.SIATag
		} else {
			return nil, fmt.Errorf("No tag found for SIA")
		}

		if peaImage != "" {
			cc.PEAImage = peaImage
		} else if releaseVersion != "" {
			cc.PEAImage = releaseInfo.PEAImage + ":" + releaseInfo.PEATag
		} else {
			return nil, fmt.Errorf("No tag found for PEA")
		}

		if feederImage != "" {
			cc.FeederImage = feederImage
		} else if releaseVersion != "" {
			cc.FeederImage = releaseInfo.FeederServiceImage + ":" + releaseInfo.FeederServiceTag
		} else {
			return nil, fmt.Errorf("No tag found for feeder-service")
		}
		if spireImage != "" {
			cc.SPIREAgentImage = spireImage
		} else if releaseVersion != "" {
			cc.SPIREAgentImage = "public.ecr.aws/k9v9d5v2/spire-agent:latest"
			// TODO: once the image is pushed to dockerhub
			//cc.SPIREAgentImage = "accuknox/spire-agent" + ":" + releaseInfo.SPIREAgentImageTag
		} else {
			return nil, fmt.Errorf("No tag found for spire-agent")
		}
		if discoverImage != "" {
			cc.DiscoverImage = discoverImage
		} else if releaseVersion != "" {
			cc.DiscoverImage = releaseInfo.DiscoverImage + ":" + releaseInfo.DiscoverTag
		} else {
			return nil, fmt.Errorf("No tag found for discover")
		}
		if sumEngineImage != "" {
			cc.SumEngineImage = sumEngineImage
		} else if releaseVersion != "" {
			cc.SumEngineImage = releaseInfo.SumEngineImage + ":" + releaseInfo.SumEngineTag
		} else {
			return nil, fmt.Errorf("No tag found for summary-engine")
		}
		if hardeningAgentImage != "" {
			cc.HardeningAgentImage = hardeningAgentImage
		} else if releaseVersion != "" {
			cc.HardeningAgentImage = releaseInfo.HardeningAgentImage + ":" + releaseInfo.HardeningAgentTag
		} else {
			return nil, fmt.Errorf("No tag found for hardening-agent")
		}

	case VMMode_Systemd:
		cc.PeaTag = GetSystemdTag(peaVersionTag, releaseInfo.PEATag)
		cc.SiaTag = GetSystemdTag(siaVersionTag, releaseInfo.SIATag)
		cc.FsTag = GetSystemdTag(feederVersionTag, releaseInfo.FeederServiceTag)
		cc.SpireTag = GetSystemdTag("", releaseInfo.SPIREAgentImageTag)
		cc.SumEngineTag = GetSystemdTag("", releaseInfo.SumEngineTag)
		cc.DiscoverTag = GetSystemdTag("", releaseInfo.DiscoverTag)
		cc.HardeningAgentTag = GetSystemdTag("", releaseInfo.HardeningAgentTag)
	}

	return cc, nil
}

// prints join command - currently only with the default ports
// TODO: handle complex configuration
func (cc *ClusterConfig) PrintJoinCommand(vmmode VMMode) {
	command := ""
	switch vmmode {
	case VMMode_Docker:
		command = fmt.Sprintf("knoxctl onboard vm node --vm-mode=\"docker\" --cp-node-addr=%s", cc.CPNodeAddr)

	case VMMode_Systemd:
		command = fmt.Sprintf("knoxctl onboard vm node --vm-mode=\"systemd\" --cp-node-addr=%s", cc.CPNodeAddr)
	}

	fmt.Println(command)
}

func getDefaultPosture(auditPostureVal, blockPostureVal, ruleType string) string {
	if auditPostureVal == "all" || (auditPostureVal == "" && blockPostureVal == "") {
		return "audit"
	} else if blockPostureVal == "all" {
		return "block"
	}

	if strings.Contains(auditPostureVal, ruleType) {
		return "audit"
	} else if strings.Contains(blockPostureVal, ruleType) {
		return "block"
	}

	// unrecognized or default
	return "audit"
}
