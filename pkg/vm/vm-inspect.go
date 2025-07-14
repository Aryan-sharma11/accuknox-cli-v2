package vm

import (
	"sync"

	"github.com/accuknox/accuknox-cli-v2/pkg/logger"
	"github.com/accuknox/accuknox-cli-v2/pkg/onboard"
)

func InspectVM() error {

	var (
		kaCompatible      *NodeInfo
		portsAvailability map[string]string
		installedAgents   map[string]string
		vmMode            onboard.VMMode
		nodeType          onboard.NodeType
		wg                sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		kaCompatible = kubeArmorCompatibility()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ports, err := checkPorts()
		if err != nil {
			logger.Error("error checking ports", err)
			return
		}
		portsAvailability = ports
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		agents, vm, node := getInstalledAgents()
		installedAgents = agents
		vmMode = vm
		nodeType = node
	}()

	wg.Wait()

	if vmMode != "" && nodeType != "" && kaCompatible != nil {
		kaCompatible.NodeType = nodeType
		kaCompatible.VmMode = vmMode
	}

	printNodeData(*kaCompatible)
	printPortData(portsAvailability)
	if len(installedAgents) > 0 {
		printAgentsData(installedAgents)
	}

	return nil

}
