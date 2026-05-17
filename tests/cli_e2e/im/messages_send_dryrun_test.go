// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import (
	"context"
	"os"
	"testing"
	"time"

	clie2e "github.com/larksuite/cli/tests/cli_e2e"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const dryRunInteractiveCardContent = `{
  "config": {"wide_screen_mode": true},
  "elements": [{"tag": "div", "text": {"tag": "plain_text", "content": "hello from card"}}]
}`

func TestMessagesSendDryRun_ContentFile(t *testing.T) {
	setIMDryRunConfigEnv(t)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(dir+"/card.json", []byte(dryRunInteractiveCardContent), 0o644))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	result, err := clie2e.RunCmd(ctx, clie2e.Request{
		Args: []string{
			"im", "+messages-send",
			"--chat-id", "oc_dryrun",
			"--msg-type", "interactive",
			"--content", "@card.json",
			"--dry-run",
		},
		DefaultAs: "bot",
		WorkDir:   dir,
	})
	require.NoError(t, err)
	result.AssertExitCode(t, 0)

	assertMessagesSendDryRunContent(t, result.Stdout, dryRunInteractiveCardContent)
}

func TestMessagesSendDryRun_ContentStdin(t *testing.T) {
	setIMDryRunConfigEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	result, err := clie2e.RunCmd(ctx, clie2e.Request{
		Args: []string{
			"im", "+messages-send",
			"--chat-id", "oc_dryrun",
			"--msg-type", "interactive",
			"--content", "-",
			"--dry-run",
		},
		DefaultAs: "bot",
		Stdin:     []byte(dryRunInteractiveCardContent),
	})
	require.NoError(t, err)
	result.AssertExitCode(t, 0)

	assertMessagesSendDryRunContent(t, result.Stdout, dryRunInteractiveCardContent)
}

func assertMessagesSendDryRunContent(t *testing.T, stdout, wantContent string) {
	t.Helper()
	require.True(t, gjson.Valid(stdout), "dry-run stdout must be JSON: %s", stdout)
	require.Equal(t, "/open-apis/im/v1/messages", gjson.Get(stdout, "api.0.url").String())
	require.Equal(t, "POST", gjson.Get(stdout, "api.0.method").String())
	require.Equal(t, "chat_id", gjson.Get(stdout, "api.0.params.receive_id_type").String())
	require.Equal(t, "interactive", gjson.Get(stdout, "api.0.body.msg_type").String())
	require.Equal(t, "oc_dryrun", gjson.Get(stdout, "api.0.body.receive_id").String())
	require.Equal(t, wantContent, gjson.Get(stdout, "api.0.body.content").String())
}

func setIMDryRunConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	t.Setenv("LARKSUITE_CLI_APP_ID", "im_dryrun_test")
	t.Setenv("LARKSUITE_CLI_APP_SECRET", "im_dryrun_secret")
	t.Setenv("LARKSUITE_CLI_BRAND", "feishu")
}
