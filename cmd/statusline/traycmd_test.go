package main

import (
	"bytes"
	"slices"
	"testing"

	"statusline/internal/tray"
)

func TestUpdateWindowsTray(t *testing.T) {
	var app trayApp
	var tooltips []string
	var icons [][]byte
	titles := []string{"z87 w45", "z87 w44", "c87", "c48", "z87 w44", "…"}
	for _, title := range titles {
		app.updateWindowsTray(title, func(tooltip string) {
			tooltips = append(tooltips, tooltip)
		}, func(icon []byte) {
			icons = append(icons, icon)
		})
	}
	if !slices.Equal(tooltips, titles) {
		t.Errorf("tooltips = %q, want %q", tooltips, titles)
	}
	labels := []string{"87", "48", "87", "–"}
	if len(icons) != len(labels) {
		t.Fatalf("got %d icon updates, want %d", len(icons), len(labels))
	}
	for i, label := range labels {
		if !bytes.Equal(icons[i], tray.Icon(label)) {
			t.Errorf("icon update %d does not show %q", i, label)
		}
	}
}
