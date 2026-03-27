# Default Output Dir Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 让未传 `-out` 的解包命令默认输出到当前工作目录下的 `output/{appid}-{应用名称}/{版本号}`。

**Architecture:** 在命令入口层计算新的默认输出目录，而不是改底层解包流程。`all` 命令直接复用扫描结果中的名称和版本，默认解包命令按输入路径推断版本并通过现有查询链路补应用名称。

**Tech Stack:** Go 1.23、标准库 `os`/`filepath`/`regexp`、现有 `locator` 与 `cmd.Execute`

---

### Task 1: 抽取默认输出路径 helper

**Files:**
- Create: `internal/cmd/output_path.go`
- Modify: `/Users/cai/code/crack/Gwxapkg/internal/cmd/cmd.go`
- Test: none (用户明确要求不新增测试代码)

**Step 1: 写路径构造 helper**

在 `internal/cmd` 中新增：

```go
func BuildDefaultOutputDir(cwd, appID, appName, version string) string
```

要求：
- 根目录固定为 `filepath.Join(cwd, "output")`
- 应用目录为 `{appid}-{应用名称}`，名称为空则退化为 `{appid}`
- 版本号为空时使用 `unknown`

**Step 2: 写名称清洗 helper**

新增内部 helper 清洗目录名中的非法字符和多余空白。

**Step 3: 写版本识别 helper**

新增：

```go
func DetectVersionFromInput(input string) string
```

规则：
- 目录输入取 basename
- 文件输入取父目录 basename
- basename 纯数字则返回，否则 `unknown`

### Task 2: 让 `all` 命令使用新默认目录

**Files:**
- Modify: `main.go`
- Modify: `internal/locator/remote.go`
- Test: none (手动验证)

**Step 1: 在 `all` 命令里补名称**

在扫描结果后调用现有名称补全逻辑，确保 `matched.AppName` 尽可能可用。

**Step 2: 仅在 `-out` 为空时生成默认目录**

在 `handleAllCommand` 中：
- 调 `os.Getwd()`
- 用 `matched.AppID`、`matched.AppName`、`matched.Version` 生成默认目录
- 若 `Getwd` 失败则回退现有行为

### Task 3: 让默认解包命令使用新默认目录

**Files:**
- Modify: `main.go`
- Modify: `internal/locator/remote.go`
- Test: none (手动验证)

**Step 1: 在默认命令里补应用名称**

新增一个对外 helper，例如：

```go
func ResolveMiniProgramName(appID, input string) string
```

优先用缓存/远程查询补齐名称。

**Step 2: 仅在 `-out` 为空时生成默认目录**

在 `handleDefaultCommand` 中：
- 调 `os.Getwd()`
- 调版本识别 helper
- 调名称解析 helper
- 生成默认输出目录

### Task 4: 轻量验证

**Files:**
- Test: none (手动验证)

**Step 1: 验证 `all` 命令默认输出**

Run:

```bash
go run . all -id=wxea4fc0bdd2da0288
```

Expected:
- 终端输出的“输出目录”是 `./output/wxea4fc0bdd2da0288-上蔬云采OTO农产品交易平台/72`

**Step 2: 验证默认解包命令命中版本号**

Run:

```bash
go run . -id=wxea4fc0bdd2da0288 -in='<版本目录或其下 wxapkg 文件>'
```

Expected:
- 默认输出目录命中版本号 `72`

**Step 3: 验证默认解包命令的 `unknown` 降级**

Run:

```bash
go run . -id=wxea4fc0bdd2da0288 -in='./some/path/without/version'
```

Expected:
- 默认输出目录最后一级为 `unknown`
