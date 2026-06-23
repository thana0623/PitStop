package cmd

import (
	"fmt"
	"os"

	"pitstop/internal/config"
	"pitstop/internal/envcheck"
	"pitstop/internal/localconfig"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "检查环境是否满足项目要求",
	Long:  "读取 pitstop.yaml 的 requirements 段，与本机环境对比，输出检查结果。",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheck()
	},
}

func init() {
	checkCmd.Flags().StringVarP(&cfgFile, "config", "c", "pitstop.yaml", "配置文件路径")
	rootCmd.AddCommand(checkCmd)
}

func runCheck() error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if len(cfg.Requirements) == 0 {
		fmt.Println("No requirements specified in pitstop.yaml.")
		return nil
	}

	localCfg, err := localconfig.Load("")
	if err != nil {
		return fmt.Errorf("loading local config: %w", err)
	}

	checker := envcheck.New(cfg.Requirements, localCfg)
	results := checker.Check()

	fmt.Println(envcheck.FormatResults(results))

	if !envcheck.CheckAllPassed(results) {
		os.Exit(1)
	}

	return nil
}
