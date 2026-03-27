# Scan App Name Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 让 `scan` 命令在本地扫描结果中尽量展示从已解出包文件里提取到的小程序名称。

**Architecture:** 保持 `internal/locator` 继续负责扫描包目录，但在构造 `MiniProgramInfo` 时额外跑一次本地名称提取。提取逻辑只读取包目录下的文本文件，抽取候选标题字符串并按启发式打分，拿高置信名称；拿不到时再参考 `magbone/wxapkg` 的做法，调用外部 AppID 查询接口补名称，并做本地缓存。

**Tech Stack:** Go 1.23、标准库 `filepath`/`regexp`/`unicode`、现有 CLI 输出层

---

### Task 1: 补充扫描结果名称字段

**Files:**
- Modify: `internal/locator/locator.go`
- Create: `internal/locator/name.go`

**Step 1: 扩展扫描结果结构**

给 `MiniProgramInfo` 增加 `AppName` 字段，并在扫描到版本目录后调用名称提取函数。

**Step 2: 实现最小名称提取**

从包目录里的 `.js` `.json` `.html` `.wxml` 文本文件抽取候选字符串，优先保留类似 `title`、`name`、`navigationBarTitleText` 等键对应的值，再对通用字符串做保守筛选。

**Step 3: 实现启发式打分**

对候选名称按重复次数、是否含“平台/商城/市场/小程序/云采”等业务关键词、是否明显像页面标题或提示文案进行加减分，返回最高分且超过阈值的名称。

### Task 2: 外部查询兜底

**Files:**
- Create: `internal/locator/remote.go`
- Modify: `main.go`

**Step 1: 参考 `magbone/wxapkg` 查询接口**

在 `scan` 命令里按 AppID 调用 `https://kainy.cn/api/weapp/info/`，读取返回的 `nickname` 字段。

**Step 2: 增加本地缓存**

把已查到的 AppID -> 名称结果缓存到用户目录下，减少重复请求。

### Task 3: 展示 scan 输出名称

**Files:**
- Modify: `main.go`
- Modify: `internal/ui/ui.go`

**Step 1: 修改输出签名**

让 `PrintMiniProgram` 支持接收应用名称。

**Step 2: 保持降级行为**

当提取不到名称时，保留现有 AppID 输出；当提取到名称时，首行展示名称，详细信息里补充 AppID。

### Task 4: 轻量验证

**Files:**
- None

**Step 1: 运行命令验证**

Run: `go run . scan`

Expected:
- 对已解出的 `wxea4fc0bdd2da0288` 能显示 `上蔬云采OTO农产品交易平台`
- 提取失败的包继续正常展示，不影响扫描结果

**Step 2: 用户要求下跳过新增自动化测试**

本次按用户明确要求不新增测试代码，使用命令行手动验证作为完成依据。
