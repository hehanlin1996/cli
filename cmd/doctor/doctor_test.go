// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"encoding/json"
	iofs "io/fs"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/internal/vfs"
)

type permissionDeniedDoctorFS struct{ vfs.FS }

func (p permissionDeniedDoctorFS) ReadFile(name string) ([]byte, error) {
	return nil, &iofs.PathError{Op: "open", Path: name, Err: iofs.ErrPermission}
}

func TestNewCmdDoctor_FlagParsing(t *testing.T) {
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{
		AppID: "test-app", AppSecret: "test-secret", Brand: core.BrandFeishu,
	})

	cmd := NewCmdDoctor(f)
	cmd.SetArgs([]string{"--offline"})

	// We only test flag parsing; skip actual execution by intercepting RunE.
	var gotOffline bool
	origRunE := cmd.RunE
	cmd.RunE = func(cmd2 *cobra.Command, args []string) error {
		v, _ := cmd2.Flags().GetBool("offline")
		gotOffline = v
		return nil
	}
	_ = origRunE

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotOffline {
		t.Error("expected --offline to be true")
	}
}

func TestFinishDoctor(t *testing.T) {
	t.Run("all pass returns nil", func(t *testing.T) {
		f, stdout, _, _ := cmdutil.TestFactory(t, nil)
		checks := []checkResult{
			pass("check1", "ok"),
			skip("check2", "skipped"),
		}
		err := finishDoctor(f, checks)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		var result struct {
			OK bool `json:"ok"`
		}
		json.Unmarshal(stdout.Bytes(), &result)
		if !result.OK {
			t.Error("expected ok=true")
		}
	})

	t.Run("any fail returns error", func(t *testing.T) {
		f, stdout, _, _ := cmdutil.TestFactory(t, nil)
		checks := []checkResult{
			pass("check1", "ok"),
			fail("check2", "bad", "fix it"),
		}
		err := finishDoctor(f, checks)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var result struct {
			OK bool `json:"ok"`
		}
		json.Unmarshal(stdout.Bytes(), &result)
		if result.OK {
			t.Error("expected ok=false")
		}
	})
}

func TestNetworkChecks_Offline(t *testing.T) {
	ep := core.Endpoints{Open: "https://open.feishu.cn", MCP: "https://mcp.feishu.cn"}
	opts := &DoctorOptions{Ctx: context.Background(), Offline: true}
	checks := networkChecks(opts.Ctx, opts, ep)
	if len(checks) != 2 {
		t.Fatalf("expected 2 checks, got %d", len(checks))
	}
	for _, c := range checks {
		if c.Status != "skip" {
			t.Errorf("expected skip, got %s for %s", c.Status, c.Name)
		}
	}
}

func TestDoctorRun_ConfigPermissionDeniedIncludesRepairHint(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	oldFS := vfs.DefaultFS
	vfs.DefaultFS = permissionDeniedDoctorFS{FS: oldFS}
	t.Cleanup(func() { vfs.DefaultFS = oldFS })

	f, stdout, _, _ := cmdutil.TestFactory(t, nil)
	err := doctorRun(&DoctorOptions{Factory: f, Ctx: context.Background(), Offline: true})
	if err == nil {
		t.Fatal("expected doctor to fail for unreadable config")
	}

	var result struct {
		OK     bool          `json:"ok"`
		Checks []checkResult `json:"checks"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("doctor output is not JSON: %v\n%s", err, stdout.String())
	}
	if result.OK {
		t.Fatalf("expected ok=false")
	}
	if len(result.Checks) < 2 || result.Checks[1].Name != "config_file" {
		t.Fatalf("expected config_file check after cli_version, got %+v", result.Checks)
	}
	check := result.Checks[1]
	if !strings.Contains(check.Message, "failed to load config") || !strings.Contains(check.Message, "permission denied") {
		t.Fatalf("message = %q, want failed load with permission denied", check.Message)
	}
	if !strings.Contains(check.Hint, "chmod 600") || !strings.Contains(check.Hint, "lark-cli config show") {
		t.Fatalf("hint = %q, want permission-repair guidance", check.Hint)
	}
}
