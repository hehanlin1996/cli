# lark-cli 反馈自动修复：当日 Issue 池设计

## 背景

`lark-workflow-feedback-autofix` 现在会从固定飞书 Base 拉取每日反馈、分类、查重，并在 `hehanlin1996/cli` 创建缺失 issue。当前流程在创建 issue 后，后续“是否是问题、是否能修、是否已有 PR 修复”的判断容易只围绕本轮新建 issue 展开，遗漏当天已经存在的 GitHub issue。

本次设计把“当日新增 issue”提升为一等输入，确保后续 triage、可修性判断和 PR 查重覆盖完整的当日 issue 池。

## 目标

- 处理日期为 `YYYY-MM-DD` 时，收集两个仓库中 `created:<日期> type:issue` 的所有 issue：
  - `hehanlin1996/cli`
  - `larksuite/cli`
- 后续问题分类、可修性判断、已有 PR 修复判断全部基于统一的 `issue-pool.json`。
- 对 `larksuite/cli` issue 保持只读；如果可修，直接以该上游 issue URL 作为修复目标，在 `hehanlin1996/cli` 分支上修复并创建 PR。
- 对 `hehanlin1996/cli` issue 保持现有写入边界。
- 写入 GitHub issue、PR 和飞书日报必须使用中文；`hehanlin1996/cli` issue 不依赖 label/tag。
- 修 bug、改代码、生成 PR 前必须使用 Superpowers，至少包括 TDD 和完成前验证；遇到异常行为时使用 systematic-debugging。

## 非目标

- 不在 `larksuite/cli` 创建 issue、comment、branch 或 PR。
- 不处理处理日期之前创建但当天更新、评论、重新打开的 issue。
- 不实现完整 GitHub 关联图；本次只做明确引用和保守的弱相似匹配。
- 不把弱相似 PR 当作已修复结论。

## 数据流

1. 拉取固定 Base 反馈并标准化为 `feedback-items.jsonl`。
2. 反馈分类、聚合、上游查重。
3. 如需创建下游 issue，主控在 `hehanlin1996/cli` 创建中文 issue。
4. 创建完成后生成 `issue-pool.json`，覆盖两仓当日新增 issue。
5. 对 `issue-pool.json` 中每个 issue 生成 `issue-triage-results.json`。
6. 对每个 issue 同时查 `hehanlin1996/cli` 和 `larksuite/cli` PR，生成 `issue-pr-match-results.json`。
7. 对进入 issue flow 且没有强匹配修复 PR 的 issue 生成 `autofix-feasibility-results.json`。
8. 对低风险、可写 RED 测试的 issue 进入自动修复流程。
9. ledger 和日报汇总 Base 反馈、issue 池覆盖、PR 匹配、可修性和自动修复结果。

## 组件

### `list-daily-issues.js`

新增脚本。输入 `--date YYYY-MM-DD`，输出统一 issue 池 JSON。

查询范围：

- `gh issue list --repo hehanlin1996/cli --search "created:<date> type:issue"`
- `gh issue list --repo larksuite/cli --search "created:<date> type:issue"`

输出结构：

```json
{
  "date": "YYYY-MM-DD",
  "repos": ["hehanlin1996/cli", "larksuite/cli"],
  "issues": [
    {
      "source_repo": "larksuite/cli",
      "number": 123,
      "url": "https://github.com/larksuite/cli/issues/123",
      "title": "issue title",
      "body": "issue body",
      "state": "OPEN",
      "labels": [],
      "createdAt": "2026-05-25T00:00:00Z",
      "updatedAt": "2026-05-25T01:00:00Z",
      "origin": "upstream_created_today"
    }
  ]
}
```

`origin` 取值：

- `downstream_created_today`
- `upstream_created_today`

### `issue-triage-results.json`

主控或子代理对 `issue-pool.json` 中每条 issue 做分类。分类沿用现有反馈分类体系：

- 进入问题流：`bug`、`cli-improvement`、`product-gap`
- 不进入问题流：`usage-question`、`permission-auth`、`docs-gap`、`needs-more-info`、`noise`、`platform-limitation`

每条结果必须带 `issue_url`、`source_repo`、`category`、`enters_issue_flow`、`confidence`、`summary`、`evidence`、`missing_info`、`reason`。

### `search-issue-prs.js`

新增脚本。输入 `issue-pool.json`，对每条 issue 同时查询两个仓库的 PR：

- `hehanlin1996/cli`
- `larksuite/cli`

匹配优先级：

1. 强匹配：PR 标题、正文或评论明确包含 issue URL、`owner/repo#number`，或在同仓上下文中包含 `#number`。
2. 修复匹配：PR 正文包含 `fixes`、`closes`、`resolves` 等关闭语义并引用该 issue。
3. 弱匹配：标题或正文相似但没有明确引用，仅写入 `similar_prs`，不视为已修复。

