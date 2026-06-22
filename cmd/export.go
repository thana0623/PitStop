package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"pitstop/internal/config"
	"pitstop/internal/scriptgen"

	"github.com/spf13/cobra"
)

var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "导出启动/停止脚本",
	Long:  "读取 pitstop.yaml，生成 start.sh / stop.sh / start.bat / stop.bat 到指定目录。",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExport()
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "scripts", "脚本输出目录")
	exportCmd.Flags().StringVarP(&cfgFile, "config", "c", "pitstop.yaml", "配置文件路径")
	rootCmd.AddCommand(exportCmd)
}

func runExport() error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Resolve output directory to absolute path.
	absOutput, err := filepath.Abs(exportOutput)
	if err != nil {
		return fmt.Errorf("resolving output path: %w", err)
	}

	gen := scriptgen.New(cfg, absOutput)
	files, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("generating scripts: %w", err)
	}

	fmt.Printf("📁 Scripts generated in %s\n\n", absOutput)
	for _, f := range files {
		info, _ := os.Stat(f)
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		fmt.Printf("  ✅ %s (%d bytes)\n", filepath.Base(f), size)
	}

	fmt.Printf("\nUsage:\n")
	fmt.Printf("  cd %s\n", filepath.Dir(absOutput))
	fmt.Printf("  ./start.sh    # Linux/macOS\n")
	fmt.Printf("  start.bat     # Windows\n")

	return nil
}
