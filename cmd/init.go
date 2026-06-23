package cmd

import (
	"fmt"
	"os"

	"pitstop/internal/initcmd"

	"github.com/spf13/cobra"
)

var (
	initDir   string
	initForce bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "扫描项目并生成 pitstop.yaml",
	Long:  "自动检测项目类型（SpringBoot/Vue/React/Go 等），生成 pitstop.yaml 配置文件。",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

func init() {
	initCmd.Flags().StringVar(&initDir, "dir", ".", "项目目录路径")
	initCmd.Flags().StringVarP(&cfgFile, "output", "o", "pitstop.yaml", "输出文件路径")
	initCmd.Flags().BoolVar(&initForce, "force", false, "覆盖已有文件")
	rootCmd.AddCommand(initCmd)
}

func runInit() error {
	// Resolve absolute path.
	absDir, err := absPath(initDir)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	fmt.Printf("Scanning %s ...\n\n", absDir)

	result, err := initcmd.Scan(absDir)
	if err != nil {
		return fmt.Errorf("scanning project: %w", err)
	}

	if len(result.Services) == 0 {
		fmt.Fprintln(os.Stderr, "No services detected. Make sure you're in a project directory with pom.xml, package.json, go.mod, or manage.py.")
		os.Exit(1)
	}

	// Print detected services.
	fmt.Println("Detected services:")
	for _, svc := range result.Services {
		fmt.Printf("  ✅ %s (%s) — port %d\n", svc.Name, svc.Type, svc.Port)
	}
	fmt.Println()

	// Generate the config file.
	if err := initcmd.Generate(result, cfgFile, initForce); err != nil {
		return fmt.Errorf("generating config: %w", err)
	}

	fmt.Printf("Generated %s — please review and customize.\n", cfgFile)
	return nil
}

func absPath(p string) (string, error) {
	if p == "." {
		return os.Getwd()
	}
	// If already absolute, return as-is.
	if len(p) > 1 && p[1] == ':' {
		return p, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return cwd + "/" + p, nil
}
