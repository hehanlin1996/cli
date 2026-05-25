# Feedback Autofix Issue Pool Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `lark-workflow-feedback-autofix` evaluate every issue created on the processing date in both `hehanlin1996/cli` and `larksuite/cli`, not only the issue created by the current feedback run.

**Architecture:** Add two focused GitHub helper scripts: one builds a normalized daily issue pool, the other searches PRs in both repositories for explicit issue references and conservative weak matches. Update the Skill workflow and prompts so issue triage, PR matching, and autofix feasibility operate on the issue pool.

**Tech Stack:** Node.js CommonJS scripts, `gh` CLI, existing Skill script test runner, Markdown Skill docs.

---

## File Structure

- Create `skills/lark-workflow-feedback-autofix/scripts/list-daily-issues.js`: query both GitHub repositories for `created:<date> type:issue` and output normalized `issue-pool.json`.
- Create `skills/lark-workflow-feedback-autofix/scripts/search-issue-prs.js`: read `issue-pool.json`, search PRs in both repositories, classify explicit and weak PR matches.
- Modify `skills/lark-workflow-feedback-autofix/scripts/test-github-commands.js`: add tests for daily issue listing and PR matching helper behavior.
- Modify `skills/lark-workflow-feedback-autofix/SKILL.md`: change the workflow from “拉取今日下游 issue” to “生成两仓当日 issue 池并对池内所有 issue 做判断”.
- Modify `skills/lark-workflow-feedback-autofix/prompts/github-issue-triage.md`: clarify input is an item from `issue-pool.json`.
- Modify `skills/lark-workflow-feedback-autofix/prompts/autofix-feasibility.md`: clarify upstream issue can be the target URL while writes still go to `hehanlin1996/cli`.
- Modify `skills/lark-workflow-feedback-autofix/prompts/daily-report.md`: require issue pool coverage, PR match results, and skipped-by-existing-PR summaries.

## Task 1: Add Daily Issue Pool Script

**Files:**
- Create: `/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/list-daily-issues.js`
- Modify: `/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/test-github-commands.js`

- [ ] **Step 1: Write the failing tests**

In `test-github-commands.js`, add this import block after the existing imports:

```js
const {
  DAILY_ISSUE_REPOS,
  buildDailyIssueListArgs,
  buildIssuePool,
  normalizeIssue
} = require("./list-daily-issues.js");
```

Add these assertions before `console.log("github command checks passed");`:

```js
assert.deepStrictEqual(DAILY_ISSUE_REPOS, [
  { repo: "hehanlin1996/cli", origin: "downstream_created_today" },
  { repo: "larksuite/cli", origin: "upstream_created_today" }
]);

assert.deepStrictEqual(buildDailyIssueListArgs({
  repo: "larksuite/cli",
  date: "2026-05-25",
  limit: "100"
}), [
  "issue", "list",
  "--repo", "larksuite/cli",
  "--state", "all",
  "--search", "created:2026-05-25 type:issue",
  "--limit", "100",
  "--json", "number,title,body,state,labels,url,createdAt,updatedAt"
]);

assert.deepStrictEqual(normalizeIssue({
  repo: "larksuite/cli",
  origin: "upstream_created_today"
}, {
  number: 42,
  title: "问题标题",
  body: "问题正文",
  state: "OPEN",
  labels: [{ name: "bug" }],
  url: "https://github.com/larksuite/cli/issues/42",
  createdAt: "2026-05-25T01:00:00Z",
  updatedAt: "2026-05-25T02:00:00Z"
}), {
  source_repo: "larksuite/cli",
  number: 42,
  url: "https://github.com/larksuite/cli/issues/42",
  title: "问题标题",
  body: "问题正文",
  state: "OPEN",
  labels: ["bug"],
  createdAt: "2026-05-25T01:00:00Z",
  updatedAt: "2026-05-25T02:00:00Z",
  origin: "upstream_created_today"
});

assert.deepStrictEqual(buildIssuePool({
  date: "2026-05-25",
  repoResults: [
    {
      repo: "hehanlin1996/cli",
      origin: "downstream_created_today",
      issues: [{
        number: 20,
        title: "下游问题",
        body: "正文",
        state: "OPEN",
        labels: [],
        url: "https://github.com/hehanlin1996/cli/issues/20",
        createdAt: "2026-05-25T01:00:00Z",
        updatedAt: "2026-05-25T01:00:00Z"
      }]
    },
    {
      repo: "larksuite/cli",
      origin: "upstream_created_today",
      issues: [{
        number: 1051,
        title: "上游问题",
        body: "正文",
        state: "OPEN",
        labels: [{ name: "enhancement" }],
        url: "https://github.com/larksuite/cli/issues/1051",
        createdAt: "2026-05-25T02:00:00Z",
        updatedAt: "2026-05-25T02:00:00Z"
      }]
    }
  ]
}), {
  date: "2026-05-25",
  complete: true,
  repos: ["hehanlin1996/cli", "larksuite/cli"],
  issues: [
    {
      source_repo: "hehanlin1996/cli",
      number: 20,
      url: "https://github.com/hehanlin1996/cli/issues/20",
      title: "下游问题",
      body: "正文",
      state: "OPEN",
      labels: [],
      createdAt: "2026-05-25T01:00:00Z",
      updatedAt: "2026-05-25T01:00:00Z",
      origin: "downstream_created_today"
    },
    {
      source_repo: "larksuite/cli",
      number: 1051,
      url: "https://github.com/larksuite/cli/issues/1051",
      title: "上游问题",
      body: "正文",
      state: "OPEN",
      labels: ["enhancement"],
      createdAt: "2026-05-25T02:00:00Z",
      updatedAt: "2026-05-25T02:00:00Z",
      origin: "upstream_created_today"
    }
  ],
  errors: []
});
```

