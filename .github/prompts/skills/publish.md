---
name: publish
icon: 🚀
description: 一键安全发布：审计 + 版本 bump + npm + git + link
version: 1
created: 2026-06-03
updated: 2026-06-03
---

# 🚀 安全发布

一键完成隐私审计 + 版本 bump + npm publish + git push + npm link + README 更新。

## 使用方式

```bash
# 基本发布（patch 版本）
pmcp publish --token npm_xxx

# 指定版本类型
pmcp publish --bump minor --token npm_xxx

# 自定义 commit message
pmcp publish --message "feat: 新功能" --token npm_xxx

# 跳过审计（不推荐）
pmcp publish --skip-audit --token npm_xxx
```

## 发布流程

1. 🔒 隐私审计 — 扫描密钥/Token/密码
2. 📦 版本号 bump — patch/minor/major
3. 📝 README 同步 — 更新版本号
4. 📤 npm publish — 发布到 npm
5. 🔗 git push — 推送到远程
6. 🔗 npm link — 更新全局链接

## 参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--token` | npm Access Token | 无（必须提供） |
| `--bump` | 版本类型 patch/minor/major | patch |
| `--message` | git commit message | chore: bump version to x.x.x |
| `--skip-audit` | 跳过隐私审计 | false |

## 下一步

- 发布前用 `/audit` 先单独检查隐私信息
- 用 `/tools` 查看项目工具清单
