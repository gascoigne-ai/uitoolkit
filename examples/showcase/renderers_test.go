package main

import (
	"slices"
	"strings"
	"testing"
)

func TestRendererTable(t *testing.T) {
	tests := []struct {
		name, renderer, wantError string
	}{
		{"ebiten opens", "ebiten", ""},
		{"wgpu is not built", "wgpu", `renderer "wgpu" not built: build with -tags wgpu (task 016)`},
		{"unknown name lists the names in order", "metal", `unknown renderer "metal": the renderers are ebiten, wgpu`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opened, err := openRenderer(test.renderer)
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("error = %v, want %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if opened.Device == nil || opened.ReadInput == nil || opened.Run == nil {
				t.Fatalf("backend is incomplete: %+v", opened)
			}
		})
	}
	names := make([]string, len(renderers))
	for i, row := range renderers {
		names[i] = row.Name
	}
	if unique := slices.Compact(slices.Sorted(slices.Values(names))); len(unique) != len(names) || len(names) == 0 {
		t.Errorf("renderer names = %v, want unique and not empty", names)
	}
}
