package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"pitstop/internal/config"
	"pitstop/internal/envcheck"
	"pitstop/internal/localconfig"
	"pitstop/internal/logger"
	"pitstop/internal/process"

	"github.com/spf13/cobra"
)

var cfgFile string
var skipCheck bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "启动所有服务",
	Long:  "读取 pitstop.yaml 配置，检查环境，启动 SpringBoot + Vue 服务，支持热重载。",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStart()
	},
}

func init() {
	startCmd.Flags().StringVarP(&cfgFile, "config", "c", "pitstop.yaml", "配置文件路径")
	startCmd.Flags().BoolVar(&skipCheck, "skip-check", false, "跳过环境检查")
	rootCmd.AddCommand(startCmd)
}

func runStart() error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Preflight: environment check.
	if !skipCheck && len(cfg.Requirements) > 0 {
		localCfg, err := localconfig.Load("")
		if err != nil {
			return fmt.Errorf("loading local config: %w", err)
		}
		checker := envcheck.New(cfg.Requirements, localCfg)
		results := checker.Check()
		if !envcheck.CheckAllPassed(results) {
			fmt.Println(envcheck.FormatResults(results))
			fmt.Println("\nUse --skip-check to start anyway.")
			return fmt.Errorf("environment check failed")
		}
		fmt.Println("✅ Environment check passed.")
	}

	fmt.Printf("PitStop — %s\n", cfg.Project.Name)
	fmt.Printf("Starting %d service(s)...\n\n", len(cfg.Services))

	// Initialize logger.
	logDir := cfg.Logging.Dir
	log := logger.New(logDir, cfg.Logging.Stdout)
	defer log.Close()

	// Initialize process manager.
	pidDir := filepath.Join(logDir, ".pids")
	mgr := process.New(pidDir)

	// Start each service.
	for name, svc := range cfg.Services {
		workDir, err := resolvePath(svc.Path)
		if err != nil {
			return fmt.Errorf("service %q: %w", name, err)
		}

		writer, err := log.Writer(name)
		if err != nil {
			return fmt.Errorf("service %q log: %w", name, err)
		}

		fmt.Printf("  Starting [%s] %s in %s\n", name, svc.Type, workDir)
		fmt.Printf("    Command: %s\n", svc.Command)
		if svc.HealthCheck != "" {
			fmt.Printf("    Health:  %s\n", svc.HealthCheck)
		}

		if err := mgr.Start(name, workDir, svc.Command, writer); err != nil {
			return fmt.Errorf("starting %s: %w", name, err)
		}
	}

	fmt.Printf("\nAll services started. Logs: %s\n", logDir)

	// Health checks.
	healthURLs := make(map[string]string)
	for name, svc := range cfg.Services {
		if svc.HealthCheck != "" {
			healthURLs[name] = svc.HealthCheck
		}
	}

	if len(healthURLs) > 0 {
		fmt.Println("Running health checks...")
		hc := process.NewHealthChecker()
		healthCtx, healthCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer healthCancel()

		results := hc.CheckAll(healthCtx, healthURLs)
		for _, r := range results {
			if r.Ready {
				fmt.Printf("  ✅ [%s] ready (%s)\n", r.Service, r.URL)
			} else {
				fmt.Printf("  ❌ [%s] not ready (%s): %v\n", r.Service, r.URL, r.Err)
			}
		}
	}

	fmt.Println("\nPress Ctrl+C to stop all services.")

	// Wait for interrupt signal.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	fmt.Println("\nShutting down...")

	// Graceful shutdown with timeout.
	shutdownTimeout := 10 * time.Second
	errs := mgr.StopAll(shutdownTimeout)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  Shutdown error: %v\n", e)
		}
		return fmt.Errorf("shutdown completed with %d error(s)", len(errs))
	}

	fmt.Println("All services stopped.")
	return nil
}

// resolvePath resolves a relative path to an absolute path.
func resolvePath(p string) (string, error) {
	if filepath.IsAbs(p) {
		return p, nil
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("resolving path %s: %w", p, err)
	}
	return abs, nil
}
