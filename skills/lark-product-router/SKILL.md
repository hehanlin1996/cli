---
name: lark-product-router
version: 1.0.0
description: "飞书/Lark 产品路由指南：当用户的问题涉及飞书产品、lark-cli 能力边界、应该使用哪个 lark-* skill、或需要官方文档依据时，先用本 skill 定位正确产品域、CLI 命令、领域 skill 和 OpenAPI 文档入口。"
metadata:
  requires:
    bins: ["lark-cli"]
---

# Lark Product Router

> **前置条件：** 先阅读 [`../lark-shared/SKILL.md`](../lark-shared/SKILL.md) 了解认证、身份切换、权限不足和安全规则。

本 skill 是飞书产品问题的入口路由器。它不直接替代领域 skill，而是先把用户问题分到正确产品域，再交给对应 `lark-*` skill 或 `lark-openapi-explorer` 处理，避免凭记忆猜测能力、scope 和 API 路径。

## 路由原则

1. **先定产品域，再定工具**：不要只按关键词调用命令，先判断用户到底在问消息、云文档、多维表格、审批、日历还是开放平台能力。
2. **CLI 优先，官方文档兜底**：已有 shortcut 或领域 skill 时优先使用；没有封装或能力边界不清时，交给 `lark-openapi-explorer` 查官方 OpenAPI 文档。
3. **区分身份边界**：涉及个人资源、云空间、邮箱、日历时优先考虑 `--as user`；涉及应用消息、机器人、事件订阅时优先考虑 `--as bot`，但以具体 skill 说明为准。
4. **不猜测参数和 scope**：执行原生 API 前必须先看 `lark-cli schema` 或官方文档；遇到权限错误按 `lark-shared` 的权限不足处理流程补授权。

## 产品域路由表

| 用户问题类型 | 优先 skill | CLI 探查命令 | 官方文档兜底 |
|---|---|---|---|
| 安装、升级、认证、profile、scope、`_notice` | `lark-shared` | `lark-cli config --help`; `lark-cli auth --help`; `lark-cli update --help` | `lark-openapi-explorer` |
| 消息发送、群聊、话题、消息搜索、表情、图片文件下载 | `lark-im` | `lark-cli im --help` | `https://open.feishu.cn/llms.txt` 的 IM / Messenger 模块 |
| 云文档 Docx 创建、读取、更新、评论、权限 | `lark-doc` 或 `lark-drive` | `lark-cli docs --help`; `lark-cli drive --help` | Docs / Drive / CCM 模块 |
| 多维表格 Base、字段、记录、视图、表单、仪表盘、自动化流程 | `lark-base` | `lark-cli base --help` | Bitable / Base 模块 |
| 电子表格读写、导出、工作表管理 | `lark-sheets` | `lark-cli sheets --help` | Sheets 模块 |
| 幻灯片创建、页面读写、素材替换 | `lark-slides` | `lark-cli slides --help` | Slides 模块 |
| 知识库空间、节点、Wiki 文档组织 | `lark-wiki` | `lark-cli wiki --help` | Wiki / Drive 模块 |
| 云空间文件上传、下载、导出、搜索、权限申请 | `lark-drive` | `lark-cli drive --help` | Drive 模块 |
| 日程、忙闲、会议室、参会人、日历邀请 | `lark-calendar` | `lark-cli calendar --help` | Calendar 模块 |
| 飞书任务、任务清单、关注人、提醒、任务评论 | `lark-task` | `lark-cli task --help` | Task 模块 |
| 邮箱邮件读写、草稿、模板、线程、监听 | `lark-mail` | `lark-cli mail --help` | Mail 模块 |
| 通讯录用户搜索、open_id 解析、部门信息 | `lark-contact` | `lark-cli contact --help` | Contact 模块 |
| 视频会议历史记录、妙记产物、会议录制 | `lark-vc` 或 `lark-minutes` | `lark-cli vc --help`; `lark-cli minutes --help` | VC / Minutes 模块 |
| 让机器人加入或离开正在进行的视频会议 | `lark-vc-agent` | 先读 `skills/lark-vc-agent/SKILL.md`；会议记录查询仍用 `lark-cli vc --help` | VC 模块 |
| 审批实例、审批任务、同意、拒绝、转交 | `lark-approval` | `lark-cli approval --help` | Approval 模块 |
| OKR 目标、关键结果、进展、对齐关系 | `lark-okr` | `lark-cli okr --help` | OKR 模块 |
| 考勤打卡记录 | `lark-attendance` | `lark-cli attendance --help` | Attendance 模块 |
| Markdown 云文件 | `lark-markdown` | `lark-cli markdown --help` | Drive / Docs 模块 |
| 白板、流程图、结构图、可视化 DSL | `lark-whiteboard` | `lark-cli whiteboard --help` | Whiteboard / Board 模块 |
| 实时事件、WebSocket 订阅、事件消费 | `lark-event` | `lark-cli event --help` | Event / Callback 模块 |
| 自定义 skill 封装 lark-cli 流程 | `lark-skill-maker` | `lark-cli --help` | 先读现有领域 skill |
| CLI 没有封装、需要确认底层 OpenAPI | `lark-openapi-explorer` | `lark-cli schema <service.resource.method>`; `lark-cli api --help` | `https://open.feishu.cn/llms.txt` 或 `https://open.larksuite.com/llms.txt` |

