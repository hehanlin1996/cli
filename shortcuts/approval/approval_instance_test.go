// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package approval

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/larksuite/cli/shortcuts/common"
	"github.com/spf13/cobra"
)

func TestBuildApprovalInstanceBodyRequiresJSONObject(t *testing.T) {
	runtime := newApprovalInstanceRuntime(`[]`, "")
	if _, err := buildApprovalInstanceBody(runtime); err == nil {
		t.Fatalf("expected non-object JSON to fail")
	}
}

func TestApprovalInstancePreviewDryRun(t *testing.T) {
	runtime := newApprovalInstanceRuntime(`{"approval_code":"code","user_id":"ou_xxx","form":"[]"}`, "open_id")

	dry := ApprovalInstancePreview.DryRun(context.Background(), runtime)
	calls := dryRunCalls(t, dry)
	if got := calls[0].Method; got != "POST" {
		t.Fatalf("method = %q, want POST", got)
	}
	if got := calls[0].URL; got != "/open-apis/approval/v4/instances/preview" {
		t.Fatalf("url = %q", got)
	}
	if got := calls[0].Params["user_id_type"]; got != "open_id" {
		t.Fatalf("user_id_type = %#v", got)
	}
	body := calls[0].Body
	if got := body["approval_code"]; got != "code" {
		t.Fatalf("approval_code = %#v", got)
	}
}

func TestApprovalInstanceCreateDryRun(t *testing.T) {
	runtime := newApprovalInstanceRuntime(`{"approval_code":"code","user_id":"ou_xxx","form":"[]"}`, "")

	dry := ApprovalInstanceCreate.DryRun(context.Background(), runtime)
	calls := dryRunCalls(t, dry)
	if got := calls[0].Method; got != "POST" {
		t.Fatalf("method = %q, want POST", got)
	}
	if got := calls[0].URL; got != "/open-apis/approval/v4/instances" {
		t.Fatalf("url = %q", got)
	}
	if _, ok := calls[0].Params["user_id_type"]; ok {
		t.Fatalf("did not expect empty user_id_type in params: %#v", calls[0].Params)
	}
}

type dryRunEnvelope struct {
	API []dryRunCall `json:"api"`
}

type dryRunCall struct {
	Method string         `json:"method"`
	URL    string         `json:"url"`
	Params map[string]any `json:"params"`
	Body   map[string]any `json:"body"`
}

func dryRunCalls(t *testing.T, dry *common.DryRunAPI) []dryRunCall {
	t.Helper()
	data, err := json.Marshal(dry)
	if err != nil {
		t.Fatal(err)
	}
	var envelope dryRunEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.API) == 0 {
		t.Fatalf("dry-run has no API calls: %s", data)
	}
	return envelope.API
}

func newApprovalInstanceRuntime(data, userIDType string) *common.RuntimeContext {
	cmd := &cobra.Command{Use: "+instance-preview"}
	cmd.Flags().String("data", data, "")
	cmd.Flags().String("user-id-type", userIDType, "")
	return common.TestNewRuntimeContext(cmd, nil)
}
