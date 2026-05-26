// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package vc

import (
	"context"
	"strings"
	"testing"
	"time"

	clie2e "github.com/larksuite/cli/tests/cli_e2e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVCNotesMinuteTokens_DryRun(t *testing.T) {
	setDryRunConfigEnv(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	result, err := clie2e.RunCmd(ctx, clie2e.Request{
		Args: []string{
			"vc", "+notes",
			"--minute-tokens", "obcnexampleminute",
			"--dry-run",
		},
		DefaultAs: "user",
	})
	require.NoError(t, err)
	result.AssertExitCode(t, 0)

	output := result.Stdout
	assert.True(t, strings.Contains(output, "GET"), "dry-run should contain GET method, got: %s", output)
	assert.True(t, strings.Contains(output, "/open-apis/minutes/v1/minutes/{minute_token}"), "dry-run should contain minute detail path, got: %s", output)
	assert.True(t, strings.Contains(output, "/open-apis/minutes/v1/minutes/{minute_token}/artifacts"), "dry-run should contain artifacts path, got: %s", output)
	assert.True(t, strings.Contains(output, "/open-apis/minutes/v1/minutes/{minute_token}/transcript"), "dry-run should contain transcript path, got: %s", output)
}

func setDryRunConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("LARKSUITE_CLI_APP_ID", "cli_dryrun_test")
	t.Setenv("LARKSUITE_CLI_APP_SECRET", "dryrun_secret")
	t.Setenv("LARKSUITE_CLI_BRAND", "feishu")
}
