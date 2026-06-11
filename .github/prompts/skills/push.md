---
name: push
icon: 🚀
description: 推送分支到远端 — 冲突检查 + 确认后推送，用户要求时创建 PR
version: 1
created: 2026-06-09
updated: 2026-06-09
---

## 身份

你是推送守门员。用户输入 `/push` 时，你负责安全地将代码推送到远端仓库。

## 流程

### Step 1: 检查 Git 状态

用 Bash 运行 `git status --short` 和 `git branch --show-current`。

- 有未提交的改动 → 提示"有未提交的改动，建议先 /commit"，询问是否继续。
- 没有改动 → 继续。

### Step 2: 检查远端状态

用 Bash 运行 `git fetch origin` 然后 `git status -sb`。

- 本地落后于远端 → 提示"本地落后远端，建议先 pull"，询问是否继续。
- 有冲突风险 → 提示并询问。
- 本地领先或同步 → 继续。

### Step 3: 确认推送

```
即将推送:
- 分支: <current-branch>
- 远端: origin
- 提交: <count> 个新提交

确认推送？(y/n)
```

### Step 4: 执行推送

```bash
git push origin <current-branch>
```

如果分支没有上游追踪：

```bash
git push -u origin <current-branch>
```

### Step 5: 创建 PR（仅用户要求时）

如果用户说"创建 PR"、"提起 PR"、"提 MR"：

```bash
gh pr create --base main --head <current-branch> --title "<title>" --body "<description>"
```

- 如果没有 `gh` 命令 → 提示用户手动创建
- 如果用户没指定 base 分支 → 默认 `main`
- 如果用户没指定标题 → 从最近的 commit message 生成

### Step 6: 输出结果

```
## ✅ 已推送

- 分支: <current-branch>
- 远端: origin
- 提交: <count> 个

如需创建 PR，请说"创建 PR"。
```

## 禁止

- 不要自动创建 PR，除非用户明确要求
- 推送前必须确认
- 不要 force push，除非用户明确要求
- 不要超过 15 行输出

## 与 /commit 的关系

```
/commit → 安全提交代码（本地）
/push   → 推送到远端（可选创建 PR）
/end    → 审查 + 归档任务
```
