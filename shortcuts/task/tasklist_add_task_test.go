// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package task

import (
	"errors"
	"strings"
	"testing"

	"github.com/larksuite/cli/internal/httpmock"
	"github.com/larksuite/cli/internal/output"
)

func TestAddTaskToTasklist_Success(t *testing.T) {
	f, stdout, _, reg := taskShortcutTestFactory(t)
	warmTenantToken(t, f, reg)

	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/task/v2/tasks/task-1/add_tasklist",
		Body: map[string]interface{}{
			"code": 0, "msg": "success",
			"data": map[string]interface{}{
				"task": map[string]interface{}{
					"guid": "task-1",
				},
			},
		},
	})

	s := AddTaskToTasklist
	s.AuthTypes = []string{"bot", "user"}

	args := []string{"+tasklist-task-add", "--tasklist-id", "tl-123", "--task-id", "task-1", "--section-guid", "sec-456", "--as", "bot", "--format", "json"}
	err := runMountedTaskShortcut(t, s, args, f, stdout)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"tasklist_guid":"tl-123"`) && !strings.Contains(out, `"tasklist_guid": "tl-123"`) {
		t.Errorf("expected tasklist_guid in output, got: %s", out)
	}
}

func TestAddTaskToTasklist_FailedTasksReturnsError(t *testing.T) {
	f, stdout, _, reg := taskShortcutTestFactory(t)
	warmTenantToken(t, f, reg)

	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/task/v2/tasks/task-missing/add_tasklist",
		Body: map[string]interface{}{
			"code": 1470404,
			"msg":  "The task with guid 'task-missing' cannot be found or has been deleted.",
		},
	})

	s := AddTaskToTasklist
	s.AuthTypes = []string{"bot", "user"}

	args := []string{"+tasklist-task-add", "--tasklist-id", "tl-123", "--task-id", "task-missing", "--as", "bot", "--format", "json"}
	err := runMountedTaskShortcut(t, s, args, f, stdout)
	if err == nil {
		t.Fatalf("expected failed task add to return an error")
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