- [ ] **Step 2: Run the failing test**

Run:

```bash
node /Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/test-github-commands.js
```

Expected: FAIL with `Cannot find module './list-daily-issues.js'`.

- [ ] **Step 3: Implement `list-daily-issues.js`**

Create `/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/list-daily-issues.js` with:

```js
// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

"use strict";

const { spawnSync } = require("child_process");
const { normalizeDate, runMain, writeJson } = require("./common.js");

const DAILY_ISSUE_REPOS = [
  { repo: "hehanlin1996/cli", origin: "downstream_created_today" },
  { repo: "larksuite/cli", origin: "upstream_created_today" }
];

const ISSUE_JSON_FIELDS = "number,title,body,state,labels,url,createdAt,updatedAt";

function buildDailyIssueListArgs(options) {
  const date = normalizeDate(options.date);
  return [
    "issue", "list",
    "--repo", options.repo,
    "--state", "all",
    "--search", `created:${date} type:issue`,
    "--limit", String(options.limit || "100"),
    "--json", ISSUE_JSON_FIELDS
  ];
}

function labelNames(labels) {
  if (!Array.isArray(labels)) return [];
  return labels.map((label) => {
    if (label && typeof label === "object" && label.name) return String(label.name);
    return String(label || "");
  }).filter(Boolean);
}

function normalizeIssue(repoConfig, issue) {
  return {
    source_repo: repoConfig.repo,
    number: issue.number,
    url: issue.url,
    title: issue.title || "",
    body: issue.body || "",
    state: issue.state || "",
    labels: labelNames(issue.labels),
    createdAt: issue.createdAt || "",
    updatedAt: issue.updatedAt || "",
    origin: repoConfig.origin
  };
}

function buildIssuePool(options) {
  const date = normalizeDate(options.date);
  const repoResults = options.repoResults || [];
  const errors = [];
  const issues = [];

  for (const result of repoResults) {
    if (result.error) {
      errors.push({
        repo: result.repo,
        error: result.error
      });
      continue;
    }
    for (const issue of result.issues || []) {
      issues.push(normalizeIssue({
        repo: result.repo,
        origin: result.origin
      }, issue));
    }
  }

  return {
    date,
    complete: errors.length === 0,
    repos: DAILY_ISSUE_REPOS.map((item) => item.repo),
    issues,
    errors
  };
}

function runGh(args) {
  const result = spawnSync("gh", args, { encoding: "utf8" });
  if (result.status !== 0) {
    throw new Error(result.stderr || result.stdout || `gh exited ${result.status}`);
  }
  return JSON.parse(result.stdout || "[]");
}

async function main(args) {
  const date = normalizeDate(args.date);
  const repoResults = DAILY_ISSUE_REPOS.map((repoConfig) => {
    try {
      return {
        ...repoConfig,
        issues: runGh(buildDailyIssueListArgs({
          repo: repoConfig.repo,
          date,
          limit: args.limit || "100"
        }))
      };
    } catch (err) {
      return {
        ...repoConfig,
        error: err && err.message ? err.message : String(err)
      };
    }
  });
  writeJson(buildIssuePool({ date, repoResults }));
}

if (require.main === module) runMain(main);

module.exports = {
  DAILY_ISSUE_REPOS,
  buildDailyIssueListArgs,
  buildIssuePool,
  labelNames,
  normalizeIssue
};
```

