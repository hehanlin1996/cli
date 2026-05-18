// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT
package doc

import (
	"reflect"
	"testing"

	"github.com/larksuite/cli/shortcuts/common"
	"github.com/spf13/cobra"
)

// ── V2 tests ──

func TestValidCommandsV2(t *testing.T) {
	expected := map[string]bool{
		"str_replace":             true,
		"block_delete":            true,
		"block_insert_after":      true,
		"block_copy_insert_after": true,
		"block_replace":           true,
		"block_move_after":        true,
		"overwrite":               true,
		"append":                  true,
	}
	if len(validCommandsV2) != len(expected) {
		t.Fatalf("expected %d commands, got %d", len(expected), len(validCommandsV2))
	}
	for cmd := range validCommandsV2 {
		if !expected[cmd] {
			t.Fatalf("unexpected command %q in validCommandsV2", cmd)
		}
	}
}

// ── V1 tests ──

func TestIsWhiteboardCreateMarkdown(t *testing.T) {
	t.Run("blank whiteboard tags", func(t *testing.T) {
		markdown := "<whiteboard type=\"blank\"></whiteboard>\n<whiteboard type=\"blank\"></whiteboard>"
		if !isWhiteboardCreateMarkdown(markdown) {
			t.Fatalf("expected blank whiteboard markdown to be treated as whiteboard creation")
		}
	})

	t.Run("mermaid code block", func(t *testing.T) {
		markdown := "```mermaid\ngraph TD\nA-->B\n```"
		if !isWhiteboardCreateMarkdown(markdown) {
			t.Fatalf("expected mermaid markdown to be treated as whiteboard creation")
		}
	})

	t.Run("plain markdown", func(t *testing.T) {
		markdown := "## plain text"
		if isWhiteboardCreateMarkdown(markdown) {
			t.Fatalf("did not expect plain markdown to be treated as whiteboard creation")
		}
	})
}

func TestNormalizeWhiteboardResult(t *testing.T) {
	t.Run("adds empty board_tokens when whiteboard creation response omits it", func(t *testing.T) {
		result := map[string]interface{}{
			"success": true,
		}

		normalizeWhiteboardResult(result, "<whiteboard type=\"blank\"></whiteboard>")

		got, ok := result["board_tokens"].([]string)
		if !ok {
			t.Fatalf("expected board_tokens to be []string, got %T", result["board_tokens"])
		}
		if len(got) != 0 {
			t.Fatalf("expected empty board_tokens, got %#v", got)
		}
	})

	t.Run("normalizes board_tokens to string slice", func(t *testing.T) {
		result := map[string]interface{}{
			"board_tokens": []interface{}{"board_1", "board_2"},
		}

		normalizeWhiteboardResult(result, "<whiteboard type=\"blank\"></whiteboard>")

		want := []string{"board_1", "board_2"}
		got, ok := result["board_tokens"].([]string)
		if !ok {
			t.Fatalf("expected board_tokens to be []string, got %T", result["board_tokens"])
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("board_tokens mismatch: got %#v want %#v", got, want)
		}
	})

	t.Run("leaves non whiteboard response unchanged", func(t *testing.T) {
		result := map[string]interface{}{
			"success": true,
		}

		normalizeWhiteboardResult(result, "## plain text")

		if _, ok := result["board_tokens"]; ok {
			t.Fatalf("did not expect board_tokens for non-whiteboard markdown")
		}
	})
}

func TestBuildUpdateArgsV1NormalizesMarkdownEscapes(t *testing.T) {
	cmd := &cobra.Command{Use: "+update"}
	cmd.Flags().String("doc", "doc_token", "")
	cmd.Flags().String("mode", "append", "")
	cmd.Flags().String("markdown", `SAMPLE\_RATE and 1\+1`, "")
	runtime := common.TestNewRuntimeContext(cmd, nil)

	args := buildUpdateArgsV1(runtime)
	if got := args["markdown"]; got != "SAMPLE_RATE and 1+1" {
		t.Fatalf("markdown = %#v, want normalized markdown escapes", got)
	}
}

func TestBuildCreateArgsV1NormalizesMarkdownEscapes(t *testing.T) {
	cmd := &cobra.Command{Use: "+create"}
	cmd.Flags().String("markdown", `SAMPLE\_RATE and 1\+1`, "")
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("folder-token", "", "")
	cmd.Flags().String("wiki-node", "", "")
	cmd.Flags().String("wiki-space", "", "")
	runtime := common.TestNewRuntimeContext(cmd, nil)

	args := buildCreateArgsV1(runtime)
	if got := args["markdown"]; got != "SAMPLE_RATE and 1+1" {
		t.Fatalf("markdown = %#v, want normalized markdown escapes", got)
	}
}
