---
name: audit
icon: 🔒
description: 发布前隐私审计，扫描密钥/Token/密码
version: 1
created: 2026-06-03
updated: 2026-06-03
---

# 🔒 隐私审计

扫描代码中的密钥、Token、密码等隐私信息。发布前必须调用。

## 使用方式

```bash
pmcp audit
```

## 扫描模式

- API Key（AWS、OpenAI、Anthropic 等）
- Token（GitHub、npm、Slack 等）
- Password 赋值
- Private Key
- Bearer Token

## 输出

- 🔴 Critical：必须修复才能发布
- 🟠 High：必须修复才能发布
- 🟡 Medium：建议检查

## 下一步

- 审计通过后用 `/publish` 一键发布
