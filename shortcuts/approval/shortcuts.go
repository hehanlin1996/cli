// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package approval

import "github.com/larksuite/cli/shortcuts/common"

// Shortcuts returns all approval shortcuts.
func Shortcuts() []common.Shortcut {
	return []common.Shortcut{
		ApprovalInstancePreview,
		ApprovalInstanceCreate,
	}
}
