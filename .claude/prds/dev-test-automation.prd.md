# PitStop — 一键启动热重载工具

## Problem

AI 辅助开发中，代码编写只需 30 分钟，但启动+验证需要 60+ 分钟（2 倍以上）。测试环节占开发时间 50%+，是最大瓶颈。当前方案：
- IDE 热启动快但太吃内存（WSL + Docker + Claude Code + IDEA = OOM）
- CI/CD 太慢（5-10 分钟）
- 每次改完代码都要手动重启，循环往复

核心诉求：**一键启动项目，改完代码自动热重载，省掉手动重启的时间。**

## Evidence

- 实测：设计+编码 30 分钟，启动+测试 60+ 分钟，测试时间 ≥ 2× 编码时间
- 内存经常爆：WSL + Docker + Claude Code + 微信 + IDE 同时运行
- CI/CD 单次 5-10 分钟，反馈循环太慢

## Users

- **Primary**: 本人 — 独立开发者，使用 AI 辅助开发，需要快速验证代码可用性
- **Not for**: 团队协作、CI/CD 流水线、自动化 E2E 测试

## Hypothesis

我们相信 **一个 Go CLI 工具，配置化一键启动 + 热重载 + stdout 日志收集** 会 **把每次代码验证的重启时间从分钟级降到秒级** 给 **独立 AI 开发者**。
当 **`pitstop start` 启动后改代码自动重启、`pitstop stop` 一键停止、`pitstop report` 查看日志摘要** 时，我们知道做对了。

## Success Metrics

| 指标 | 目标 | 衡量方式 |
|------|------|----------|
| 首次启动 | `pitstop start` 一条命令 | 配置好 YAML 后执行 |
| 热重载延迟 | < 30 秒（SpringBoot DevTools 重启） | 改代码到服务就绪 |
| 内存占用 | 低于开 IDE | 进程内存对比 |
| 手动操作 | 改完代码后零手动重启 | 使用体验 |

## 技术选型

| 项目 | 选型 | 理由 |
|------|------|------|
| 语言 | Go | 单二进制、低内存、跨平台 |
| CLI | cobra / urfave/cli | Go CLI 标准库 |
| 配置 | YAML | 人类可读、支持注释 |
| SpringBoot 热重载 | spring-boot-devtools | 官方方案，监听 classpath 自动重启 |
| Vue 热重载 | Vite HMR / webpack-dev-server | 前端内置，无需额外处理 |
| 日志 | stdout 收集 | 简单直接，后续可扩展 |

## 架构

```
pitstop start
│
├── 解析 YAML 配置
│
├── 启动 SpringBoot（mvn spring-boot:run / java -jar）
│   └── DevTools 监听 classpath → 自动重启
│
├── 启动 Vue（npm run dev / pnpm dev）
│   └── Vite HMR 自动生效
│
├── 收集 stdout 日志 → 写入日志文件
│
├── 健康检查（HTTP 轮询端口）
│
└── pitstop stop → 优雅停止所有进程

pitstop report
│
└── 读取日志文件 → 生成摘要（错误数、警告数、关键信息）
```

## CLI 命令

```bash
pitstop start          # 启动项目（读取当前目录 pitstop.yaml）
pitstop stop           # 优雅停止所有服务
pitstop report         # 生成日志摘要报告
pitstop start -c path  # 指定配置文件路径
```

## 配置文件格式（pitstop.yaml）

```yaml
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
```

## Scope

**V1（本次）— 一键启动 + 热重载 + 日志收集**

1. YAML 配置解析
2. `pitstop start` — 启动 SpringBoot + Vue
3. SpringBoot DevTools 热重载
4. Vue HMR（dev server 内置）
5. stdout 日志收集 → 写入文件
6. 健康检查（HTTP 轮询端口就绪）
7. `pitstop stop` — 优雅停止所有进程
8. `pitstop report` — 日志摘要（错误/警告计数）

**V2（后续）— 服务器依赖 + 完整报告**
- 服务器依赖管理（Redis/MQ/MySQL 远程配置）
- 结构化测试报告（Markdown / HTML）
- Django 支持
- 增量热替换（替代全量重启）

**Out of scope**
- CI/CD 集成
- 自动化 E2E 测试
- 多人协作
- 代码部署

## Open Questions

- [ ] SpringBoot 启动方式：mvn spring-boot:run vs java -jar？（DevTools 在两种方式下的行为差异）
- [ ] 多服务项目的配置结构：一个 YAML 管多个服务 vs 每个服务一个 YAML？

## Risks

| 风险 | 可能性 | 影响 | 缓解措施 |
|------|--------|------|----------|
| DevTools 重启慢（>30s） | 中 | 中 | 可接受，后续优化为增量 |
| Maven 进程占用内存高 | 中 | 中 | 可用 java -jar 替代 |
| 文件监听在 WSL 下延迟 | 低 | 低 | 使用 polling fallback |

---
*Status: DRAFT — 需求阶段。实现规划通过 /plan 生成。*
