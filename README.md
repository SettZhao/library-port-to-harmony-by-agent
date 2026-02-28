# library-port-to-harmony-by-agent

一个基于 **GitHub Copilot Agent** 的半自动化工具链，将 Android/Java 开源库移植到 HarmonyOS（API 12+）平台。

通过多个专职 Agent 协同工作，结合 MCP 实时 API 查询和领域知识 Skill，完成从分析、文档、代码迁移到构建验证的完整移植流程。

---

## 项目结构

```
library-port-to-harmony-by-agent/
├── .github/
│   ├── agents/                  # Agent 定义文件
│   │   ├── port-to-harmony.agent.md   # 入口协调者
│   │   ├── planner.agent.md           # 分析规划
│   │   ├── analyzer.agent.md          # API 映射分析
│   │   ├── documenter.agent.md        # 文档生成
│   │   ├── migrator.agent.md          # 代码迁移
│   │   └── builder.agent.md           # 构建验证
│   ├── instructions/
│   │   └── harmony-dev-rules.instructions.md  # 全局规则（自动注入所有 Agent）
│   └── skills/
│       └── android-to-harmonyos/      # 领域知识库
│           ├── SKILL.md               # 知识索引
│           ├── skills/                # 分类知识子模块
│           └── references/            # 文档模板与参考
├── mcp-servers/
│   └── harmony-docs/            # HarmonyOS API MCP 服务
│       ├── main.go
│       ├── harmony-docs.exe     # 预编译可执行文件
│       └── harmony-API-reference/  # 完整鸿蒙 API 文档（本地）
└── Template/                    # HarmonyOS 项目模板（直接用于移植）
    ├── library/                 # 库代码目录
    └── entry/                   # Demo + 测试目录
```

---

## 核心组件

### Agents — 多 Agent 协作工作流

6 个专职 Agent 按流水线方式协作，每个 Agent 有明确的职责边界：

```
[port-to-harmony]  收集需求、启动工作流
       ↓
   [planner]       只读分析库架构，输出移植计划
       ↓
   [analyzer]      用 MCP 工具查询每个 Android API 的鸿蒙对应项，生成映射表
       ↓
  [documenter]     生成 三方库规格.md + 方案设计.md（编码前强制完成）
       ↓
   [migrator]      逐模块迁移代码（Java/Kotlin → ArkTS / JNI → NAPI / View → ArkUI）
       ↓
   [builder]       hvigorw 构建 → hdc 安装 → hypium 测试，循环修复直到全部通过
```

| Agent | 工具权限 |
|-------|---------|
| `port-to-harmony` | read, search, todo |
| `planner` | read, search, harmony-docs MCP |
| `analyzer` | harmony-docs MCP（全部 4 个工具） |
| `documenter` | read, edit |
| `migrator` | read, edit, execute, harmony-docs MCP |
| `builder` | execute, read, edit |

### MCP Server — harmony-docs

本地运行的 MCP 服务，提供对完整鸿蒙 API 文档的实时查询，替代过期的静态文档。

**4 个工具：**

| 工具 | 功能 |
|------|------|
| `list_api_modules` | 列出全部可用的 HarmonyOS Kit/模块 |
| `get_module_apis` | 获取某 Kit 内所有 API 文件的元数据 |
| `get_api_detail` | 获取具体 API 文件的完整签名、参数说明、枚举、结构体 |
| `search_api` | 按关键词跨模块搜索 API |

### Skill — android-to-harmonyos

领域知识库，被 Agent 按需读取，提供：

| 子模块 | 内容 |
|--------|------|
| `skills/code-migration` | Java/Kotlin → ArkTS 语法映射、代码模式 |
| `skills/native-migration` | JNI/NDK → NAPI 迁移样例 |
| `skills/ui-migration` | Android View/Compose → ArkUI 映射表 |
| `skills/build-and-test` | hvigorw 构建命令与成功/失败判断标准 |

### Instructions — 全局规则

`.github/instructions/harmony-dev-rules.instructions.md` 自动注入所有 Agent，强制执行：
- 项目结构约束（`bundleName`、目录路径）
- API 查询必须走 MCP（禁止依赖过期静态文档）
- 构建顺序与成功标准
- ArkTS vs TypeScript 差异规则（50+ 条禁止项与替代方案）
- 测试规范、文档生成规范

---

## 快速开始

### 前置要求

- **VS Code** 1.99+，安装 [GitHub Copilot](https://marketplace.visualstudio.com/items?itemName=GitHub.copilot) 扩展
- **Go** 1.21+（如需重新编译 MCP Server）
- **hvigorw**（鸿蒙构建工具，随 DevEco Studio 安装）
- **hdc**（鸿蒙设备连接工具）
- 已连接的 HarmonyOS 真机或模拟器

### 步骤 1：配置 MCP Server

在 VS Code 的 `settings.json` 或项目的 `.vscode/mcp.json` 中添加：

```json
{
  "mcp": {
    "servers": {
      "harmony-docs": {
        "type": "stdio",
        "command": "${workspaceFolder}/mcp-servers/harmony-docs/harmony-docs.exe"
      }
    }
  }
}
```

> Linux/macOS 用户需先编译：
> ```bash
> cd mcp-servers/harmony-docs
> go build -o harmony-docs .
> ```

### 步骤 2：打开 Copilot Agent 面板

在 VS Code 中按 `Ctrl+Shift+P` → **GitHub Copilot: Open Chat**（或侧边栏 Copilot 图标），切换到 **Agent 模式**（`@` 符号）。

### 步骤 3：启动移植

在 Copilot Chat 中输入：

```
选择port-to-harmony的agent 我想移植 OkHttp 3.14.9，源码在 /path/to/okhttp，目标是 HarmonyOS API 12+
```

Agent 会依次：
1. 确认库信息
2. 自动 handoff 给 `planner` 分析架构
3. `analyzer` 用 MCP 查询所有 API 对应关系
4. `documenter` 生成两份移植文档
5. `migrator` 逐模块迁移代码到 `Template/`
6. `builder` 执行构建验证循环

### 步骤 4：单独调用某个 Agent

也可以跳过入口直接调用特定阶段：

```
@analyzer 分析 OkHttp 中的 OkHttpClient 和 Request 类，查询鸿蒙对应 API
@builder  请执行完整构建验证
@migrator 迁移 network 模块的代码
```

---

## 移植产物

每次移植结束后，`Template/` 目录下会生成：

```
Template/
├── 三方库规格.md          # Android/OH 双侧接口规格对照
├── 方案设计.md            # 架构设计、迁移决策、测试方案
├── library/src/main/ets/  # 移植后的 ArkTS 库代码
├── entry/src/main/ets/pages/Index.ets  # Demo 示例
└── entry/src/ohosTest/    # hypium 测试用例
```

---

## 注意事项

- `Template/AppScope/app.json5` 中的 `bundleName: "com.example.template"` **不可修改**
- `oh-package.json5` 文件**不允许修改**
- 所有 HarmonyOS API 查询必须通过 MCP 进行，禁止依赖本仓库内的静态文档（已过期）
- 构建工具为 `hvigorw`，非 Gradle 或 npm 