- [ ] **Step 4: Run the test**

Run:

```bash
node /Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/test-github-commands.js
```

Expected: PASS with `github command checks passed`.

- [ ] **Step 5: Record installed Skill change**

These Skill files live under `/Users/bytedance/.agents/skills`, outside the `larksuite/cli` Git worktree. Do not run `git add` for these installed Skill files from the repo. Record the changed file paths in the final implementation summary and rely on the Skill self-test in Task 5 for verification.

## Task 2: Add PR Match Script

**Files:**
- Create: `/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/search-issue-prs.js`
- Modify: `/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/test-github-commands.js`

- [ ] **Step 1: Write the failing tests**

In `test-github-commands.js`, add this import block after the `list-daily-issues.js` import:

```js
const {
  PR_SEARCH_REPOS,
  buildPrSearchArgs,
  classifyPrMatch,
  issueReferenceTokens,
  buildIssuePrMatches
} = require("./search-issue-prs.js");
```

Add these assertions before `console.log("github command checks passed");`:

```js
const upstreamIssue = {
  source_repo: "larksuite/cli",
  number: 1051,
  url: "https://github.com/larksuite/cli/issues/1051",
  title: "支持 card.action.trigger 回调",
  body: "希望 event consume 支持 card.action.trigger"
};

assert.deepStrictEqual(PR_SEARCH_REPOS, ["hehanlin1996/cli", "larksuite/cli"]);
assert.deepStrictEqual(issueReferenceTokens(upstreamIssue), [
  "https://github.com/larksuite/cli/issues/1051",
  "larksuite/cli#1051"
]);
assert.deepStrictEqual(buildPrSearchArgs({
  repo: "hehanlin1996/cli",
  query: "larksuite/cli#1051",
  limit: "50"
}), [
  "pr", "list",
  "--repo", "hehanlin1996/cli",
  "--state", "all",
  "--search", "larksuite/cli#1051",
  "--limit", "50",
  "--json", "number,title,body,state,url,createdAt,updatedAt,comments"
]);

assert.deepStrictEqual(classifyPrMatch({
  issue: upstreamIssue,
  prRepo: "hehanlin1996/cli",
  pr: {
    number: 12,
    title: "修复回调消费",
    body: "fixes larksuite/cli#1051",
    url: "https://github.com/hehanlin1996/cli/pull/12"
  }
}), {
  repo: "hehanlin1996/cli",
  number: 12,
  url: "https://github.com/hehanlin1996/cli/pull/12",
  match_type: "fixes",
  evidence: "PR text contains closing reference for larksuite/cli#1051"
});

assert.deepStrictEqual(classifyPrMatch({
  issue: upstreamIssue,
  prRepo: "larksuite/cli",
  pr: {
    number: 13,
    title: "修复回调消费",
    body: "关联 #1051",
    url: "https://github.com/larksuite/cli/pull/13"
  }
}), {
  repo: "larksuite/cli",
  number: 13,
  url: "https://github.com/larksuite/cli/pull/13",
  match_type: "matched",
  evidence: "PR text contains same-repo issue reference #1051"
});

assert.strictEqual(classifyPrMatch({
  issue: upstreamIssue,
  prRepo: "hehanlin1996/cli",
  pr: {
    number: 14,
    title: "支持 card action trigger",
    body: "没有明确引用 issue",
    url: "https://github.com/hehanlin1996/cli/pull/14"
  }
}), null);

assert.deepStrictEqual(buildIssuePrMatches({
  date: "2026-05-25",
  issuePool: { issues: [upstreamIssue] },
  prResults: [
    {
      issue_url: upstreamIssue.url,
      issue: upstreamIssue,
      repo: "hehanlin1996/cli",
      prs: [{
        number: 12,
        title: "修复回调消费",
        body: "fixes larksuite/cli#1051",
        url: "https://github.com/hehanlin1996/cli/pull/12"
      }]
    }
  ]
}), {
  date: "2026-05-25",
  items: [{
    issue_url: "https://github.com/larksuite/cli/issues/1051",
    source_repo: "larksuite/cli",
    has_fix_pr: true,
    fix_prs: [{
      repo: "hehanlin1996/cli",
      number: 12,
      url: "https://github.com/hehanlin1996/cli/pull/12",
      match_type: "fixes",
      evidence: "PR text contains closing reference for larksuite/cli#1051"
    }],
    similar_prs: [],
    errors: []
  }]
});
```

