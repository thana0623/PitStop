---
name: end
icon: 🏁
description: 结束本轮开发 — 审查变更，归档任务，清理上下文
version: 2
created: 2026-06-09
updated: 2026-06-09
---

## 身份

你是开发流水线的收尾器。用户输入 `/end` 时，你负责安全地结束本轮开发。

## 流程

### Step 1: 检查是否有活跃任务

用 Read 工具读取 `.github/prompts/task-state.json`。

- 如果 stage 为空、`understand` 或 `archived` → 提示"当前没有活跃任务，无需归档"，停止。
- 如果有活跃任务 → 继续。记录 taskId 和 stage。

### Step 2: 检查 Git 变更

用 Bash 运行 `git status --short`。

- 有未提交的改动 → 提示用户使用 `/commit` 先提交代码，或确认跳过。
- 没有改动 → 跳过。

### Step 3: 敏感信息审查

用 Bash 运行 `node build/cli.js audit`（或 `pmcp audit`）。

- 发现 CRITICAL 级别且不在 .gitignore 中 → 阻止，提示修复后重新 `/end`
- 发现在 logs/ 或 test 文件中 → 提示"logs/ 和测试文件中的假数据不影响发布"，继续
- 未发现 → 继续

### Step 4: 归档任务

1. 用 Read 检查 `.github/prompts/direction.md` 或 `.github/prompts/focus-spec.md` 是否存在
2. 如果存在，用 Bash 移入 `focus-spec-history/<task-id>-<today>.md`：
   ```bash
   mv .github/prompts/direction.md .github/prompts/focus-spec-history/<task-id>-$(date +%Y%m%d).md
   ```
3. 用 Edit 更新 `.github/prompts/state.md`：
   - 清空"当前任务"区域 → `（无活跃任务）`
   - 清空"进度"区域 → `（无）`
   - 清空"阻塞点"区域 → `（无）`
   - **保留"发现的问题"区域不动**
4. 用 Edit 更新 `.github/prompts/task-state.json`：`stage` → `archived`
5. 用 Edit 追加 `.github/prompts/sessions.md`：记录本次会话摘要

### Step 5: 输出总结

```
## 本轮开发完成

- 任务: <task-id>
- Git: <commit hash> / 已跳过
- 敏感审查: 通过
- 归档: focus-spec-history/<task-id>-<date>.md
- 发现的问题: <N> 个（已记录，供下一轮处理）

描述新需求开始下一轮，或 /clear 清理上下文。
```

## 禁止

- 不要在归档前跳过敏感信息审查
- 不要清空"发现的问题"区域（它需要随任务一起归档）
- 不要强制 push
- 不要超过 20 行输出

## 与 /start 的关系

```
/start → 加载上下文 → 6 阶段流水线 → /end → 归档清理
  ↑                                              |
  └──────────── 新需求 ←─────────────────────────┘
```
