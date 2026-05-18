// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"context"
	"testing"

	"github.com/larksuite/cli/shortcuts/common"
	"github.com/spf13/cobra"
)

func TestBuildFetchBodyIncludesSceneFromContext(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), docsSceneContextKey, " DoubaoCLI ")
	runtime := newFetchBodyTestRuntime(ctx)

	body := buildFetchBody(runtime)
	if got := body["scene"]; got != "DoubaoCLI" {
		t.Fatalf("scene = %#v, want %q", got, "DoubaoCLI")
	}
}

func TestBuildCreateBodyIncludesSceneFromContext(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), docsSceneContextKey, "DoubaoCLI")
	runtime := newCreateBodyTestRuntime(ctx)

	body := buildCreateBody(runtime)
	if got := body["scene"]; got != "DoubaoCLI" {
		t.Fatalf("scene = %#v, want %q", got, "DoubaoCLI")
	}
}

func TestBuildUpdateBodyIncludesSceneFromContext(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), docsSceneContextKey, "DoubaoCLI")
	runtime := newUpdateBodyTestRuntime(ctx)

	body := buildUpdateBody(runtime)
	if got := body["scene"]; got != "DoubaoCLI" {
		t.Fatalf("scene = %#v, want %q", got, "DoubaoCLI")
	}
}

func TestBuildCreateBodyNormalizesMarkdownEscapesOnlyForMarkdown(t *testing.T) {
	runtime := newCreateBodyTestRuntime(context.Background())
	if err := runtime.Cmd.Flags().Set("doc-format", "markdown"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Cmd.Flags().Set("content", `SAMPLE\_RATE and 1\+1`); err != nil {
		t.Fatal(err)
	}

	body := buildCreateBody(runtime)
	if got := body["content"]; got != "SAMPLE_RATE and 1+1" {
		t.Fatalf("content = %#v, want normalized markdown escapes", got)
	}

	xmlRuntime := newCreateBodyTestRuntime(context.Background())
	if err := xmlRuntime.Cmd.Flags().Set("content", `<p>SAMPLE\_RATE</p>`); err != nil {
		t.Fatal(err)
	}
	xmlBody := buildCreateBody(xmlRuntime)
	if got := xmlBody["content"]; got != `<p>SAMPLE\_RATE</p>` {
		t.Fatalf("xml content = %#v, want unchanged XML content", got)
	}
}

func TestBuildUpdateBodyNormalizesMarkdownEscapesOnlyForMarkdown(t *testing.T) {
	runtime := newUpdateBodyTestRuntime(context.Background())
	if err := runtime.Cmd.Flags().Set("doc-format", "markdown"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Cmd.Flags().Set("content", `SAMPLE\_RATE and 1\+1`); err != nil {
		t.Fatal(err)
	}

	body := buildUpdateBody(runtime)
	if got := body["content"]; got != "SAMPLE_RATE and 1+1" {
		t.Fatalf("content = %#v, want normalized markdown escapes", got)
	}

	xmlRuntime := newUpdateBodyTestRuntime(context.Background())
	if err := xmlRuntime.Cmd.Flags().Set("content", `<p>SAMPLE\_RATE</p>`); err != nil {
		t.Fatal(err)
	}
	xmlBody := buildUpdateBody(xmlRuntime)
	if got := xmlBody["content"]; got != `<p>SAMPLE\_RATE</p>` {
		t.Fatalf("xml content = %#v, want unchanged XML content", got)
	}
}

func TestBuildFetchBodyOmitsEmptyScene(t *testing.T) {
	t.Parallel()

	runtime := newFetchBodyTestRuntime(context.Background())

	body := buildFetchBody(runtime)
	if _, ok := body["scene"]; ok {
		t.Fatalf("did not expect empty scene in fetch body: %#v", body)
	}
}

func newFetchBodyTestRuntime(ctx context.Context) *common.RuntimeContext {
	cmd := &cobra.Command{Use: "+fetch"}
	cmd.Flags().String("doc-format", "xml", "")
	cmd.Flags().String("detail", "simple", "")
	cmd.Flags().Int("revision-id", -1, "")
	cmd.Flags().String("scope", "full", "")
	cmd.Flags().String("start-block-id", "", "")
	cmd.Flags().String("end-block-id", "", "")
	cmd.Flags().String("keyword", "", "")
	cmd.Flags().Int("context-before", 0, "")
	cmd.Flags().Int("context-after", 0, "")
	cmd.Flags().Int("max-depth", -1, "")
	return common.TestNewRuntimeContextWithCtx(ctx, cmd, nil)
}

func newCreateBodyTestRuntime(ctx context.Context) *common.RuntimeContext {
	cmd := &cobra.Command{Use: "+create"}
	cmd.Flags().String("doc-format", "xml", "")
	cmd.Flags().String("content", "<title>hello</title>", "")
	cmd.Flags().String("parent-token", "", "")
	cmd.Flags().String("parent-position", "", "")
	return common.TestNewRuntimeContextWithCtx(ctx, cmd, nil)
}

func newUpdateBodyTestRuntime(ctx context.Context) *common.RuntimeContext {
	cmd := &cobra.Command{Use: "+update"}
	cmd.Flags().String("doc-format", "xml", "")
	cmd.Flags().String("command", "append", "")
	cmd.Flags().Int("revision-id", 0, "")
	cmd.Flags().String("content", "<p>hello</p>", "")
	cmd.Flags().String("pattern", "", "")
	cmd.Flags().String("block-id", "", "")
	cmd.Flags().String("src-block-ids", "", "")
	return common.TestNewRuntimeContextWithCtx(ctx, cmd, nil)
}