- [ ] **Step 2: Run the failing test**

Run:

```bash
node /Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/test-github-commands.js
```

Expected: FAIL with `Cannot find module './search-issue-prs.js'`.

- [ ] **Step 3: Implement `search-issue-prs.js`**

Create `/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/search-issue-prs.js` with:

```js
// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

"use strict";

const { spawnSync } = require("child_process");
const { readJsonFile, runMain, writeJson } = require("./common.js");

const PR_SEARCH_REPOS = ["hehanlin1996/cli", "larksuite/cli"];
const PR_JSON_FIELDS = "number,title,body,state,url,createdAt,updatedAt,comments";

function buildPrSearchArgs(options) {
  return [
    "pr", "list",
    "--repo", options.repo,
    "--state", "all",
    "--search", String(options.query || ""),
    "--limit", String(options.limit || "50"),
    "--json", PR_JSON_FIELDS
  ];
}

function issueReferenceTokens(issue, prRepo) {
  const tokens = [
    issue.url,
    `${issue.source_repo}#${issue.number}`
  ].filter(Boolean);
  if (prRepo && prRepo === issue.source_repo) {
    tokens.push(`#${issue.number}`);
  }
  return [...new Set(tokens)];
}

function prText(pr) {
  const comments = Array.isArray(pr.comments)
    ? pr.comments.map((comment) => comment.body || "").join("\n")
    : "";
  return [pr.title || "", pr.body || "", comments].join("\n");
}

function closingReferenceRegex(issue) {
  const fullRef = `${issue.source_repo}#${issue.number}`.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const urlRef = String(issue.url || "").replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  return new RegExp(`\\b(fix|fixes|fixed|close|closes|closed|resolve|resolves|resolved)\\b[\\s\\S]{0,80}(${fullRef}|${urlRef})`, "i");
}

function classifyPrMatch(options) {
  const issue = options.issue;
  const prRepo = options.prRepo;
  const pr = options.pr;
  const text = prText(pr);
  const fullRef = `${issue.source_repo}#${issue.number}`;

  if (closingReferenceRegex(issue).test(text)) {
    return {
      repo: prRepo,
      number: pr.number,
      url: pr.url,
      match_type: "fixes",
      evidence: `PR text contains closing reference for ${fullRef}`
    };
  }
  if (issue.url && text.includes(issue.url)) {
    return {
      repo: prRepo,
      number: pr.number,
      url: pr.url,
      match_type: "matched",
      evidence: `PR text contains issue URL ${issue.url}`
    };
  }
  if (text.includes(fullRef)) {
    return {
      repo: prRepo,
      number: pr.number,
      url: pr.url,
      match_type: "matched",
      evidence: `PR text contains issue reference ${fullRef}`
    };
  }
  if (prRepo === issue.source_repo && text.includes(`#${issue.number}`)) {
    return {
      repo: prRepo,
      number: pr.number,
      url: pr.url,
      match_type: "matched",
      evidence: `PR text contains same-repo issue reference #${issue.number}`
    };
  }
  return null;
}

function weakSimilarity(issue, pr) {
  const issueWords = new Set(String(`${issue.title} ${issue.body}`).toLowerCase().split(/[^\p{L}\p{N}.+-]+/u).filter((word) => word.length >= 4));
  const prWords = new Set(String(`${pr.title} ${pr.body}`).toLowerCase().split(/[^\p{L}\p{N}.+-]+/u).filter((word) => word.length >= 4));
  let overlap = 0;
  for (const word of issueWords) {
    if (prWords.has(word)) overlap += 1;
  }
  return overlap >= 3;
}

