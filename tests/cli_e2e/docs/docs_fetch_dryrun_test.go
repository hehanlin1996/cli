// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package docs

import (
	"context"
	"testing"
	"time"

	clie2e "github.com/larksuite/cli/tests/cli_e2e"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestDocsFetchDryRun_TextFormatUsesRawContent(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	t.Setenv("LARKSUITE_CLI_APP_ID", "docs_fetch_dryrun_test")
	t.Setenv("LARKSUITE_CLI_APP_SECRET", "docs_fetch_dryrun_secret")
	t.Setenv("LARKSUITE_CLI_BRAND", "feishu")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	result, err := clie2e.RunCmd(ctx, clie2e.Request{
		Args: []string{
			"docs", "+fetch",
			"--api-version", "v2",
			"--doc", "docxTextDryRun",
			"--doc-format", "text",
			"--dry-run",
		},
		DefaultAs: "bot",
	})
	require.NoError(t, err)
	result.AssertExitCode(t, 0)

	out := result.Stdout
	if got := gjson.Get(out, "api.0.url").String(); got != "/open-apis/docs_ai/v1/documents/docxTextDryRun/fetch" {
		t.Fatalf("url=%q, want docs_ai fetch\nstdout:\n%s", got, out)
	}
	if got := gjson.Get(out, "api.0.body.format").String(); got != "raw_content" {
		t.Fatalf("body.format=%q, want raw_content\nstdout:\n%s", got, out)
	}
}
