package widget

import (
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
)

// An empty stroke-linecap resets the cap to butt; an empty stroke-linejoin
// leaves the inherited join alone.
func TestEmptyLineCapResetsToButtAndEmptyLineJoinInherits(t *testing.T) {
	src := `<svg viewBox="0 0 4 4" stroke-linecap="round" stroke-linejoin="bevel"><path d="M0 0L1 1" stroke-linecap="" stroke-linejoin=""/></svg>`
	icon, err := parseSVG("t.svg", []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if shape := icon.shapes[0]; shape.cap != render.LineCapButt || shape.join != render.LineJoinBevel {
		t.Errorf("cap %v join %v, want butt and bevel", shape.cap, shape.join)
	}
}
