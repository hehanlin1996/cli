// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/shortcuts/common"
)

func TestDocsUpdateV2MarkdownGuardrailWarnings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		docFormat   string
		command     string
		content     string
		wantCount   int
		wantNeedles []string
	}{
		{
			name:      "markdown append warns about fidelity",
			docFormat: "markdown",
			command:   "append",
			content:   "# Hello",
			wantCount: 1,
			wantNeedles: []string{
				"Markdown v2",
				"text-oriented",
				"style",
				"rich block",
				"XML",
			},
		},
		{
			name:      "markdown overwrite adds overwrite warning",
			docFormat: "markdown",
			command:   "overwrite",
			content:   "# Replacement",
			wantCount: 2,
			wantNeedles: []string{
				"overwrite",
				"entire document",
				"title",
				"images",
				"whiteboards",
				"attachments",
				"embedded blocks",
				"block_*",
				"fetch with ids",
			},
		},
		{
			name:      "markdown block delete without content does not warn",
			docFormat: "markdown",
			command:   "block_delete",
			content:   "",
			wantCount: 0,
		},
		{
			name:      "xml overwrite does not warn",
			docFormat: "xml",
			command:   "overwrite",
			content:   "<title>Replacement</title>",
			wantCount: 0,
		},
		{
			name:      "markdown str replace with replacement content warns",
			docFormat: "markdown",
			command:   "str_replace",
			content:   "new text",
			wantCount: 1,
		},
		{
			name:      "markdown str replace without content does not warn",
			docFormat: "markdown",
			command:   "str_replace",
			content:   "",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			warnings := docsUpdateV2MarkdownGuardrailWarnings(tt.command, tt.docFormat, tt.content)
			if len(warnings) != tt.wantCount {
				t.Fatalf("got %d warnings, want %d: %#v", len(warnings), tt.wantCount, warnings)
			}
			joined := strings.Join(warnings, "\n")
			for _, needle := range tt.wantNeedles {
				if !strings.Contains(joined, needle) {
					t.Errorf("warnings missing %q:\n%s", needle, joined)
				}
			}
		})
	}
}

func TestDryRunUpdateV2MarkdownIncludesGuardrailWarnings(t *testing.T) {
	t.Parallel()

	runtime := newUpdateV2GuardrailRuntime("markdown", "overwrite", "# Replacement")
	dryRun := dryRunUpdateV2(context.Background(), runtime)

	b, err := json.Marshal(dryRun)
	if err != nil {
		t.Fatalf("marshal dry run: %v", err)
	}
	var payload struct {
		Warnings []string `json:"warnings"`
	}
	if err := json.Unmarshal(b, &payload); err != nil {
		t.Fatalf("unmarshal dry run: %v\n%s", err, string(b))
	}
	if len(payload.Warnings) != 2 {
		t.Fatalf("dry-run warnings = %#v, want markdown + overwrite warnings\njson:%s", payload.Warnings, string(b))
	}

	joined := strings.Join(payload.Warnings, "\n")
	for _, needle := range []string{"Markdown v2", "XML", "overwrite", "block_*", "fetch with ids"} {
		if !strings.Contains(joined, needle) {
			t.Errorf("dry-run warnings missing %q:\n%s", needle, joined)
		}
	}
}

func TestEmitDocsUpdateV2MarkdownGuardrailWarningsToStderr(t *testing.T) {
	t.Parallel()

	runtime := newUpdateV2GuardrailRuntime("markdown", "append", "# Hello")
	var stderr bytes.Buffer
	runtime.Factory = &cmdutil.Factory{
		IOStreams: cmdutil.NewIOStreams(strings.NewReader(""), io.Discard, &stderr),
	}

	emitDocsUpdateV2MarkdownGuardrailWarnings(runtime)

	got := stderr.String()
	if !strings.Contains(got, "warning: ") || !strings.Contains(got, "Markdown v2") || !strings.Contains(got, "XML") {
		t.Fatalf("stderr missing markdown guardrail warning:\n%s", got)
	}
}

func newUpdateV2GuardrailRuntime(docFormat, command, content string) *common.RuntimeContext {
	cmd := &cobra.Command{Use: "+update"}
	cmd.Flags().String("doc", "doxcnGuardrailDryRun", "")
	cmd.Flags().String("doc-format", docFormat, "")
	cmd.Flags().String("command", command, "")
	cmd.Flags().Int("revision-id", -1, "")
	cmd.Flags().String("content", content, "")
	cmd.Flags().String("pattern", "old text", "")
	cmd.Flags().String("block-id", "blk123", "")
	cmd.Flags().String("src-block-ids", "", "")
	return common.TestNewRuntimeContextWithCtx(context.Background(), cmd, nil)
}