function dedupePrMatches(matches) {
  const seen = new Set();
  const out = [];
  for (const match of matches) {
    const key = `${match.repo}#${match.number}:${match.match_type}`;
    if (seen.has(key)) continue;
    seen.add(key);
    out.push(match);
  }
  return out;
}

function buildIssuePrMatches(options) {
  const items = [];
  const issuePool = options.issuePool || { issues: [] };
  const prResults = options.prResults || [];

  for (const issue of issuePool.issues || []) {
    const fixPrs = [];
    const similarPrs = [];
    const errors = [];
    const matchingResults = prResults.filter((result) => result.issue_url === issue.url);

    for (const result of matchingResults) {
      if (result.error) {
        errors.push({ repo: result.repo, error: result.error });
        continue;
      }
      for (const pr of result.prs || []) {
        const match = classifyPrMatch({ issue, prRepo: result.repo, pr });
        if (match) {
          fixPrs.push(match);
        } else if (weakSimilarity(issue, pr)) {
          similarPrs.push({
            repo: result.repo,
            number: pr.number,
            url: pr.url,
            match_type: "similar",
            evidence: "PR title/body shares multiple terms with issue but has no explicit reference"
          });
        }
      }
    }

    const dedupedFixPrs = dedupePrMatches(fixPrs);
    items.push({
      issue_url: issue.url,
      source_repo: issue.source_repo,
      has_fix_pr: dedupedFixPrs.length > 0,
      fix_prs: dedupedFixPrs,
      similar_prs: dedupePrMatches(similarPrs),
      errors
    });
  }

  return {
    date: options.date,
    items
  };
}

function runGh(args) {
  const result = spawnSync("gh", args, { encoding: "utf8" });
  if (result.status !== 0) {
    throw new Error(result.stderr || result.stdout || `gh exited ${result.status}`);
  }
  return JSON.parse(result.stdout || "[]");
}

async function main(args) {
  const issuePool = readJsonFile(args.input);
  const prResults = [];
  const limit = args.limit || "50";

  for (const issue of issuePool.issues || []) {
    for (const repo of PR_SEARCH_REPOS) {
      const prsByKey = new Map();
      try {
        for (const token of issueReferenceTokens(issue, repo)) {
          for (const pr of runGh(buildPrSearchArgs({ repo, query: token, limit }))) {
            prsByKey.set(`${repo}#${pr.number}`, pr);
          }
        }
        prResults.push({
          issue_url: issue.url,
          issue,
          repo,
          prs: [...prsByKey.values()]
        });
      } catch (err) {
        prResults.push({
          issue_url: issue.url,
          issue,
          repo,
          error: err && err.message ? err.message : String(err)
        });
      }
    }
  }

  writeJson(buildIssuePrMatches({
    date: issuePool.date,
    issuePool,
    prResults
  }));
}

if (require.main === module) runMain(main);

module.exports = {
  PR_SEARCH_REPOS,
  buildIssuePrMatches,
  buildPrSearchArgs,
  classifyPrMatch,
  issueReferenceTokens,
  prText,
  weakSimilarity
};
```

- [ ] **Step 4: Run the test**

Run:

```bash
node /Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/test-github-commands.js
```

Expected: PASS with `github command checks passed`.

- [ ] **Step 5: Record installed Skill change**

These Skill files live under `/Users/bytedance/.agents/skills`, outside the `larksuite/cli` Git worktree. Do not run `git add` for these installed Skill files from the repo. Record the changed file paths in the final implementation summary and rely on the Skill self-test in Task 5 for verification.

## Task 3: Update Workflow Documentation and Prompts

**Files:**
- Modify: `/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/SKILL.md`
- Modify: `/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/prompts/github-issue-triage.md`
- Modify: `/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/prompts/autofix-feasibility.md`
- Modify: `/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/prompts/daily-report.md`

- [ ] **Step 1: Update `SKILL.md` workflow**

In `SKILL.md`, replace the total workflow list items 7-10 with:

```markdown
7. 主控生成当日 issue 池：读取 `hehanlin1996/cli` 和 `larksuite/cli` 中 `created:<处理日期> type:issue` 的所有 issue。
8. subagent 针对当日 issue 池逐条判断问题类型。
9. 主控针对当日 issue 池查询 `hehanlin1996/cli` 和 `larksuite/cli` 中是否已有 PR 明确修复。
10. subagent 对进入问题流且没有强匹配修复 PR 的 issue 判断可修性。
11. 低风险问题用隔离 worktree 修复。`larksuite/cli` issue 可作为只读修复目标 URL，但分支和 PR 仍只写 `hehanlin1996/cli`。
12. 完成测试和审查后，在 `hehanlin1996/cli` 创建中文 PR。
13. 新建或更新飞书云文档中文日报。
```

In the execution command section, replace the existing downstream issue list command with:

```bash
node skills/lark-workflow-feedback-autofix/scripts/list-daily-issues.js --date 2026-05-25 > skills/lark-workflow-feedback-autofix/.feedback-autofix/runs/2026-05-25/issue-pool.json
node skills/lark-workflow-feedback-autofix/scripts/search-issue-prs.js --input skills/lark-workflow-feedback-autofix/.feedback-autofix/runs/2026-05-25/issue-pool.json > skills/lark-workflow-feedback-autofix/.feedback-autofix/runs/2026-05-25/issue-pr-match-results.json
```

Add this paragraph after those commands:

```markdown
`issue-pool.json` 是后续 GitHub issue 分类、可修性判断、已有 PR 修复判断的唯一输入集合。不得只处理本轮刚创建的 issue。`larksuite/cli` issue 只读；如果可修，PR 仍创建到 `hehanlin1996/cli`，并在 PR 正文中关联上游 issue URL。
```

- [ ] **Step 2: Update `github-issue-triage.md`**

Replace the first two paragraphs with:

```markdown
# 今日 GitHub issue 分类子代理

