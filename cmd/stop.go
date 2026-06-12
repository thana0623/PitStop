package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"pitstop/internal/config"
	"pitstop/internal/process"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "优雅停止所有服务",
	Long:  "读取 PID 文件，发送 SIGTERM 信号优雅停止所有服务。",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStop()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop() error {
	// Load config to get log dir.
	cfg, err := config.Load(cfgFile)
	if err != nil {
		// If config not found, try default PID dir.
		return stopFromDir(".logs/.pids")
	}

	pidDir := filepath.Join(cfg.Logging.Dir, ".pids")
	return stopFromDir(pidDir)
}

func stopFromDir(pidDir string) error {
	fmt.Println("Stopping all services...")

	mgr := process.New(pidDir)
	timeout := 10 * time.Second
	errs := mgr.StopFromPIDFiles(timeout)

	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", e)
		}
		return fmt.Errorf("stop completed with %d error(s)", len(errs))
	}

	// Clean up PID directory.
	os.RemoveAll(pidDir)

	fmt.Println("All services stopped.")
	return nil
}
