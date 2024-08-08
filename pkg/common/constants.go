package common

import (
	"path/filepath"
	"runtime"
	"time"
)

const (
	SpecialRegexChars = `.*+?()|[]{}^$`

	ServiceName          = "discovery-engine" // Subject to change
	Port           int64 = 8090
	DELabel              = "app=discovery-engine"
	AccuknoxAgents       = "accuknox-agents"

	APIGroupCilium            = "cilium.io"
	APIGroupKubearmorSecurity = "security.kubearmor.com"
	APIGroupNetworking        = "networking.k8s.io"
	APIGroupRBACAuth          = "rbac.authorization.k8s.io"

	ClusterRole                 = "ClusterRole"
	ServiceAccount              = "ServiceAccount"
	GRPC                        = "grpc"
	AMQP                        = "amqp"
	AMQPPort              int32 = 5672
	GRPCPort              int32 = 8090
	Management                  = "management"
	ManagementPort        int32 = 15672
	CRDName                     = "discoveredpolicies.security.kubearmor.com"
	ConfigMapDirPath            = "pkg/configmaps"
	ClusterRoleViewName         = "dev2-view-cluster-resources"
	ClusterRoleManageName       = "dev2-manage-policies"

	SumEngine          = "summary-engine"
	SumEngineImage     = "accuknox/dev2-sumengine:latest"
	Offlaoder          = "offloader"
	OffloaderImage     = "public.ecr.aws/k9v9d5v2/discovery-engine-offloader:v0.1.0"
	Discover           = "discover"
	DiscoverImage      = "accx3435/dev2-discover:v1"
	Rabbitmq           = "rabbitmq"
	RabbitmqImage      = "rabbitmq:3.12.2-management"
	Hardening          = "hardening"
	HardeningImage     = "accuknox/dev2-hardening:latest"
	ServiceAccountName = "discovery-engine"

	DiscoverConfMap  = "dev2-discover"
	HardeningConfMap = "dev2-hardening"
	OffloaderConfMap = "dev2-offloader"
	SumengineConfmap = "dev2-sumengine"

	AccuknoxGithub  = "accuknox"
	AccuknoxCLIRepo = "accuknox-cli-v2"

	DefaultConfigPathDirName = ".accuknox-config"

	DefaultDockerRegistry = "docker.io"

	// KubeArmor related image/image registries are fixed as of now
	DefaultKubeArmorRepo      = "kubearmor"
	DefaultKubeArmorImage     = "kubearmor/kubearmor"
	DefaultKubeArmorInitImage = "kubearmor/kubearmor-init"
	DefaultRelayServerImage   = "accuknox/kubearmor-relay-server"
	DefaultVMAdapterImage     = "accuknox/vm-adapter"

	DefaultAccuKnoxRepo = "accuknox"
	// Agent images change/have changed over release versions
	// deprecated - do not use
	DefaultPEAImage    = "public.ecr.aws/k9v9d5v2/policy-enforcement-agent:"
	DefaultSIAImage    = "public.ecr.aws/k9v9d5v2/shared-informer-agent:"
	DefaultFeederImage = "public.ecr.aws/k9v9d5v2/feeder-service:"

	MinDockerVersion                  = "v19.0.3"
	MinDockerComposeVersion           = "v1.27.0"
	MinDockerComposeWithWaitSupported = "v2.17.0"

	DownloadDir string = "/tmp/accuknox-downloads/"

	// agents names
	KubeArmor          string = "kubearmor"
	KubeArmorVMAdapter string = "kubearmor-vm-adapter" // to identify service
	VMAdapter          string = "vm-adapter"           // for download package
	RelayServer        string = "kubearmor-relay-server"
	SpireAgent         string = "spire-agent"
	PEAAgent           string = "accuknox-policy-enforcement-agent"
	SIAAgent           string = "accuknox-shared-informer-agent"
	FeederService      string = "accuknox-feeder-service"
	SummaryEngine      string = "accuknox-sumengine"
	DiscoverAgent      string = "accuknox-discover"
	HardeningAgent     string = "accuknox-hardening-agent"

	InContainerConfigDir string = "/opt"

	//config paths for systemd mode
	KAconfigPath             string = "/opt/kubearmor/"
	VmAdapterconfigPath      string = "/opt/kubearmor-vm-adapter/"
	RelayServerconfigPath    string = "/opt/kubearmor-relay-server/"
	FSconfigPath             string = "/opt/accuknox-feeder-service/"
	PEAconfigPath            string = "/opt/accuknox-policy-enforcement-agent/"
	SIAconfigPath            string = "/opt/accuknox-shared-informer-agent/"
	SpireconfigPath          string = "/opt/spire-agent/"
	SumEngineConfigPath      string = "/opt/accuknox-sumengine/"
	DiscoverConfigPath       string = "/opt/accuknox-discover/"
	HardeningAgentConfigPath string = "/opt/accuknox-hardening-agent/"

	//TODO make configurable for policy dir in accuknox policy enforcementagent
	PeaPolicyPath string = "/opt/pea/"

	//systemd path
	SystemdDir         string = "/usr/lib/systemd/system/"
	KmuxConfigFileName string = "kmux-config.yaml"

	// KubeArmor gRPC service port
	KubeArmorGRPCAddress string = "localhost:32767"

	// Events
	OperationNetwork = "Network"
	OperationFile    = "File"
	OperationProcess = "Process"
)

