// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package task

import (
	"errors"
	"testing"

	"github.com/larksuite/cli/internal/httpmock"
	"github.com/larksuite/cli/internal/output"
)

func TestCreateTasklist_TaskCreateFailureReturnsError(t *testing.T) {
	f, stdout, _, reg := taskShortcutTestFactory(t)
	warmTenantToken(t, f, reg)

	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/task/v2/tasklists",
		Body: map[string]interface{}{
			"code": 0,
			"msg":  "success",
			"data": map[string]interface{}{
				"tasklist": map[string]interface{}{
					"guid": "tl-123",
					"name": "Daily triage",
					"url":  "https://applink.feishu.cn/client/todo/task_list?guid=tl-123",
				},
			},
		},
	})
	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/task/v2/tasks",
		Body: map[string]interface{}{
			"code": 1470404,
			"msg":  "The tasklist with guid 'tl-123' cannot be found or has been deleted.",
		},
	})

	s := CreateTasklist
	s.AuthTypes = []string{"bot", "user"}
	args := []string{
		"+tasklist-create",
		"--name", "Daily triage",
		"--data", `[{"summary":"child task"}]`,
		"--as", "bot",
		"--format", "json",
	}

	err := runMountedTaskShortcut(t, s, args, f, stdout)
	if err == nil {
		t.Fatalf("expected child task create failure to return an error")
	}

	var exitErr *output.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if exitErr.Detail == nil || exitErr.Detail.Type != "task_partial_failure" {
		t.Fatalf("unexpected error detail: %+v", exitErr.Detail)
	}
	if stdout.String() != "" {
		t.Fatalf("partial failure should not emit ok:true stdout, got: %s", stdout.String())
	}
}
