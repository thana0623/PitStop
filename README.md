<div align="center">

# 🏁 PitStop

**一键启动，改完即跑。**

AI 写完代码后的第一公里验证工具。

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)

</div>

---

## 为什么需要 PitStop？

AI 辅助开发让写代码变快了，但**验证代码是否可用**依然是最慢的环节。

| 环节 | 耗时 | 痛点 |
|------|------|------|
| 设计 + 编码 | 30 min | ✅ AI 加速 |
| 启动 + 测试 | 60+ min | ❌ 手动重启、内存爆炸、反馈慢 |

**PitStop 的目标：** 把验证时间从编码时间的 2 倍降到 0.5 倍以下。

## 它是什么？

PitStop 是一个 **Go CLI 工具**，通过一份 YAML 配置文件，一键启动你的 SpringBoot + Vue 项目。

```bash
pitstop start    # 启动所有服务
pitstop stop     # 一键停止
pitstop report   # 查看日志摘要
```

**改完代码？不用手动重启。** SpringBoot DevTools 自动热重载，Vue HMR 即时生效。

## 快速开始

### 1. 安装

```bash
go install github.com/thana0623/PitStop@latest
```

### 2. 创建配置文件

在项目根目录创建 `pitstop.yaml`：

```yaml
project:
  name: "my-app"

services:
  backend:
    type: springboot
    path: "./backend"
    command: "mvn spring-boot:run"
    port: 8080
    health_check: "http://localhost:8080/actuator/health"

  frontend:
    type: vue
    path: "./frontend"
    command: "pnpm dev"
    port: 5173
    health_check: "http://localhost:5173"

logging:
  dir: "./logs"
  stdout: true
```

### 3. 启动

```bash
pitstop start
```

输出：

```
🏁 PitStop — 启动中...
  ✅ backend   已启动 (PID: 12345) — http://localhost:8080
  ✅ frontend  已启动 (PID: 12346) — http://localhost:5173
  📋 日志输出到 ./logs/
  ⏹  按 Ctrl+C 停止所有服务
```

现在去改代码吧。保存后 SpringBoot 会自动重启，Vue 页面会即时更新。

### 4. 停止

```bash
pitstop stop
```

### 5. 查看报告

```bash
pitstop report
```

输出：

```
📊 日志摘要 — my-app
  backend:   ERROR: 2  WARN: 5  INFO: 128
  frontend:  ERROR: 0  WARN: 1  INFO: 45
  ⚠️  发现 2 个错误，详见 ./logs/backend.log
```

## 配置参考

```yaml
project:
  name: "项目名称"          # 显示在日志和报告中

services:
  <服务名>:
    type: springboot | vue  # 服务类型
    path: "./path"          # 项目目录（相对于配置文件）
    command: "启动命令"      # 如 mvn spring-boot:run / pnpm dev
    port: 8080              # 服务端口
    health_check: "URL"     # 健康检查地址（可选）

logging:
  dir: "./logs"             # 日志目录
  stdout: true              # 是否同时输出到终端
```

## 架构

```
pitstop start
│
├── 解析 pitstop.yaml
│
├── 启动 SpringBoot
│   └── DevTools 监听 classpath → 自动重启
│
├── 启动 Vue
│   └── Vite HMR 即时更新
│
├── 收集 stdout 日志 → ./logs/
│
├── 健康检查（HTTP 轮询）
│
└── Ctrl+C → 优雅停止所有进程
```

## 命令一览

| 命令 | 说明 |
|------|------|
| `pitstop start` | 启动所有服务（读取当前目录 `pitstop.yaml`） |
| `pitstop start -c path` | 指定配置文件路径 |
| `pitstop stop` | 停止所有服务 |
| `pitstop report` | 查看日志摘要 |

## 适用场景

- ✅ SpringBoot + Vue 前后端分离项目
- ✅ AI 辅助开发后的快速验证
- ✅ 本地开发环境内存紧张（不想开 IDE）
- ✅ 需要频繁重启验证的迭代开发

## 不适用场景

- ❌ CI/CD 流水线（太慢，用 GitHub Actions）
- ❌ 自动化 E2E 测试（PitStop 做人工验证）
- ❌ 多人协作部署（单人工具）
- ❌ 生产环境部署

## Roadmap

- [x] V1.0 — 一键启动 + 热重载 + 日志收集
- [ ] V1.1 — 服务器依赖管理（Redis/MQ/MySQL 远程配置）
- [ ] V1.2 — 结构化测试报告（Markdown/HTML）
- [ ] V2.0 — Django 支持
- [ ] V2.1 — 增量热替换（替代全量重启）

## License

[MIT](LICENSE)

---

<div align="center">

**PitStop** — 让 AI 代码验证快起来。

</div>
