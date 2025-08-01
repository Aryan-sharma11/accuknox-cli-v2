package discover

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/accuknox/accuknox-cli-v2/pkg/common"
	"github.com/kubearmor/kubearmor-client/k8s"
	"github.com/schollz/progressbar/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sigs.k8s.io/yaml"

	dev2policy "github.com/accuknox/dev2/api/grpc/v1/policy"
	policyType "github.com/accuknox/dev2/discover/pkg/common"
	networkingv1 "k8s.io/api/networking/v1"
)

// Global variable for the gRPC connection
var connection *grpc.ClientConn

func initConnection(c *k8s.Client, p *Options) error {
	var err error
	gRPC, err := common.ConnectGrpc(c, p.GRPC)
	if err != nil {
		return err
	}
	connection, err = grpc.Dial(gRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	return nil
}

func disconnect() {
	if connection != nil {
		err := connection.Close()
		if err != nil {
			fmt.Println("Failed to close connection")
		}
	}
}

func getNetworkPolicy(c *k8s.Client, p *Options, pf *PolicyForest) error {
	client := dev2policy.NewGetPolicyClient(connection)
	resp, err := client.GetPolicy(context.Background(), &dev2policy.PolicyRequest{
		Type: PolicyType,           // discovered
		Kind: KindK8sNetworkPolicy, // NetworkPolicy
	})
	if err != nil {
		return fmt.Errorf("failed to fetch response from discovery engine: %v", err)
	}

	if resp != nil {
		bar := initializeProgressBar(len(resp.Policies))

		errorChan := make(chan error, len(resp.Policies))
		var wg sync.WaitGroup

		for _, policy := range resp.Policies {
			wg.Add(1)

			go func() {
				for err := range errorChan {
					fmt.Println(err)
				}
			}()

			go func(policy *dev2policy.Policy) {
				defer wg.Done()

				var networkPolicy networkingv1.NetworkPolicy
				err := yaml.Unmarshal(policy.Yaml, &networkPolicy)
				if err != nil {
					errorChan <- err
				}

				if !networkPolicyFilter(networkPolicy, p) {
					return
				}

				pf.Lock()
				pf.AddNetworkPolicy(networkPolicy.ObjectMeta.Namespace, &networkPolicy)
				pf.Unlock()

				_ = bar.Add(1)
			}(policy)
		}
		wg.Wait()
		close(errorChan)

		err := bar.Finish()
		if err != nil {
			fmt.Println("Failed to finish progress bar")
		}
	}

	return nil
}

func getKaHostPolicy(c *k8s.Client, p *Options, pf *PolicyForest) error {
	client := dev2policy.NewGetPolicyClient(connection)
	resp, err := client.GetPolicy(context.Background(), &dev2policy.PolicyRequest{
		Type: PolicyType,              // discovered
		Kind: KindKubeArmorHostPolicy, // KAHostPolicy
	})
	if err != nil {
		return fmt.Errorf("failed to fetch response from discovery engine: %v", err)
	}

	if resp != nil {
		bar := initializeProgressBar(len(resp.Policies))

		errorChan := make(chan error, len(resp.Policies))
		go func() {
			for err := range errorChan {
				fmt.Println(err)
			}
		}()

		var wg sync.WaitGroup
		for _, policy := range resp.Policies {
			wg.Add(1)

			go func(policy *dev2policy.Policy) {
				defer wg.Done()

				var kaHostPolicy policyType.KubeArmorPolicy
				err := yaml.Unmarshal(policy.Yaml, &kaHostPolicy)
				if err != nil {
					errorChan <- err
				}

				if !kaPolicyFilter(kaHostPolicy, p) {
					return
				}

				pf.Lock()
				pf.AddKubearmorPolicy(kaHostPolicy.Metadata.Namespace, &kaHostPolicy)
				pf.Unlock()

				_ = bar.Add(1)
			}(policy)
		}
		wg.Wait()
		close(errorChan)

		err := bar.Finish()
		if err != nil {
			fmt.Println("Failed to finish progress bar")
		}
	}

	return nil
}

func getKaPolicy(c *k8s.Client, p *Options, pf *PolicyForest) error {
	client := dev2policy.NewGetPolicyClient(connection)
	resp, err := client.GetPolicy(context.Background(), &dev2policy.PolicyRequest{
		Type: PolicyType,          // discovered
		Kind: KindKubeArmorPolicy, // KAPolicy
	})
	if err != nil {
		return fmt.Errorf("failed to fetch response from discovery engine: %v", err)
	}

	if resp != nil {
		bar := initializeProgressBar(len(resp.Policies))

		errorChan := make(chan error, len(resp.Policies))
		go func() {
			for err := range errorChan {
				fmt.Println(err)
			}
		}()

		var wg sync.WaitGroup
		for _, policy := range resp.Policies {
			wg.Add(1)

			go func(policy *dev2policy.Policy) {
				defer wg.Done()

				var kaPolicy policyType.KubeArmorPolicy
				err := yaml.Unmarshal(policy.Yaml, &kaPolicy)
				if err != nil {
					errorChan <- err
				}

				if !kaPolicyFilter(kaPolicy, p) {
					return
				}

				pf.Lock()
				pf.AddKubearmorPolicy(kaPolicy.Metadata.Namespace, &kaPolicy)
				pf.Unlock()

				_ = bar.Add(1)
			}(policy)
		}
		wg.Wait()
		err := bar.Finish()
		if err != nil {
			fmt.Println("Failed to finish progress bar")
		}
	}

	return nil
}

// Centralized filteration based on user options
// OR based filter at flag level and AND based filter at command level.
func kaPolicyFilter(policy policyType.KubeArmorPolicy, p *Options) bool {
	// If no filters are provided, return true
	if p.noFilters() {
		return true
	}

	namespaceMatched := len(p.Namespace) == 0 && p.NamespaceRegex == nil
	labelMatched := len(p.Labels) == 0 && p.LabelsRegex == nil
	sourceMatched := len(p.Source) == 0 && p.SourceRegex == nil
	includeNetworkMatched := !p.IncludeNetwork || len(policy.Spec.Network.MatchProtocols) > 0

	// Namespace filtering
	if !namespaceMatched {
		for _, ns := range p.Namespace {
			if policy.Metadata.Namespace == ns {
				namespaceMatched = true
				break
			}
		}
		if p.NamespaceRegex != nil && !namespaceMatched {
			for _, regex := range p.NamespaceRegex {
				if regex.MatchString(policy.Metadata.Namespace) {
					namespaceMatched = true
					break
				}
			}
		}
	}

	// Labels filtering
	if !labelMatched {
		for _, label := range p.Labels {
			keyVal := strings.Split(label, "=")
			if len(keyVal) == 2 {
				if policyValue, exists := policy.Spec.Selector.MatchLabels[keyVal[0]]; exists && policyValue == keyVal[1] {
					labelMatched = true
					break
				}
			}
		}
		if p.LabelsRegex != nil && !labelMatched {
			for _, regex := range p.LabelsRegex {
				for k, v := range policy.Spec.Selector.MatchLabels {
					if regex.MatchString(k + "=" + v) {
						labelMatched = true
						break
					}
				}
				if labelMatched {
					break
				}
			}
		}
	}

	// FromSource filtering
	if !sourceMatched {
		for _, path := range policy.Spec.Process.MatchPaths {
			for _, fromSource := range path.FromSource {
				for _, src := range p.Source {
					if fromSource.Path == src || fromSource.Dir == src {
						sourceMatched = true
						break
					}
				}
				if sourceMatched {
					break
				}
				if p.SourceRegex != nil {
					for _, regex := range p.SourceRegex {
						if regex.MatchString(fromSource.Path) || regex.MatchString(fromSource.Dir) {
							sourceMatched = true
							break
						}
					}
				}
				if sourceMatched {
					break
				}
			}
			if sourceMatched {
				break
			}
		}
		if !sourceMatched {
			for _, dir := range policy.Spec.File.MatchDirectories {
				for _, fromSource := range dir.FromSource {
					for _, src := range p.Source {
						if fromSource.Path == src || fromSource.Dir == src {
							sourceMatched = true
							break
						}
					}
					if sourceMatched {
						break
					}
					if p.SourceRegex != nil {
						for _, regex := range p.SourceRegex {
							if regex.MatchString(fromSource.Path) || regex.MatchString(fromSource.Dir) {
								sourceMatched = true
								break
							}
						}
					}
					if sourceMatched {
						break
					}
				}
				if sourceMatched {
					break
				}
			}
		}
	}

	return namespaceMatched && labelMatched && sourceMatched && includeNetworkMatched
}

func networkPolicyFilter(policy networkingv1.NetworkPolicy, p *Options) bool {
	if p.noFilters() {
		return true
	}

	labelMatched := len(p.Labels) == 0 && p.LabelsRegex == nil
	namespaceMatched := len(p.Namespace) == 0 && p.NamespaceRegex == nil

	// Metadata Label filtering
	if !labelMatched {
		for _, label := range p.Labels {
			keyVal := strings.Split(label, "=")
			if len(keyVal) == 2 {
				if policyValue, exists := policy.Labels[keyVal[0]]; exists && policyValue == keyVal[1] {
					labelMatched = true
					break
				}
			}
		}
		if p.LabelsRegex != nil && !labelMatched {
			for _, regex := range p.LabelsRegex {
				for key, value := range policy.Labels {
					if regex.MatchString(key + "=" + value) {
						labelMatched = true
						break
					}
				}
				if labelMatched {
					break
				}
			}
		}
	}

	// Namespace filtering from ObjectMeta
	if !namespaceMatched {
		for _, ns := range p.Namespace {
			if policy.ObjectMeta.Namespace == ns {
				namespaceMatched = true
				break
			}
		}
		if p.NamespaceRegex != nil && !namespaceMatched {
			for _, regex := range p.NamespaceRegex {
				if regex.MatchString(policy.ObjectMeta.Namespace) {
					namespaceMatched = true
					break
				}
			}
		}
	}

	// If both criteria are met
	return labelMatched && namespaceMatched
}

func initializeProgressBar(totalCount int) *progressbar.ProgressBar {
	bar := progressbar.NewOptions(
		totalCount,
		progressbar.OptionSetDescription("Processing policies..."),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSpinnerType(9),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetElapsedTime(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowBytes(false),
		progressbar.OptionShowIts(),
	)
	return bar
}
