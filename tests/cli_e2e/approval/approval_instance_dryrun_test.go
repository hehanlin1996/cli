// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package approval

import (
	"context"
	"testing"
	"time"

	clie2e "github.com/larksuite/cli/tests/cli_e2e"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestApproval_InstancePreviewDryRun(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_APP_ID", "app")
	t.Setenv("LARKSUITE_CLI_APP_SECRET", "secret")
	t.Setenv("LARKSUITE_CLI_BRAND", "feishu")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	result, err := clie2e.RunCmd(ctx, clie2e.Request{
		Args: []string{
			"approval", "+instance-preview",
			"--data", `{"approval_code":"code","user_id":"ou_xxx","form":"[]"}`,
			"--user-id-type", "open_id",
			"--dry-run",
		},
		DefaultAs: "bot",
	})
	require.NoError(t, err)
	result.AssertExitCode(t, 0)

	require.Equal(t, "POST", gjson.Get(result.Stdout, "api.0.method").String(), "stdout:\n%s", result.Stdout)
	require.Equal(t, "/open-apis/approval/v4/instances/preview", gjson.Get(result.Stdout, "api.0.url").String(), "stdout:\n%s", result.Stdout)
	require.Equal(t, "open_id", gjson.Get(result.Stdout, "api.0.params.user_id_type").String(), "stdout:\n%s", result.Stdout)
	require.Equal(t, "code", gjson.Get(result.Stdout, "api.0.body.approval_code").String(), "stdout:\n%s", result.Stdout)
}

func TestApproval_InstanceCreateDryRun(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_APP_ID", "app")
	t.Setenv("LARKSUITE_CLI_APP_SECRET", "secret")
	t.Setenv("LARKSUITE_CLI_BRAND", "feishu")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	result, err := clie2e.RunCmd(ctx, clie2e.Request{
		Args: []string{
			"approval", "+instance-create",
			"--data", `{"approval_code":"code","user_id":"ou_xxx","form":"[]"}`,
			"--dry-run",
		},
		DefaultAs: "bot",
	})
	require.NoError(t, err)
	result.AssertExitCode(t, 0)

	require.Equal(t, "POST", gjson.Get(result.Stdout, "api.0.method").String(), "stdout:\n%s", result.Stdout)
	require.Equal(t, "/open-apis/approval/v4/instances", gjson.Get(result.Stdout, "api.0.url").String(), "stdout:\n%s", result.Stdout)
	require.False(t, gjson.Get(result.Stdout, "api.0.params.user_id_type").Exists(), "stdout:\n%s", result.Stdout)
	require.Equal(t, "code", gjson.Get(result.Stdout, "api.0.body.approval_code").String(), "stdout:\n%s", result.Stdout)
}
