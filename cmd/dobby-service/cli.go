package main

import (
	"github.com/roppenlabs/dobby-service/internal/config"
	"github.com/spf13/cobra"
)

func initCLI() *cobra.Command {
	var cliCmd = &cobra.Command{
		Use:   "dobby-service",
		Short: "dobby-service CLI to manage the service",
	}

	cliCmd.AddCommand(startCommand())
	return cliCmd
}

func startCommand() *cobra.Command {
	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Starts the service",
		Run: func(cmd *cobra.Command, args []string) {
			configFile := "application"
			if len(args) > 0 {
				configFile = args[0]
			}

			config.InitConfig(configFile)

			serverDependencies, _ := InitDependencies()
			serverDependencies.server.Run(serverDependencies.handlers)
		},
	}

	return startCmd
}
