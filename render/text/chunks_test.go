package text

import (
	"slices"
	"strings"
	"testing"

	"github.com/go-text/typesetting/bidi"
)

func TestAppendCutsFollowTheChunkRule(t *testing.T) {
	cases := []struct {
		name string
		line string
		want []int
	}{
		{"empty", "", nil},
		{"no terminator", "abc def", []int{7}},
		{"sentence terminator cuts", "Hi! There", []int{3, 9}},
		{"full stop then capital cuts", "Hi. There", []int{3, 9}},
		{"full stop then lower case does not", "etc. and", []int{8}},
		{"mark after a full stop hides the space", "Hi.\u0301 There", []int{11}},
		{"full stop without space does not", "a.b", []int{3}},
		{"cluster stays whole", "What?! Yes", []int{6, 10}},
		{"marks stay with the terminator", "Hi!\u0301 There", []int{5, 11}},
		// Newer than Go's tables: Sentence_Break STerm since Unicode 16.
		{"terminators newer than the unicode tables", "a⹠b", []int{4, 5}},
		{"right to left run is one chunk", "a ש. ע b", []int{2, 8, 10}},
		{"fallback cuts after a space past 4 KiB", strings.Repeat("x", 4100) + " y z", []int{4101, 4104}},
		{"fallback needs the bytes first", strings.Repeat("x", 4000) + " y", []int{4002}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var paragraph bidi.Paragraph
			got := appendCuts(nil, testCase.line, &paragraph)
			if !slices.Equal(got, testCase.want) {
				t.Errorf("appendCuts(%.30q) = %v, want %v", testCase.line, got, testCase.want)
			}
		})
	}
}