var (
	MatchLabels = map[string]string{"app": "discovery-engine"}
	// TODO: Add action and few other coloumns in network
	// SysProcHeader variable contains source process, destination process path, count, timestamp and status
	SysProcHeader = []string{"Src Process", "Destination Process Path", "Count", "Last Updated Time"}
	// SysFileHeader variable contains source process, destination file path, count, timestamp and status
	SysFileHeader = []string{"Src Process", "Destination File Path", "Count", "Last Updated Time"}
	// SysNwHeader variable contains protocol, command, POD/SVC/IP, Port, Namespace, and Labels
	SysNwHeader = []string{"Protocol", "Command", "POD/SVC/IP", "Port", "Count", "Last Updated Time"}
	// SysBindNwHeader variable contains protocol, command, Bind Port, Bind Address, count and timestamp
	SysBindNwHeader = []string{"Protocol", "Command", "Bind Port", "Bind Address", "Count", "Last Updated Time"}

	SystemdTagSuffix = "_" + runtime.GOOS + "-" + runtime.GOARCH

	AgentRepos = map[string]string{
		KubeArmor:      "kubearmor/kubearmor-systemd",
		VMAdapter:      "accuknox/vm-adapter-systemd",
		RelayServer:    "accuknox/kubearmor-relay-server-systemd",
		PEAAgent:       "accuknox/accuknox-policy-enforcement-agent-systemd",
		SIAAgent:       "accuknox/accuknox-shared-informer-agent-systemd",
		FeederService:  "accuknox/accuknox-feeder-service-systemd",
		SpireAgent:     "accuknox/spire-agent-systemd",
		SummaryEngine:  "accuknox/accuknox-sumengine-systemd",
		DiscoverAgent:  "accuknox/accuknox-discover-systemd",
		HardeningAgent: "accuknox/accuknox-hardening-agent-systemd",
	}

	KASystemMonitorPath string = filepath.Join(KAconfigPath, "BPF", "system_monitor.bpf.o")
)

// Timeoutes
var (
	// Sets duration of 10 seconds
	TenSeconds = time.Duration(10 * time.Second)

	// Sets duartion of 30 seconds
	ThirtySeconds = time.Duration(60 * time.Second)

	// Sets duration of 1 minute (60 seconds)
	OneMinute = time.Duration(60 * time.Second)

	// Sets duration of 5 minutes (300 seconds)
	FiveMinutes = time.Duration(5 * time.Minute)
)
