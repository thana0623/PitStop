<div align="center">

# 🏁 PitStop

**环境检查 + 启动脚本生成 — AI 写完代码后的第一公里验证。**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)

</div>

---

## 它是什么？

PitStop 是一个 Go CLI 工具，做三件事：

1. **检查环境** — 比对项目需要什么（Java 17、Node 22）和本机有什么
2. **启动项目** — 一条命令启动所有服务（SpringBoot、Vue、…）
3. **导出脚本** — 生成可独立运行的 `start.sh` / `start.bat`，跟项目走

```bash
pitstop check    # 检查环境是否满足要求
pitstop start    # 启动所有服务
pitstop stop     # 一键停止
pitstop report   # 查看日志摘要
pitstop export   # 导出启动/停止脚本
```

## 快速开始

### 1. 安装

```bash
go install github.com/thana0623/PitStop@latest
```

### 2. 创建项目配置 `pitstop.yaml`

```yaml
project:
  name: my-app

services:
  backend:
    cwd: ./backend
    command: mvn spring-boot:run
    port: 8080
    health_check: http://localhost:8080/actuator/health

  frontend:
    cwd: ./frontend
    command: npm run dev
    port: 5173
    health_check: http://localhost:5173

requirements:
  java: ">=17"
  node: ">=22"
  maven: ">=3.8"

logging:
  dir: ./logs
  stdout: true
```

### 3. 检查环境

```bash
pitstop check
```

输出：

```
Environment Check
=================

  ✅ java 17.0.1 >=17 (/usr/lib/jvm/java-17/bin/java)
  ✅ node 22.0.0 >=22 (/usr/local/bin/node)
  ❌ maven not found (need >=3.8)

❌ Some requirements not met.
```

### 4. 启动

```bash
pitstop start
```

启动前自动运行环境检查，全部通过后启动服务。跳过检查：

```bash
pitstop start --skip-check
```

### 5. 导出脚本

```bash
pitstop export -o scripts
```

生成 4 个文件：

```
scripts/
├── start.sh     # Linux/macOS 启动
├── stop.sh      # Linux/macOS 停止
├── start.bat    # Windows 启动
└── stop.bat     # Windows 停止
```

生成的脚本可以直接运行，也可以提交到 git，跟项目走。

### 6. 停止 & 报告

```bash
pitstop stop     # 优雅停止所有服务
pitstop report   # 查看日志摘要（ERROR/WARN/INFO 统计）
```

## 配置说明

### 项目配置 `pitstop.yaml`（提交到 git）

| 字段 | 说明 |
|------|------|
| `project.name` | 项目名称 |
| `services.<name>.cwd` | 工作目录（相对于配置文件） |
| `services.<name>.command` | 启动命令 |
| `services.<name>.port` | 端口号（可选） |
| `services.<name>.health_check` | 健康检查 URL（可选） |
| `requirements.<tool>` | 版本约束，如 `">=17"`、`">=3.8.0"` |
| `logging.dir` | 日志目录 |
| `logging.stdout` | 是否同时输出到终端 |

### 本机配置 `~/.pitstop/config.yaml`（不提交）

```yaml
paths:
  java: /usr/lib/jvm/java-17/bin/java
  node: /usr/local/bin/node
  maven: /usr/local/bin/mvn
ports:
  range: [10000, 60000]
```

不存在时自动从 PATH 检测。可选配置，覆盖检测结果。

## 命令一览

| 命令 | 说明 |
|------|------|
| `pitstop check` | 检查环境是否满足 requirements |
| `pitstop start` | 启动所有服务（自动检查环境） |
| `pitstop start --skip-check` | 跳过环境检查直接启动 |
| `pitstop start -c path` | 指定配置文件 |
| `pitstop stop` | 优雅停止所有服务 |
| `pitstop report` | 日志摘要报告 |
| `pitstop export` | 导出启动/停止脚本 |
| `pitstop export -o dir` | 指定脚本输出目录 |

## 架构

```
pitstop.yaml (项目配置)          ~/.pitstop/config.yaml (本机配置)
       │                                    │
       ▼                                    ▼
  ┌─────────────────────────────────────────┐
  │              PitStop CLI                │
  ├─────────────┬───────────┬───────────────┤
  │  envcheck   │  process  │  scriptgen    │
  │  版本检查    │  进程管理  │  脚本生成     │
  └─────────────┴───────────┴───────────────┘
       │              │              │
       ▼              ▼              ▼
   pitstop check  pitstop start  pitstop export
```

## 适用场景

- ✅ SpringBoot + Vue 前后端分离项目
- ✅ AI 辅助开发后的快速验证
- ✅ 新环境搭建时快速检查依赖
- ✅ 生成可移植的启动脚本

## Roadmap

- [x] V1.0 — 一键启动 + 热重载 + 日志收集
- [x] V2.0 — 环境检查 + 本机配置 + 脚本导出
- [ ] V2.1 — 服务器依赖管理（Redis/MQ/MySQL 远程配置）
- [ ] V2.2 — 结构化测试报告（Markdown/HTML）
- [ ] V3.0 — Django 支持

## License

[MIT](LICENSE)
