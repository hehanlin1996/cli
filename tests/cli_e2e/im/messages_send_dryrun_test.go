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
	"github.com/tidwall/gjson"
)

func setIMDryRunEnv(t *testing.T) {
	t.Helper()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	t.Setenv("LARKSUITE_CLI_APP_ID", "app")
	t.Setenv("LARKSUITE_CLI_APP_SECRET", "secret")
	t.Setenv("LARKSUITE_CLI_BRAND", "feishu")
}

func TestIM_MessagesSendDryRunWarnsChatMembershipNotVerified(t *testing.T) {
	setIMDryRunEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	result, err := clie2e.RunCmd(ctx, clie2e.Request{
		Args: []string{
			"im", "+messages-send",
			"--chat-id", "oc_dryrun_chat",
			"--text", "hello",
			"--dry-run",
		},
		DefaultAs: "bot",
	})
	require.NoError(t, err)
	result.AssertExitCode(t, 0)

	out := result.Stdout
	if got := gjson.Get(out, "api.0.method").String(); got != "POST" {
		t.Fatalf("method = %q, want POST\nstdout:\n%s", got, out)
	}
	if got := gjson.Get(out, "api.0.url").String(); got != "/open-apis/im/v1/messages" {
		t.Fatalf("url = %q, want messages endpoint\nstdout:\n%s", got, out)
	}
	if got := gjson.Get(out, "api.0.params.receive_id_type").String(); got != "chat_id" {
		t.Fatalf("receive_id_type = %q, want chat_id\nstdout:\n%s", got, out)
	}

	message := gjson.Get(out, "_notice.im_chat_membership.message").String()
	description := gjson.Get(out, "description").String()
	for _, want := range []string{
		"Dry-run validates request shape only",
		"does not verify that the selected bot/user is a member of the target chat",
		"Bot/User can NOT be out of the chat",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("membership notice missing %q\nstdout:\n%s", want, out)
		}
		if !strings.Contains(description, want) {
			t.Fatalf("dry-run description missing %q\nstdout:\n%s", want, out)
		}
	}
}

func TestIM_MessagesSendDryRunPretty(t *testing.T) {
	setIMDryRunEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	result, err := clie2e.RunCmd(ctx, clie2e.Request{
		Args: []string{
			"im", "+messages-send",
			"--chat-id", "oc_dryrun_chat",
			"--text", "hello",
			"--dry-run",
		},
		DefaultAs: "bot",
		Format:    "pretty",
	})
	require.NoError(t, err)
	result.AssertExitCode(t, 0)

	for _, want := range []string{
		"# Dry-run validates request shape only",
		"does not verify that the selected bot/user is a member of the target chat",
		"POST /open-apis/im/v1/messages?receive_id_type=chat_id",
		`"receive_id":"oc_dryrun_chat"`,
		`"msg_type":"text"`,
	} {
		if !strings.Contains(result.Stdout, want) {
			t.Fatalf("pretty dry-run output missing %q\nstdout:\n%s\nstderr:\n%s", want, result.Stdout, result.Stderr)
		}
	}
	if !strings.Contains(result.Stderr, "=== Dry Run ===") {
		t.Fatalf("pretty dry-run stderr missing dry-run banner\nstdout:\n%s\nstderr:\n%s", result.Stdout, result.Stderr)
	}
}

func TestIM_MessagesHelpDocumentsMediaInputs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	for _, command := range []string{"+messages-send", "+messages-reply"} {
		t.Run(command, func(t *testing.T) {
			result, err := clie2e.RunCmd(ctx, clie2e.Request{
				Args: []string{"im", command, "--help"},
			})
			require.NoError(t, err)
			result.AssertExitCode(t, 0)

			for _, want := range []string{
				"image_key",
				"file_key",
				"URL",
				"cwd-relative local file path",
				"absolute paths are rejected",
			} {
				if !strings.Contains(result.Stdout, want) {
					t.Fatalf("%s help missing %q\nstdout:\n%s\nstderr:\n%s", command, want, result.Stdout, result.Stderr)
				}
			}
		})
	}
}