你负责判断 `issue-pool.json` 中的当日新增 issue 是否属于 `bug`、`cli-improvement`、`product-gap`，或应该归入过滤分类。输入可能来自 `hehanlin1996/cli` 或 `larksuite/cli`，但你只能返回 JSON 判断，不直接写任何 GitHub 内容。

判断时优先看 issue 的用户可见现象、命令、参数、输出、错误信息和复现证据。使用咨询、权限配置、纯文档缺口、平台策略限制和证据不足不进入自动修复。
```

In the JSON example, include `source_repo`:

```json
{
  "agent": "github-issue-triage-agent",
  "issue_url": "string",
  "source_repo": "larksuite/cli",
  "category": "bug",
  "enters_issue_flow": true,
  "confidence": 0.82,
  "summary": "一句话摘要",
  "evidence": ["issue 中的关键证据"],
  "missing_info": [],
  "reason": "中文理由"
}
```

- [ ] **Step 3: Update `autofix-feasibility.md`**

Add this paragraph after the first paragraph:

```markdown
输入 issue 可能来自 `hehanlin1996/cli` 或 `larksuite/cli`。`larksuite/cli` 只能作为只读目标来源；如果判定可修，修复分支和 PR 仍只能写到 `hehanlin1996/cli`，并且 PR 正文必须关联上游 issue URL。
```

Add this field to the JSON example:

```json
  "source_repo": "larksuite/cli",
