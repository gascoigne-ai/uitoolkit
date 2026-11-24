package render_test

import (
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
)

func TestStrokeZeroValueIsButtMiterAllBevel(t *testing.T) {
	cases := []struct {
		name      string
		got, want int
	}{
		{"LineCapButt", int(render.LineCapButt), 0},
		{"LineCapRound", int(render.LineCapRound), 1},
		{"LineCapSquare", int(render.LineCapSquare), 2},
		{"LineJoinMiter", int(render.LineJoinMiter), 0},
		{"LineJoinBevel", int(render.LineJoinBevel), 1},
		{"LineJoinRound", int(render.LineJoinRound), 2},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.got != testCase.want {
				t.Errorf("%s = %d, want %d", testCase.name, testCase.got, testCase.want)
			}
		})
	}
	var stroke render.Stroke
	if stroke.Cap != render.LineCapButt || stroke.Join != render.LineJoinMiter || stroke.MiterLimit != 0 {
		t.Errorf("zero Stroke = %+v, want a butt cap, a miter join and no miter limit", stroke)
	}
}
