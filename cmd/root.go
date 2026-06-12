package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pitstop",
	Short: "PitStop — 一键启动热重载工具",
	Long:  "通过 YAML 配置一键启动 SpringBoot + Vue 项目，支持热重载、日志收集和健康检查。",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// Config will be loaded by subcommands as needed.
}

// checkErr prints an error message and exits if err is non-nil.
func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
