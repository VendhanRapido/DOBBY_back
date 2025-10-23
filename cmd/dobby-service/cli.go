package main

import (
	"fmt"

	"github.com/roppenlabs/dobby-service/internal/config"
	"github.com/roppenlabs/dobby-service/internal/utils"
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
	if err := utils.ExtractStaticFiles(); err != nil {
		fmt.Printf("Failed to initialize file system: %v\n", err)
		panic(err)
	}

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Starts the service",
		Run: func(cmd *cobra.Command, args []string) {
			configFile := "application"
			if len(args) > 0 {
				configFile = args[0]
			}

			config.InitConfig(configFile)

			serverDependencies, err := InitDependencies()
			if err != nil {
				fmt.Println(err)
				panic(err)
			}
			serverDependencies.server.Run(serverDependencies.handlers)
		},
	}

	return startCmd
}