```

- [ ] **Step 4: Update `daily-report.md` prompt**

Replace the first paragraph with:

```markdown
你负责根据 ledger 生成中文日报 Markdown。日报必须包含数据范围、Base 反馈分类统计、当日 issue 池覆盖范围、当日 issue 分类统计、已建 issue、已有 PR 修复情况、已提 PR、不可自动修原因、需要人工跟进项、失败项和测试结果。
```

Add this sentence after the title rule:

```markdown
当 `issue_pool`、`issue_triage`、`pr_matches` 或 `autofix_targets` 存在时，必须在日报中展示对应汇总；不能只展示本轮新建 issue。
```

- [ ] **Step 5: Run full Skill self-test**

Run:

```bash
node /Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/test.js
```

Expected: PASS with:

```text
skill layout checks passed
schema and common checks passed
normalize feedback checks passed
github command checks passed
lark command checks passed
ledger checks passed
```

- [ ] **Step 6: Record installed Skill documentation change**

These Skill files live under `/Users/bytedance/.agents/skills`, outside the `larksuite/cli` Git worktree. Do not run `git add` for these installed Skill files from the repo. Record the changed file paths in the final implementation summary and rely on the Skill self-test in Task 5 for verification.

## Task 4: Add Ledger and Report Integration Guidance

**Files:**
- Modify: `/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/SKILL.md`
- Modify: `/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/fixtures/ledger-valid.json`
- Modify: `/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/test-ledger.js`

- [ ] **Step 1: Write ledger tests for new optional fields**

In `test-ledger.js`, add assertions that `validateLedger` accepts:

```js
assert.deepStrictEqual(validateLedger({
  run_id: "feedback-autofix-2026-05-25",
  issue_pool: {
    repos: {
      "hehanlin1996/cli": 1,
      "larksuite/cli": 2
    },
    total: 3,
    path: "issue-pool.json"
  },
  issue_triage: {
    path: "issue-triage-results.json",
    categories: { bug: 1, "usage-question": 2 },
    issue_flow_count: 1
  },
  pr_matches: {
    path: "issue-pr-match-results.json",
    strong_match_count: 1,
    weak_match_count: 1
  },
  autofix_targets: {
    path: "autofix-feasibility-results.json",
    autofixable_count: 0,
    manual_count: 1
  },
  created_issues: [],
  created_prs: []
}), []);
```

- [ ] **Step 2: Run ledger test**

Run:

```bash
node /Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/test-ledger.js
```

Expected: PASS. The current validator allows unknown optional objects, so this test documents the ledger shape without requiring validator changes.

- [ ] **Step 3: Update `SKILL.md` ledger section**

Add this JSON block to the ledger section:

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

Add this sentence below the block:

```markdown
如果 `issue-pool.json` 生成失败或某个仓库查询失败，ledger 必须记录失败仓库；日报必须明确“覆盖不完整”，不得把部分结果当成完整覆盖。
```

- [ ] **Step 4: Run full Skill self-test**

Run:

```bash
node /Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/test.js
```

Expected: PASS.

- [ ] **Step 5: Record installed Skill ledger guidance change**

These Skill files live under `/Users/bytedance/.agents/skills`, outside the `larksuite/cli` Git worktree. Do not run `git add` for these installed Skill files from the repo. Record the changed file paths in the final implementation summary and rely on the Skill self-test in Task 5 for verification.

## Task 5: End-to-End Smoke Verification

**Files:**
- No source changes.
- Writes runtime artifacts only under `/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/.feedback-autofix/runs/<date>/`.

- [ ] **Step 1: Run complete Skill self-test**

Run:

```bash
node /Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/test.js
```

Expected: PASS with all six checks.

- [ ] **Step 2: Generate issue pool for a known date**

Use the fixed smoke-test date `2026-05-24`:

```bash
mkdir -p /Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/.feedback-autofix/runs/2026-05-24
node /Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/list-daily-issues.js --date 2026-05-24 > /Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/.feedback-autofix/runs/2026-05-24/issue-pool.json
```

Expected: command exits 0 and output JSON contains:

```json
{
  "date": "2026-05-24",
  "repos": ["hehanlin1996/cli", "larksuite/cli"],
  "issues": []
}
```

The `issues` array is live GitHub data, so its length is not fixed. The required assertion is that both repositories appear in `repos` and every issue has `source_repo` and `origin`.

- [ ] **Step 3: Search PR matches for the issue pool**

Run:

```bash
node /Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/scripts/search-issue-prs.js --input /Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/.feedback-autofix/runs/2026-05-24/issue-pool.json > /Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/.feedback-autofix/runs/2026-05-24/issue-pr-match-results.json
```

Expected: command exits 0 and output JSON contains `date` and `items`. If `issue-pool.json` has issues, every item has `issue_url`, `source_repo`, `has_fix_pr`, `fix_prs`, `similar_prs`, and `errors`.

- [ ] **Step 4: Inspect generated artifacts**

Run:

```bash
node -e 'const fs=require("fs"); const pool=JSON.parse(fs.readFileSync("/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/.feedback-autofix/runs/2026-05-24/issue-pool.json","utf8")); const prs=JSON.parse(fs.readFileSync("/Users/bytedance/.agents/skills/lark-workflow-feedback-autofix/.feedback-autofix/runs/2026-05-24/issue-pr-match-results.json","utf8")); console.log(JSON.stringify({repos:pool.repos,total:pool.issues.length,prItems:prs.items.length,complete:pool.complete}, null, 2));'
```

Expected: JSON prints both repos, an issue count, a PR match item count, and a boolean `complete`.

- [ ] **Step 5: Final git status**

Run:

```bash
git status --short --branch
```

Expected: only intentional committed changes remain; runtime artifacts under `.feedback-autofix` are outside the repo worktree.
