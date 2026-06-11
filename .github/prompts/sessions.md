# 会话记录

## 2026-06-11 PitStop PRD + Plan

### 完成内容

**需求分析（PRD）**
- 痛点：AI 代码验证占 50%+ 开发时间，测试是编码时间的 2 倍以上
- 核心目标：一键启动 + 热重载，把重启时间从分钟级降到秒级
- 技术选型：Go CLI + YAML 配置 + spring-boot-devtools + Vite HMR
- CLI 命令：`pitstop start` / `pitstop stop` / `pitstop report`
- PRD 路径：`.claude/prds/dev-test-automation.prd.md`

**实现规划（Plan）**
- 8 个任务：脚手架 → 配置解析 → 进程管理 → start 命令 → 日志收集 → 健康检查 → stop 命令 → report 命令
- 目录结构：`cmd/` + `internal/config|process|logger|report/`
- Plan 路径：`.claude/plans/dev-test-automation.plan.md`

### 待完成
- [ ] Go 环境确认（go 命令未找到）
- [ ] Task #1-8 实现

---

## 项目命令速查

```
PitStop — 一键启动热重载工具

配置文件：pitstop.yaml
CLI 命令：
  pitstop start          # 启动项目（读取当前目录 pitstop.yaml）
  pitstop start -c path  # 指定配置文件路径
  pitstop stop           # 优雅停止所有服务
  pitstop report         # 生成日志摘要报告

配置示例：
  project:
    name: "my-app"
  services:
    backend:
      type: springboot
      path: "./cdz"
      command: "mvn spring-boot:run"
      port: 12011
      health_check: "http://localhost:12011/api/health"
    frontend:
      type: vue
      path: "./cdz2"
      command: "pnpm dev"
      port: 12010
      health_check: "http://localhost:12010"
  logging:
    dir: "./logs"
    stdout: true

文档：
  PRD:  .claude/prds/dev-test-automation.prd.md
  Plan: .claude/plans/dev-test-automation.plan.md
```
