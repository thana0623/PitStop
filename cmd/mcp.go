package cmd

import (
	"pitstop/internal/mcpserver"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "启动 MCP Server（stdio 模式）",
	Long:  "以 MCP 协议启动 PitStop Server，供 Claude Code 等 AI 工具调用。",
	RunE: func(cmd *cobra.Command, args []string) error {
		return mcpserver.Run()
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
