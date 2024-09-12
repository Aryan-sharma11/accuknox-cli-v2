package scan

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/accuknox/accuknox-cli-v2/pkg/common"
	kaproto "github.com/kubearmor/KubeArmor/protobuf"
)

// NetworkEvent represents a Network Event
type NetworkEvent struct {
	// Commmand that initiated the network event
	ProcessName string `json:"processName"`

	// IP address
	RemoteIP string `json:"remoteIP,omitempty"`

	// Remote domain name
	RemoteDomain string `json:"remoteDomain,omitempty"`

	// Port
	Port int32 `json:"port,omitempty"`

	// Flow of network event (egress or ingress)
	Flow string `json:"type"`

	// PID of the caller
	PID int32 `json:"pid"`

	// Network protocol
	Protocol string `json:"protocol"`
}

// NetworkCache stores the network events for processing
type NetworkCache struct {
	// Cache holds <pid>: <network-event>
	Cache map[int32][]*NetworkEvent

	// Locks
	mu sync.RWMutex

	// Resolver
	resolver *ConcurrentDNSResolver
}

// NewNetworkCache instantiates the network cache
func NewNetworkCache() *NetworkCache {
	return &NetworkCache{
		Cache:    make(map[int32][]*NetworkEvent),
		resolver: NewResolver(100),
	}
}

// AddNetworkEvent adds a new network event in cache
func (nc *NetworkCache) AddNetworkEvent(log *kaproto.Log) {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	event := &NetworkEvent{
		PID:         log.HostPID,
		ProcessName: getActualProcessName(log.ProcessName),
	}

	event.Flow = extractNetworkFlow(log.Data, log.Resource)
	if event.Flow == "" {
		return
	}

	if strings.Contains(log.Data, "tcp_") {

		nc.handleNetworkEvent(event, log.Resource)
        event.Protocol = "TCP"
	} else if strings.Contains(log.Data, "SYS_BIND") {

		nc.handleNetworkEvent(event, log.Data)
	} else if strings.Contains(log.Data, "SYS_SOCKET") && strings.Contains(log.Resource, "SOCK_DGRAM") {

		event.Flow = "egress"
		event.Protocol = "UDP"
	}

	if event.Protocol == "" {
		return
	}

	nc.Cache[event.PID] = append(nc.Cache[event.PID], event)
}

// StartCachingEvents will cache the network log events
func (nc *NetworkCache) StartCachingEvents(logs []kaproto.Log) {
	for _, log := range logs {
		logCopy := log
		nc.AddNetworkEvent(&logCopy)
	}

	nc.ResolveDomains()
}

// handleTCPEvent handles an event if the data contains tcp
func (nc *NetworkCache) handleNetworkEvent(event *NetworkEvent, data string) {
	resources := strings.Split(data, " ")
	for _, r := range resources {
		parts := strings.SplitN(r, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]
		switch key {
		case "remoteip", "sin_addr":
			event.RemoteIP = val
		case "port", "sin_port":
			if portInt, err := strconv.ParseInt(val, 10, 32); err == nil {
				event.Port = int32(portInt)
			}
		case "protocol", "sa_family":
			event.Protocol = val
		}
	}
}

// Start resolving domains
func (nc *NetworkCache) ResolveDomains() {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	var allEvents []*NetworkEvent
	for _, events := range nc.Cache {
		allEvents = append(allEvents, events...)
	}

	nc.resolver.ResolveConcurrently(allEvents)
}

// SaveNetworkCacheJSON saves the NetworkCache data to a JSON file
func (nc *NetworkCache) SaveNetworkCacheJSON(filename string) error {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	var allEvents []*NetworkEvent
	for _, events := range nc.Cache {
		allEvents = append(allEvents, events...)
	}

	data := struct {
		NetworkEvents []*NetworkEvent `json:"networkEvents"`
	}{
		NetworkEvents: allEvents,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling network cache to JSON: %v", err)
	}

	err = common.CleanAndWrite(filename, jsonData)
	if err != nil {
		return fmt.Errorf("error writing network cache to file: %v", err)
	}

	return nil
}

// GenerateMarkdownTable generates a fancy markdown table of network events
func (nc *NetworkCache) GenerateMarkdownTable() string {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	var sb strings.Builder

	sb.WriteString("| 🔢 PID | 🖥️ Process Name | 🌐 Protocol | 🔄 Flow | 🏠 Remote IP | 🌐 Domain | 🚪 Port |\n")
	sb.WriteString("|--------|-----------------|-------------|---------|--------------|-----------|--------|\n")

	for _, events := range nc.Cache {
		for _, event := range events {
			flowEmoji := "🔼"
			if event.Flow == "ingress" {
				flowEmoji = "🔽"
			}

			domainName := event.RemoteDomain
			if domainName == "" {
				domainName = "N/A"
			}

			sb.WriteString(fmt.Sprintf("| %d | %s | ` %s ` | %s %s | %s | %s | %d |\n",
				event.PID,
				event.ProcessName,
				event.Protocol,
				flowEmoji, event.Flow,
				event.RemoteIP,
				domainName,
				event.Port))
		}
	}

	return sb.String()
}

// SaveNetworkCacheMarkdown saves the NetworkCache data to a Markdown file
func (nc *NetworkCache) SaveNetworkCacheMarkdown(filename string) error {
	markdownContent := nc.GenerateMarkdownTable()

	err := common.CleanAndWrite(filename, []byte(markdownContent))
	if err != nil {
		return fmt.Errorf("error writing network cache to markdown file: %v", err)
	}

	return nil
}
