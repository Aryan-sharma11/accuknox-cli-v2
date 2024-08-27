package onboard

import (
	_ "embed"
)

var (
	//go:embed templates/kmux-config.yaml
	kmuxConfig string

	//go:embed templates/discover-kmux-config.yaml
	discoverKmuxConfig string

	//go:embed templates/sumengine-kmux-config.yaml
	sumEngineKmuxConfig string

	//go:embed templates/hardening-agent-kmux-config.yaml
	hardeningAgentKmuxConfig string

	//go:embed templates/pea-config.yaml
	peaConfig string

	//go:embed templates/pea-rmq-kmux-config.yaml
	peaRmqKmuxConfig string

	//go:embed templates/sia-config.yaml
	siaConfig string

	//go:embed templates/discover-config.yaml
	discoverConfig string

	//go:embed templates/sumengine-config.yaml
	sumEngineConfig string

	//go:embed templates/hardening-agent-config.yaml
	hardeningAgentConfig string

	//go:embed templates/systemdTemplates/kubearmor-config.yaml
	kubeArmorConfig string

	//go:embed templates/systemdTemplates/vm-adapter-config.yaml
	vmAdapterConfig string

	//go:embed templates/spire-agent.conf
	spireAgentConfig string

	//go:embed templates/rabbitmq.conf
	rabbitmqConfig string

	//go:embed templates/definitions.json
	rabbitmqDefinitions string

	//go:embed templates/systemdTemplates/feeder-service-env
	fsEnvVal string

	spireTrustBundleURLMap = map[string]string{
		"dev":     "https://accuknox-dev-cert-spire.s3.us-east-2.amazonaws.com/ca.crt",
		"stage":   "https://accuknox-stage-cert-spire.s3.us-east-2.amazonaws.com/ca.crt",
		"demo":    "https://accuknox-demo-cert-spire.s3.us-east-2.amazonaws.com/ca.crt",
		"prod":    "https://accuknox-prod-cert-spire.s3.us-east-2.amazonaws.com/ca.crt",
		"xcitium": "https://accuknox-spire.s3.amazonaws.com/certs/xcitium/certificate.crt",
	}
)
