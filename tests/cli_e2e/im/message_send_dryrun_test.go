// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import (
	"context"
	"strings"
	"testing"
	"time"

	clie2e "github.com/larksuite/cli/tests/cli_e2e"
	"github.com/stretchr/testify/require"
)

func TestIMMessagesSendDryRun_NormalizesStyledMarkdownLinks(t *testing.T) {
	setIMDryRunConfigEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	result, err := clie2e.RunCmd(ctx, clie2e.Request{
		Args: []string{
			"im", "+messages-send",
			"--chat-id", "oc_dryrun_chat",
			"--markdown", "**[click here](https://example.com/path_with_underscore)**",
			"--dry-run",
		},
		DefaultAs: "bot",
	})
	require.NoError(t, err)
	result.AssertExitCode(t, 0)

	if got := result.Stdout; !strings.Contains(got, `[click here](https://example.com/path_with_underscore)`) ||
		strings.Contains(got, `**[click here](https://example.com/path_with_underscore)**`) {
		t.Fatalf("dry-run markdown was not normalized:\n%s", got)
	}
}

func setIMDryRunConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	t.Setenv("LARKSUITE_CLI_APP_ID", "im_dryrun_test")
	t.Setenv("LARKSUITE_CLI_APP_SECRET", "im_dryrun_secret")
	t.Setenv("LARKSUITE_CLI_BRAND", "feishu")
}