输出结构：

```json
{
  "date": "YYYY-MM-DD",
  "items": [
    {
      "issue_url": "https://github.com/larksuite/cli/issues/123",
      "source_repo": "larksuite/cli",
      "has_fix_pr": true,
      "fix_prs": [
        {
          "repo": "hehanlin1996/cli",
          "number": 12,
          "url": "https://github.com/hehanlin1996/cli/pull/12",
          "match_type": "fixes",
          "evidence": "PR body contains fixes larksuite/cli#123"
        }
      ],
      "similar_prs": []
    }
  ]
}
```

### `autofix-feasibility-results.json`

只对满足以下条件的 issue 生成：

- 在 `issue-triage-results.json` 中进入 issue flow。
- 在 `issue-pr-match-results.json` 中没有强匹配修复 PR。

安全门沿用现有规则：认证、密钥、token、profile、权限模型、CI、发布、依赖大改、generated metadata、真实 API 写入验证、跨仓库联动、产品决策、需要补充复现信息、没有可写 RED 测试时不可自动修。

## PR 查重规则

强匹配和修复匹配会阻止重复自动修复。弱匹配只进入日报，不阻止可修性判断。

对于 `larksuite/cli` issue：

- 上游 issue 只读。
- 如果已有 `larksuite/cli` 或 `hehanlin1996/cli` 强匹配 PR，日报记录“已有修复 PR”，不重复修。
- 如果无强匹配 PR 且可修，在 `hehanlin1996/cli` 创建修复分支和 PR，PR 正文必须关联上游 issue URL。

对于 `hehanlin1996/cli` issue：

- 如果已有两仓任一强匹配 PR，日报记录，不重复修。
- 如果无强匹配 PR 且可修，沿用现有下游修复流程。

## Ledger

`ledger.json` 新增以下汇总字段：

```json
{
  "issue_pool": {
    "repos": {
      "hehanlin1996/cli": 0,
      "larksuite/cli": 0
    },
    "total": 0,
    "path": "issue-pool.json"
  },
  "issue_triage": {
    "path": "issue-triage-results.json",
    "categories": {},
    "issue_flow_count": 0
  },
  "pr_matches": {
    "path": "issue-pr-match-results.json",
    "strong_match_count": 0,
    "weak_match_count": 0
  },
  "autofix_targets": {
    "path": "autofix-feasibility-results.json",
    "autofixable_count": 0,
    "manual_count": 0
  }
}
```

现有 `created_issues`、`created_prs`、`failures`、`resolved_issues` 保留。

## 日报

日报新增或强化以下章节：

- 当日 issue 覆盖范围：两仓各多少条，总数多少。
- 当日 issue 分类统计：基于 `issue-triage-results.json`。
- 已有 PR 修复情况：强匹配 PR、弱匹配 PR 分开列。
- 自动修复候选：列出为什么可修、为什么不可修、为什么因已有 PR 跳过。
- 本轮 Base 反馈新建 issue 与当日 issue 池的关系。

## 错误处理

- 任一仓库 issue list 失败时，记录失败仓库和错误；不要把部分结果伪装成完整覆盖。
- PR 查询失败时，保留 issue triage 结果，但对应 issue 的 PR 匹配状态标记为 `unknown`，日报列入失败项。
- `hehanlin1996/cli` issue 创建失败时，仍继续生成 issue pool；issue pool 反映目标仓库真实当日 issue 状态。
- 弱匹配 PR 不阻止自动修复，但必须出现在日报里。

## 测试计划

- `list-daily-issues.js` 单元测试：
  - 构造两个仓库的 `gh issue list` 输出，确认合并、`source_repo` 和 `origin` 正确。
  - 一个仓库失败时，确认输出或错误状态明确。
- `search-issue-prs.js` 单元测试：
  - PR body 包含 issue URL 时强匹配。
  - PR body 包含 `fixes owner/repo#number` 时修复匹配。
  - 标题相似但无引用时只进入 `similar_prs`。
  - 同时查询两仓 PR。
- 现有 `test-github-commands.js` 扩展：
  - 验证 issue list 覆盖两个仓库。
  - 验证 downstream issue 创建不带 label。
- Skill 完整自测：
  - `node skills/lark-workflow-feedback-autofix/scripts/test.js`

## 实施约束

- 实现时必须先使用 `superpowers:test-driven-development`，先写失败测试再实现。
- 完成前必须使用 `superpowers:verification-before-completion`。
- 遇到 bug、测试失败或异常行为时必须使用 `superpowers:systematic-debugging`。
- 不在 `larksuite/cli` 写入任何内容。
- PR 和日报输出必须为中文。
