package cmd

import (
	"fmt"

	"pitstop/internal/config"
	"pitstop/internal/report"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "生成日志摘要报告",
	Long:  "读取日志文件，统计 ERROR/WARN/INFO 数量，输出摘要。",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReport()
	},
}

func init() {
	reportCmd.Flags().StringVarP(&cfgFile, "config", "c", "pitstop.yaml", "配置文件路径")
	rootCmd.AddCommand(reportCmd)
}

func runReport() error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	r := report.New(cfg.Logging.Dir)
	stats, err := r.Generate()
	if err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	fmt.Print(report.Format(stats))
	return nil
}
