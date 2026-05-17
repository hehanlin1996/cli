// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package output

import (
	"strings"
	"testing"
)

// TestClassifyLarkError_DriveCreateShortcutConstraints verifies known Drive shortcut errors map to actionable hints.
func TestClassifyLarkError_DriveCreateShortcutConstraints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		code         int
		wantExitCode int
		wantType     string
		wantHint     string
	}{
		{
			name:         "resource contention",
			code:         LarkErrDriveResourceContention,
			wantExitCode: ExitAPI,
			wantType:     "conflict",
			wantHint:     "avoid concurrent duplicate requests",
		},
		{
			name:         "cross tenant unit",
			code:         LarkErrDriveCrossTenantUnit,
			wantExitCode: ExitAPI,
			wantType:     "cross_tenant_unit",
			wantHint:     "same tenant and region/unit",
		},
		{
			name:         "cross brand",
			code:         LarkErrDriveCrossBrand,
			wantExitCode: ExitAPI,
			wantType:     "cross_brand",
			wantHint:     "same brand environment",
		},
		{
			name:         "sheets float image invalid dims",
			code:         LarkErrSheetsFloatImageInvalidDims,
			wantExitCode: ExitAPI,
			wantType:     "invalid_params",
			wantHint:     "--width / --height / --offset-x / --offset-y",
		},
		{
			name:         "drive permission apply rate limit",
			code:         LarkErrDrivePermApplyRateLimit,
			wantExitCode: ExitAPI,
			wantType:     "rate_limit",
			wantHint:     "5 times per day",
		},
		{
			name:         "drive permission apply not applicable",
			code:         LarkErrDrivePermApplyNotApplicable,
			wantExitCode: ExitAPI,
			wantType:     "invalid_params",
			wantHint:     "does not accept a permission-apply request",
		},
		{
			name:         "ownership mismatch",
			code:         LarkErrOwnershipMismatch,
			wantExitCode: ExitAPI,
			wantType:     "ownership_mismatch",
			wantHint:     "messages-resources-download",
		},
		{
			name:         "monthly api quota exceeded",
			code:         LarkErrMonthlyQuotaExceeded,
			wantExitCode: ExitAPI,
			wantType:     "quota_exceeded",
			wantHint:     "Do not retry repeatedly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotExitCode, gotType, gotHint := ClassifyLarkError(tt.code, "raw msg")
			if gotExitCode != tt.wantExitCode {
				t.Fatalf("exitCode=%d, want %d", gotExitCode, tt.wantExitCode)
			}
			if gotType != tt.wantType {
				t.Fatalf("type=%q, want %q", gotType, tt.wantType)
			}
			if gotHint == "" {
				t.Fatal("expected non-empty hint")
			}
			if !strings.Contains(gotHint, tt.wantHint) {
				t.Fatalf("hint=%q, want substring %q", gotHint, tt.wantHint)
			}
		})
	}
}

func TestClassifyLarkError_AppProfileMismatchSignals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code int
		msg  string
	}{
		{
			name: "generic invalid param",
			code: LarkErrInvalidParam,
			msg:  "TAT API error: [10003] invalid param",
		},
		{
			name: "open id cross app",
			code: 0,
			msg:  "open_id cross app",
		},
		{
			name: "message recall not sender",
			code: 0,
			msg:  "message recall failed: current bot is not the sender",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotExitCode, gotType, gotHint := ClassifyLarkError(tt.code, tt.msg)
			if gotExitCode != ExitAPI {
				t.Fatalf("exitCode=%d, want %d", gotExitCode, ExitAPI)
			}
			if gotType != "app_profile_mismatch" {
				t.Fatalf("type=%q, want %q", gotType, "app_profile_mismatch")
			}
			for _, want := range []string{"profile/app_id", "lark-cli profile list", "--profile", "lark-cli config bind", "lark-cli auth login"} {
				if !strings.Contains(gotHint, want) {
					t.Fatalf("hint=%q, want substring %q", gotHint, want)
				}
			}
			if strings.Contains(strings.ToLower(gotHint), "app_secret") {
				t.Fatalf("hint must not mention secrets: %q", gotHint)
			}
		})
	}
}

func TestErrAPI_MonthlyQuotaExceededHasNextStep(t *testing.T) {
	t.Parallel()

	err := ErrAPI(LarkErrMonthlyQuotaExceeded, "API error: [99991403] This month's API call quota has been exceeded", nil)
	if err.Code != ExitAPI {
		t.Fatalf("exit code = %d, want %d", err.Code, ExitAPI)
	}
	if err.Detail == nil {
		t.Fatal("expected detail")
	}
	if err.Detail.Type != "quota_exceeded" {
		t.Fatalf("type = %q, want %q", err.Detail.Type, "quota_exceeded")
	}
	for _, want := range []string{"monthly API call quota", "developer console", "request a quota increase", "contact an admin or official support", "Do not retry repeatedly"} {
		if !strings.Contains(err.Detail.Hint, want) {
			t.Fatalf("hint=%q, want substring %q", err.Detail.Hint, want)
		}
	}
}
