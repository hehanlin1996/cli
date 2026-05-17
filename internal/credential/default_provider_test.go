// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package credential

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/internal/vfs"
)

type permissionDeniedDefaultAccountFS struct{ vfs.FS }

func (p permissionDeniedDefaultAccountFS) ReadFile(name string) ([]byte, error) {
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrPermission}
}

func TestDefaultTokenProvider_Dispatches(t *testing.T) {
	// Just verify the type implements DefaultTokenResolver
	var _ DefaultTokenResolver = &DefaultTokenProvider{}
}

func TestDefaultAccountProvider_Implements(t *testing.T) {
	var _ DefaultAccountResolver = &DefaultAccountProvider{}
}

func TestDefaultAccountProvider_ConfigLoadPermissionDeniedIsNotNotConfigured(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	oldWorkspace := core.CurrentWorkspace()
	core.SetCurrentWorkspace(core.WorkspaceLocal)
	t.Cleanup(func() { core.SetCurrentWorkspace(oldWorkspace) })
	oldFS := vfs.DefaultFS
	vfs.DefaultFS = permissionDeniedDefaultAccountFS{FS: oldFS}
	t.Cleanup(func() { vfs.DefaultFS = oldFS })

	_, err := NewDefaultAccountProvider(nil, "").ResolveAccount(context.Background())
	if err == nil {
		t.Fatal("expected error for unreadable config")
	}
	var cfgErr *core.ConfigError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("error type = %T, want *core.ConfigError", err)
	}
	if cfgErr.Message == "not configured" || strings.Contains(cfgErr.Message, "not configured") {
		t.Fatalf("permission-denied config must preserve load cause, got %q", cfgErr.Message)
	}
	if !strings.Contains(cfgErr.Message, "failed to load config") || !strings.Contains(cfgErr.Message, "permission denied") {
		t.Fatalf("message = %q, want failed load with permission denied", cfgErr.Message)
	}
	if !strings.Contains(cfgErr.Hint, "chmod 600") || !strings.Contains(cfgErr.Hint, "lark-cli config show") {
		t.Fatalf("hint = %q, want permission-repair guidance", cfgErr.Hint)
	}
}