## 执行流程

### Step 1：把用户问题改写成产品域

提取用户想操作的资源、动作、身份和输出。例如：

| 用户说法 | 产品域判断 |
|---|---|
| "给某个用户开文档编辑权限" | 云文档 + 云空间权限，优先 `lark-doc` / `lark-drive` |
| "机器人在群里收不到 @ 事件" | IM + 事件订阅 + bot 身份，优先 `lark-im` / `lark-event` |
| "Base 附件写入成功但实际没有" | 多维表格字段/记录/附件，优先 `lark-base` |
| "创建日程并找空会议室" | 日历 + 会议室，优先 `lark-calendar` |

### Step 2：检查领域 skill 和 CLI help

先读对应 skill 的 `SKILL.md`，再用 help 确认当前 CLI 是否有现成 shortcut：

```bash
lark-cli <service> --help
lark-cli <service> +<shortcut> --help
```

如果 help 中没有对应命令，不要直接编造参数；转入 Step 4。

### Step 3：按领域 skill 执行

领域 skill 已覆盖时，按它的前置条件、身份选择、scope 和验证步骤执行。涉及写入或删除时，遵守 `lark-shared` 的风险确认规则。

### Step 4：官方文档兜底

当 CLI 没有封装、能力边界不清、或用户问的是平台能力限制时，调用 `lark-openapi-explorer`：

1. 从 `https://open.feishu.cn/llms.txt` 定位模块文档。
2. 获取具体 API 文档，确认方法、路径、参数、scope、限制和错误码。
3. 如果 API 可用但 CLI 未封装，可用 `lark-cli api` 裸调。
4. 如果官方文档没有能力或限制明确，直接告诉用户"当前不能仅靠 lark-cli 解决"，并列出需要平台或 OpenAPI 团队确认的问题。

## 输出规范

向用户回复时包含：

- **推荐路径**：使用哪个 `lark-*` skill 或 CLI 命令。
- **身份与权限**：建议 `--as user` 还是 `--as bot`，缺哪些 scope。
- **能力边界**：CLI 已支持、可用原生 API、还是需要平台能力确认。
- **下一步命令**：可执行的最小命令或需要用户补充的信息。

## 参考

- [lark-shared](../lark-shared/SKILL.md) — 认证、身份切换、scope 和安全规则
- [lark-openapi-explorer](../lark-openapi-explorer/SKILL.md) — 官方 OpenAPI 文档挖掘和 `lark-cli api` 裸调
