---
name: tools
icon: 🔧
description: 扫描项目已安装工具，输出能力清单
version: 1
created: 2026-06-03
updated: 2026-06-03
---

# 🔧 工具扫描

扫描项目已安装的工具（npm 依赖、MCP Server），输出能力清单和使用建议。

## 使用方式

运行 `pmcp tools` 命令：

```bash
pmcp tools
```

## 输出内容

- 📦 npm 依赖中的已知工具
- 🔌 MCP Server 配置中的工具
- 每个工具的关键命令和适用场景

## 下一步

- 用 `/recommend <场景>` 获取特定场景下的工具推荐
- 用 `/audit` 检查项目中的隐私信息
