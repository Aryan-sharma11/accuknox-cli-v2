// SPDX-License-Identifier: Apache-2.0
// Copyright 2022 Authors of KubeArmor

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/accuknox/accuknox-cli-v2/pkg/summary"
	"github.com/spf13/cobra"
)

var summaryOptions summary.Options

// summaryCmd represents the summary command
var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Observability from discovery engine",
	Long:  `Discovery engine keeps the telemetry information from KubeArmor, knoxctl connects to it to provide this as observability data`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rawArgs := strings.Join(os.Args[2:], " ")
		parseArgs, err := summary.ProcessArgs(rawArgs)
		if err != nil {
			return fmt.Errorf("errors processing args: %v", err)
		}

		if err := summary.Summary(client, *parseArgs); err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(summaryCmd)

	summaryCmd.Flags().StringVar(&summaryOptions.GRPC, "gRPC", "", "gRPC server information")
	summaryCmd.Flags().StringArrayVarP(&summaryOptions.Namespace, "namespace", "n", []string{}, "Namespace")
	summaryCmd.Flags().StringVarP(&summaryOptions.Operation, "operation", "o", "", "Summary filter type : process|file|network ")
	summaryCmd.Flags().StringVarP(&summaryOptions.View, "view", "v", "json", "Print data on console in table or json format")
	summaryCmd.Flags().BoolVar(&summaryOptions.Dump, "dump", false, "Dump json data to knoxctl_out directory and skip TUI")
	summaryCmd.Flags().BoolVar(&summaryOptions.Glance, "glance", false, "Glance at the summary data")
	summaryCmd.Flags().BoolVar(&summaryOptions.NoTUI, "no-tui", false, "Disable TUI and progress bar")
	summaryCmd.Flags().StringVar(&summaryOptions.OutputTo, "out", "", "write out files to a specified directory")
	//summaryCmd.Flags().BoolVar(&summaryOptions.Aggregation, "agg", false, "Aggregate destination files/folder path")
}
