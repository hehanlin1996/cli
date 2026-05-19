// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package task

import (
	"context"
	"testing"
	"time"

	clie2e "github.com/larksuite/cli/tests/cli_e2e"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestTask_TasklistDryRun(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	t.Setenv("LARKSUITE_CLI_APP_ID", "task_dryrun_test")
	t.Setenv("LARKSUITE_CLI_APP_SECRET", "task_dryrun_secret")
	t.Setenv("LARKSUITE_CLI_BRAND", "feishu")

	t.Run("create tasklist with seeded child tasks", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		t.Cleanup(cancel)

		result, err := clie2e.RunCmd(ctx, clie2e.Request{
			Args: []string{
				"task", "+tasklist-create",
				"--name", "Daily",
				"--data", `[{"summary":"child task"}]`,
				"--dry-run",
			},
			DefaultAs: "bot",
		})
		require.NoError(t, err)
		result.AssertExitCode(t, 0)

		out := result.Stdout
		if count := gjson.Get(out, "api.#").Int(); count != 1 {
			t.Fatalf("expected 1 API call, got %d\nstdout:\n%s", count, out)
		}
		if method := gjson.Get(out, "api.0.method").String(); method != "POST" {
			t.Fatalf("api[0].method = %q, want POST\nstdout:\n%s", method, out)
		}
		if url := gjson.Get(out, "api.0.url").String(); url != "/open-apis/task/v2/tasklists" {
			t.Fatalf("api[0].url = %q, want /open-apis/task/v2/tasklists\nstdout:\n%s", url, out)
		}
		if got := gjson.Get(out, "api.0.params.user_id_type").String(); got != "open_id" {
			t.Fatalf("api[0].params.user_id_type = %q, want open_id\nstdout:\n%s", got, out)
		}
		if got := gjson.Get(out, "api.0.body.name").String(); got != "Daily" {
			t.Fatalf("api[0].body.name = %q, want Daily\nstdout:\n%s", got, out)
		}
		if desc := gjson.Get(out, "api.0.desc").String(); desc != "2. Create Tasks within the new tasklist (concurrently)" {
			t.Fatalf("api[0].desc = %q, want seeded child task dry-run description\nstdout:\n%s", desc, out)
		}
	})

	t.Run("add task to tasklist", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		t.Cleanup(cancel)

		result, err := clie2e.RunCmd(ctx, clie2e.Request{
			Args: []string{
				"task", "+tasklist-task-add",
				"--tasklist-id", "tl-123",
				"--task-id", "task-1",
				"--section-guid", "sec-456",
				"--dry-run",
			},
			DefaultAs: "bot",
		})
		require.NoError(t, err)
		result.AssertExitCode(t, 0)

		out := result.Stdout
		if count := gjson.Get(out, "api.#").Int(); count != 1 {
			t.Fatalf("expected 1 API call, got %d\nstdout:\n%s", count, out)
		}
		if method := gjson.Get(out, "api.0.method").String(); method != "POST" {
			t.Fatalf("api[0].method = %q, want POST\nstdout:\n%s", method, out)
		}
		if url := gjson.Get(out, "api.0.url").String(); url != "/open-apis/task/v2/tasks/task-1/add_tasklist" {
			t.Fatalf("api[0].url = %q, want /open-apis/task/v2/tasks/task-1/add_tasklist\nstdout:\n%s", url, out)
		}
		if got := gjson.Get(out, "api.0.params.user_id_type").String(); got != "open_id" {
			t.Fatalf("api[0].params.user_id_type = %q, want open_id\nstdout:\n%s", got, out)
		}
		if got := gjson.Get(out, "api.0.body.tasklist_guid").String(); got != "tl-123" {
			t.Fatalf("api[0].body.tasklist_guid = %q, want tl-123\nstdout:\n%s", got, out)
		}
		if got := gjson.Get(out, "api.0.body.section_guid").String(); got != "sec-456" {
			t.Fatalf("api[0].body.section_guid = %q, want sec-456\nstdout:\n%s", got, out)
		}
	})
}
