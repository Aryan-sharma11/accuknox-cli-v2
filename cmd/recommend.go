// SPDX-License-Identifier: Apache-2.0
// Copyright 2022 Authors of KubeArmor

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/accuknox/accuknox-cli-v2/pkg/recommend"
	"github.com/spf13/cobra"
)

var recommendOptions recommend.Options

// recommendCmd represents the recommend command
var recommendCmd = &cobra.Command{
	Use:   "recommend",
	Short: "Recommend Policies",
	Long:  `Get recommended hardening policies generated by Discovery Engine`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rawArgs := strings.Join(os.Args[2:], " ")
		parseArgs, err := recommend.ProcessArgs(rawArgs)
		if err != nil {
			return fmt.Errorf("error processing args: %v", err)
		}

		if err := recommend.Recommend(client, parseArgs); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(recommendCmd)

	recommendCmd.Flags().StringSliceVarP(&recommendOptions.Policy, "policy", "p", []string{"KubeArmorPolicy"}, "Types of policy that can be recommended: KubeArmorPolicy|KyvernoPolicy")
	recommendCmd.Flags().StringSliceVarP(&recommendOptions.Namespace, "namespace", "n", []string{}, "Filter by Namespace")
	recommendCmd.Flags().StringVarP(&recommendOptions.Grpc, "gRPC", "", "", "gRPC address of discovery engine")
	recommendCmd.Flags().BoolVar(&recommendOptions.Dump, "dump", false, "Dump policies to knoxctl_out directory and skip TUI")
	recommendCmd.Flags().StringVarP(&recommendOptions.View, "view", "v", "", "View policies as table, yaml or json.")
}
