---
name: recommend
icon: 🎯
description: 根据开发场景推荐应使用的工具
version: 1
created: 2026-06-03
updated: 2026-06-03
---

# 🎯 场景推荐

根据当前开发场景，推荐应使用的工具和操作建议。

## 使用方式

```bash
pmcp recommend <场景>
```

## 可用场景

- `coding` — 写代码
- `testing` — 写测试
- `debugging` — 调试
- `reviewing` — 代码审查
- `refactoring` — 重构
- `deploying` — 部署
- `documenting` — 写文档
- `exploring` — 理解代码
- `planning` — 需求规划

## 示例

```bash
pmcp recommend testing
pmcp recommend debugging
pmcp recommend exploring
```

## 输出内容

- ✅ 推荐工具（已安装）
- ⚠️ 建议安装的工具
- 💡 操作建议

## 下一步

- 用 `/tools` 查看所有已安装工具
- 用 `/audit` 检查隐私信息
