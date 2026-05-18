// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package approval

import (
	"context"
	"net/http"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"

	"github.com/larksuite/cli/shortcuts/common"
)

const (
	approvalInstancePreviewPath = "/open-apis/approval/v4/instances/preview"
	approvalInstanceCreatePath  = "/open-apis/approval/v4/instances"
)

var approvalInstanceFlags = []common.Flag{
	{Name: "data", Desc: "approval instance request body JSON object; supports @file and - stdin", Required: true, Input: []string{common.File, common.Stdin}},
	{Name: "user-id-type", Desc: "user ID type used in request fields", Enum: []string{"open_id", "user_id", "union_id"}},
}

var ApprovalInstancePreview = common.Shortcut{
	Service:     "approval",
	Command:     "+instance-preview",
	Description: "Preview an approval instance flow before submission",
	Risk:        "read",
	Scopes:      []string{"approval:instance:write"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags:       approvalInstanceFlags,
	Validate:    validateApprovalInstance,
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return dryRunApprovalInstance(runtime, approvalInstancePreviewPath)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		return executeApprovalInstance(runtime, approvalInstancePreviewPath)
	},
}

var ApprovalInstanceCreate = common.Shortcut{
	Service:     "approval",
	Command:     "+instance-create",
	Description: "Create an approval instance",
	Risk:        "write",
	Scopes:      []string{"approval:instance:write"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags:       approvalInstanceFlags,
	Validate:    validateApprovalInstance,
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return dryRunApprovalInstance(runtime, approvalInstanceCreatePath)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		return executeApprovalInstance(runtime, approvalInstanceCreatePath)
	},
}

func validateApprovalInstance(_ context.Context, runtime *common.RuntimeContext) error {
	_, err := buildApprovalInstanceBody(runtime)
	return err
}

func buildApprovalInstanceBody(runtime *common.RuntimeContext) (map[string]any, error) {
	var body map[string]any
	if err := common.ParseJSON([]byte(runtime.Str("data")), &body); err != nil {
		return nil, common.FlagErrorf("--data must be a valid JSON object: %v", err)
	}
	return body, nil
}

func dryRunApprovalInstance(runtime *common.RuntimeContext, path string) *common.DryRunAPI {
	body, _ := buildApprovalInstanceBody(runtime)
	return common.NewDryRunAPI().
		POST(path).
		Params(buildApprovalInstanceDryRunParams(runtime)).
		Body(body)
}

func executeApprovalInstance(runtime *common.RuntimeContext, path string) error {
	body, err := buildApprovalInstanceBody(runtime)
	if err != nil {
		return err
	}
	data, err := runtime.DoAPIJSON(http.MethodPost, path, buildApprovalInstanceQueryParams(runtime), body)
	if err != nil {
		return err
	}
	runtime.Out(data, nil)
	return nil
}

func buildApprovalInstanceDryRunParams(runtime *common.RuntimeContext) map[string]any {
	if userIDType := strings.TrimSpace(runtime.Str("user-id-type")); userIDType != "" {
		return map[string]any{"user_id_type": userIDType}
	}
	return nil
}

func buildApprovalInstanceQueryParams(runtime *common.RuntimeContext) larkcore.QueryParams {
	if userIDType := strings.TrimSpace(runtime.Str("user-id-type")); userIDType != "" {
		return larkcore.QueryParams{"user_id_type": []string{userIDType}}
	}
	return nil
}
