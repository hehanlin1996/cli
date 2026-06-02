// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"testing"

	"github.com/larksuite/cli/internal/i18n"
)

func TestGetInitMsg_Zh(t *testing.T) {
	msg := getInitMsg("zh")
	if msg != initMsgZh {
		t.Error("expected zh message set")
	}
	if msg.SelectAction != "选择操作" {
		t.Errorf("unexpected SelectAction: %s", msg.SelectAction)
	}
}

func TestGetInitMsg_En(t *testing.T) {
	msg := getInitMsg("en")
	if msg != initMsgEn {
		t.Error("expected en message set")
	}
	if msg.SelectAction != "Select action" {
		t.Errorf("unexpected SelectAction: %s", msg.SelectAction)
	}
}

func TestGetInitMsg_DefaultsToZh(t *testing.T) {
	for _, lang := range []i18n.Lang{"", "unknown", "xyz", "invalid"} {
		msg := getInitMsg(lang)
		if msg != initMsgZh {
			t.Errorf("getInitMsg(%q) should default to zh", lang)
		}
	}
}

func TestInitMsgZh_AllFieldsNonEmpty(t *testing.T) {
	assertAllFieldsNonEmpty(t, initMsgZh, "zh")
}

func TestInitMsgEn_AllFieldsNonEmpty(t *testing.T) {
	assertAllFieldsNonEmpty(t, initMsgEn, "en")
}

func assertAllFieldsNonEmpty(t *testing.T, msg *initMsg, label string) {
	t.Helper()
	fields := map[string]string{
		"SelectAction":         msg.SelectAction,
		"CreateNewApp":         msg.CreateNewApp,
		"ConfigExistingApp":    msg.ConfigExistingApp,
		"Platform":             msg.Platform,
		"SelectPlatform":       msg.SelectPlatform,
		"Feishu":               msg.Feishu,
		"ScanQRCode":           msg.ScanQRCode,
		"ScanOrOpenLink":       msg.ScanOrOpenLink,
		"WaitingForScan":       msg.WaitingForScan,
		"OpenLinkNonTTY":       msg.OpenLinkNonTTY,
		"WaitingForScanNonTTY": msg.WaitingForScanNonTTY,
		"DetectedLarkTenant":   msg.DetectedLarkTenant,
		"AppCreated":           msg.AppCreated,
		"ConfigSaved":          msg.ConfigSaved,
		"LangPreferenceSet":    msg.LangPreferenceSet,
	}
	for name, val := range fields {
		if val == "" {
			t.Errorf("%s.%s is empty", label, name)
		}
	}
}

func TestInitMsg_FormatStrings(t *testing.T) {
	for _, lang := range []i18n.Lang{i18n.LangZhCN, i18n.LangEnUS} {
		msg := getInitMsg(lang)
		// AppCreated and ConfigSaved should contain %s for App ID
		got := fmt.Sprintf(msg.AppCreated, "cli_test123")
		if got == msg.AppCreated {
			t.Errorf("%s AppCreated has no format verb", lang)
		}
		got = fmt.Sprintf(msg.ConfigSaved, "cli_test123")
		if got == msg.ConfigSaved {
			t.Errorf("%s ConfigSaved has no format verb", lang)
		}
	}
}

func TestInitMsgZh_ConfigExistingAppMentionsExisting(t *testing.T) {
	msg := getInitMsg(i18n.LangZhCN)
	// The "use existing app" option must clearly convey "已有" (existing) so
	// users can discover it without guidance — regression test for #68.
	if !containsChinese(msg.ConfigExistingApp, "已有") {
		t.Errorf("zh ConfigExistingApp should contain '已有', got %q", msg.ConfigExistingApp)
	}
}

func TestInitMsgZh_CreateNewAppDoesNotPushRecommend(t *testing.T) {
	msg := getInitMsg(i18n.LangZhCN)
	// "推荐" label on the first option biases users away from the "existing"
	// option — regression test for #68.
	if containsChinese(msg.CreateNewApp, "推荐") {
		t.Errorf("zh CreateNewApp should not contain '推荐', got %q", msg.CreateNewApp)
	}
}

func TestInitMsgEn_ConfigExistingAppMentionsExisting(t *testing.T) {
	msg := getInitMsg(i18n.LangEnUS)
	if !containsWord(msg.ConfigExistingApp, "existing") {
		t.Errorf("en ConfigExistingApp should contain 'existing', got %q", msg.ConfigExistingApp)
	}
}

func TestInitMsgEn_CreateNewAppDoesNotPushRecommended(t *testing.T) {
	msg := getInitMsg(i18n.LangEnUS)
	if containsWord(msg.CreateNewApp, "Recommended") {
		t.Errorf("en CreateNewApp should not contain 'Recommended', got %q", msg.CreateNewApp)
	}
}

func containsChinese(s, sub string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			// Has CJK — do substring check
		}
	}
	return len(sub) > 0 && containsSubstring(s, sub)
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || (len(s) > 0 && findSubstring(s, sub)))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func containsWord(s, word string) bool {
	return findSubstring(s, word)
}

func TestGetInitMsg_BilingualCollapse(t *testing.T) {
	// The TUI is bilingual (zh + en). Only English-bucket languages return the
	// English struct — by canonical locale ("en_us") or legacy short ("en").
	// Everything else (zh, the other codes, invalid, "") returns Chinese.
	tests := []struct {
		lang       i18n.Lang
		shouldBeEn bool
	}{
		{i18n.LangZhCN, false},
		{i18n.LangEnUS, true},
		{"en", true}, // legacy short value
		{i18n.LangJaJP, false},
		{"fr_fr", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.lang), func(t *testing.T) {
			msg := getInitMsg(tt.lang)
			if msg == nil {
				t.Fatal("getInitMsg returned nil")
			}
			want := initMsgZh
			if tt.shouldBeEn {
				want = initMsgEn
			}
			if msg != want {
				t.Errorf("getInitMsg(%q) returned wrong struct", tt.lang)
			}
		})
	}
}
