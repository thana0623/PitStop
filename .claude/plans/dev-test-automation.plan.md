# Plan: PitStop — 一键启动热重载工具

**Source PRD**: `.claude/prds/dev-test-automation.prd.md`
**Selected Milestone**: V1 — 一键启动 + 热重载 + 日志收集
**Complexity**: Medium

## Summary

构建一个 Go CLI 工具，通过 YAML 配置一键启动 SpringBoot + Vue 项目，收集 stdout 日志，支持健康检查和优雅停止。SpringBoot 依赖 DevTools 实现热重载，Vue 依赖 dev server 内置 HMR。

## Patterns to Mirror

全新项目，无现有代码可参考。采用 Go 社区标准实践：

| Category | Source | Pattern |
|---|---|---|
| CLI | cobra 文档 | `cmd/` 目录，每个命令一个文件 |
| Config | Go YAML 标准 | `config/` 目录，struct + yaml tag |
| Process | os/exec | 子进程管理，信号处理 |
| Logging | Go 标准库 log | 分级日志，写文件 |

## 目录结构

```
pitstop/
├── main.go                    # 入口
├── go.mod
├── go.sum
├── cmd/
│   ├── root.go               # root 命令
│   ├── start.go              # start 命令
│   ├── stop.go               # stop 命令
│   └── report.go             # report 命令
├── internal/
│   ├── config/
│   │   └── config.go         # YAML 配置解析
│   ├── process/
│   │   └── manager.go        # 进程管理（启动/停止/健康检查）
│   ├── logger/
│   │   └── logger.go         # 日志收集（stdout → 文件）
│   └── report/
│       └── reporter.go       # 日志摘要生成
├── pitstop.yaml               # 示例配置
└── README.md
```

## Files to Change

| File | Action | Why |
|---|---|---|
| `go.mod` | CREATE | Go 模块初始化 |
| `main.go` | CREATE | 程序入口 |
| `cmd/root.go` | CREATE | cobra root 命令 |
| `cmd/start.go` | CREATE | start 命令实现 |
| `cmd/stop.go` | CREATE | stop 命令实现 |
| `cmd/report.go` | CREATE | report 命令实现 |
| `internal/config/config.go` | CREATE | YAML 配置解析 |
| `internal/process/manager.go` | CREATE | 进程管理核心 |
| `internal/logger/logger.go` | CREATE | 日志收集 |
| `internal/report/reporter.go` | CREATE | 日志摘要 |
| `pitstop.yaml` | CREATE | 示例配置文件 |
| `README.md` | CREATE | 使用说明 |

## Tasks

### Task 1: 项目脚手架
- **Action**: 初始化 Go 模块，创建目录结构，安装 cobra 依赖
- **Validate**: `go build` 成功

### Task 2: YAML 配置解析
- **Action**: 定义 Config struct，实现 YAML 解析，支持 `-c` 参数指定配置路径
- **Mirror**: Go struct + `yaml:""` tag
- **Validate**: 写测试用例，解析 pitstop.yaml 成功

### Task 3: 进程管理器
- **Action**: 实现 ProcessManager — 启动子进程、记录 PID、优雅停止（SIGTERM → 等待 → SIGKILL）
- **Mirror**: os/exec + os.Signal
- **Validate**: 启动一个简单进程（如 `sleep 100`），能正常停止

### Task 4: start 命令
- **Action**: 读取配置 → 逐个启动 service → 收集 stdout → 健康检查 → 等待信号停止
- **Mirror**: cobra RunE 模式
- **Validate**: `pitstop start` 能启动 SpringBoot + Vue，Ctrl+C 优雅停止

### Task 5: stdout 日志收集
- **Action**: 将每个 service 的 stdout/stderr 通过 io.MultiWriter 同时写入文件和终端
- **Mirror**: Go io.Pipe / io.MultiWriter
- **Validate**: 启动后 logs/ 目录有日志文件，内容与终端一致

### Task 6: 健康检查
- **Action**: 启动后 HTTP 轮询 health_check URL，超时告警但不阻塞
- **Mirror**: http.Get + time.Ticker
- **Validate**: 服务启动后健康检查通过，输出就绪日志

### Task 7: stop 命令
- **Action**: 读取 PID 文件 → 发送 SIGTERM → 等待 → 清理 PID 文件
- **Mirror**: os.FindProcess + Signal
- **Validate**: `pitstop stop` 停止所有服务，进程不存在

### Task 8: report 命令
- **Action**: 读取日志文件 → 统计 ERROR/WARN/INFO 数量 → 输出摘要
- **Mirror**: Go bufio.Scanner + 正则匹配
- **Validate**: 有错误日志时 report 显示错误计数

## Validation

```bash
# 编译
go build -o pitstop .

# 测试配置解析
./pitstop start -c pitstop.yaml

# 测试启动（观察日志输出）
# Ctrl+C 测试优雅停止

# 测试 stop
./pitstop stop

# 测试 report
./pitstop report
```

## Risks

| 风险 | 可能性 | 缓解措施 |
|---|---|---|
| DevTools 重启慢（>30s） | 中 | V1 可接受，V2 优化 |
| Maven 进程内存高 | 中 | 支持 java -jar 备选 |
| WSL 文件监听延迟 | 低 | DevTools 用 polling |
| 子进程僵尸 | 低 | WaitGroup + 信号处理 |

## Acceptance

- [ ] `pitstop start` 一条命令启动 SpringBoot + Vue
- [ ] 代码修改后 SpringBoot DevTools 自动重启
- [ ] Vue HMR 自动生效
- [ ] stdout 日志写入 logs/ 目录
- [ ] 健康检查显示服务就绪状态
- [ ] `pitstop stop` 优雅停止所有进程
- [ ] `pitstop report` 输出日志摘要

---
*Status: READY — 等待确认后开始实现。*
