package mcpserver

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"pitstop/internal/config"
	"pitstop/internal/envcheck"
	"pitstop/internal/initcmd"
	"pitstop/internal/localconfig"
	"pitstop/internal/logger"
	"pitstop/internal/process"
	"pitstop/internal/report"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Run starts the MCP server on stdio.
func Run() error {
	s := server.NewMCPServer(
		"pitstop",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s.AddTool(mcp.NewTool("pitstop_init",
		mcp.WithDescription("扫描项目目录，检测服务类型（SpringBoot/Vue/React/Go 等），生成 pitstop.yaml 配置文件"),
		mcp.WithString("dir",
			mcp.Description("项目目录路径，默认当前目录"),
		),
		mcp.WithBoolean("force",
			mcp.Description("覆盖已有 pitstop.yaml"),
		),
	), handleInit)

	s.AddTool(mcp.NewTool("pitstop_check",
		mcp.WithDescription("检查本机环境是否满足 pitstop.yaml 中的 requirements（Java/Node/Maven 等版本）"),
		mcp.WithString("config",
			mcp.Description("pitstop.yaml 路径，默认 pitstop.yaml"),
		),
	), handleCheck)

	s.AddTool(mcp.NewTool("pitstop_start",
		mcp.WithDescription("启动 pitstop.yaml 中定义的所有服务（后台运行，带健康检查）"),
		mcp.WithString("config",
			mcp.Description("pitstop.yaml 路径，默认 pitstop.yaml"),
		),
		mcp.WithBoolean("skip_check",
			mcp.Description("跳过环境检查"),
		),
	), handleStart)

	s.AddTool(mcp.NewTool("pitstop_stop",
		mcp.WithDescription("优雅停止所有 PitStop 管理的服务"),
		mcp.WithString("config",
			mcp.Description("pitstop.yaml 路径，默认 pitstop.yaml"),
		),
	), handleStop)

	s.AddTool(mcp.NewTool("pitstop_report",
		mcp.WithDescription("生成日志摘要报告，统计各服务的 ERROR/WARN/INFO 数量"),
		mcp.WithString("config",
			mcp.Description("pitstop.yaml 路径，默认 pitstop.yaml"),
		),
	), handleReport)

	return server.ServeStdio(s)
}

func handleInit(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dir := request.GetString("dir", ".")
	force := request.GetBool("force", false)

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("无效路径: %v", err)), nil
	}

	result, err := initcmd.Scan(absDir)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("扫描失败: %v", err)), nil
	}

	if len(result.Services) == 0 {
		return mcp.NewToolResultError("未检测到服务。确保目录包含 pom.xml、package.json、go.mod 或 manage.py"), nil
	}

	outputPath := filepath.Join(absDir, "pitstop.yaml")
	if err := initcmd.Generate(result, outputPath, force); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("生成配置失败: %v", err)), nil
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "已生成 %s\n\n检测到的服务:\n", outputPath)
	for _, svc := range result.Services {
		fmt.Fprintf(&buf, "  - %s (%s) — 端口 %d\n", svc.Name, svc.Type, svc.Port)
	}
	if len(result.Requirements) > 0 {
		fmt.Fprintf(&buf, "\n环境要求:\n")
		for k, v := range result.Requirements {
			fmt.Fprintf(&buf, "  - %s: %s\n", k, v)
		}
	}

	return mcp.NewToolResultText(buf.String()), nil
}

func handleCheck(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfgPath := request.GetString("config", "pitstop.yaml")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("加载配置失败: %v", err)), nil
	}

	if len(cfg.Requirements) == 0 {
		return mcp.NewToolResultText("pitstop.yaml 中未定义 requirements，跳过检查。"), nil
	}

	localCfg, err := localconfig.Load("")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("加载本地配置失败: %v", err)), nil
	}

	checker := envcheck.New(cfg.Requirements, localCfg)
	results := checker.Check()

	return mcp.NewToolResultText(envcheck.FormatResults(results)), nil
}

func handleStart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfgPath := request.GetString("config", "pitstop.yaml")
	skipCheck := request.GetBool("skip_check", false)

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("加载配置失败: %v", err)), nil
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "PitStop — %s\n", cfg.Project.Name)

	// Environment check
	if !skipCheck && len(cfg.Requirements) > 0 {
		localCfg, err := localconfig.Load("")
		if err == nil {
			checker := envcheck.New(cfg.Requirements, localCfg)
			results := checker.Check()
			if !envcheck.CheckAllPassed(results) {
				fmt.Fprintf(&buf, "\n环境检查未通过:\n%s\n", envcheck.FormatResults(results))
				return mcp.NewToolResultError(buf.String()), nil
			}
			fmt.Fprintf(&buf, "✅ 环境检查通过\n")
		}
	}

	// Initialize logger
	logDir := cfg.Logging.Dir
	log := logger.New(logDir, cfg.Logging.Stdout)

	// Start services
	mgr := process.New(filepath.Join(logDir, ".pids"))
	fmt.Fprintf(&buf, "\n启动 %d 个服务...\n", len(cfg.Services))

	started := 0
	for name, svc := range cfg.Services {
		workDir := svc.Path
		if workDir == "" {
			workDir = "."
		}
		absPath, err := filepath.Abs(workDir)
		if err != nil {
			absPath = workDir
		}

		writer, err := log.Writer(name)
		if err != nil {
			fmt.Fprintf(&buf, "  ❌ %s: 日志初始化失败: %v\n", name, err)
			continue
		}
		if err := mgr.Start(name, absPath, svc.Command, writer); err != nil {
			fmt.Fprintf(&buf, "  ❌ %s: %v\n", name, err)
			continue
		}
		fmt.Fprintf(&buf, "  ✅ %s (%s) — 端口 %d\n", name, svc.Type, svc.Port)
		started++
	}

	// Write PID files info
	fmt.Fprintf(&buf, "\n日志目录: %s\n", logDir)
	fmt.Fprintf(&buf, "已启动 %d/%d 个服务。使用 pitstop stop 停止。\n", started, len(cfg.Services))

	return mcp.NewToolResultText(buf.String()), nil
}

func handleStop(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfgPath := request.GetString("config", "pitstop.yaml")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		// Fallback: try default PID dir
		pidDir := filepath.Join(".logs", ".pids")
		return stopFromPIDDir(pidDir)
	}

	pidDir := filepath.Join(cfg.Logging.Dir, ".pids")
	return stopFromPIDDir(pidDir)
}

func stopFromPIDDir(pidDir string) (*mcp.CallToolResult, error) {
	mgr := process.New(pidDir)
	errs := mgr.StopFromPIDFiles(10_000_000_000) // 10 seconds

	if len(errs) > 0 {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "停止完成，但有 %d 个错误:\n", len(errs))
		for _, e := range errs {
			fmt.Fprintf(&buf, "  - %v\n", e)
		}
		return mcp.NewToolResultText(buf.String()), nil
	}

	os.RemoveAll(pidDir)
	return mcp.NewToolResultText("所有服务已停止。"), nil
}

func handleReport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfgPath := request.GetString("config", "pitstop.yaml")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("加载配置失败: %v", err)), nil
	}

	r := report.New(cfg.Logging.Dir)
	stats, err := r.Generate()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("生成报告失败: %v", err)), nil
	}

	return mcp.NewToolResultText(report.Format(stats)), nil
}
