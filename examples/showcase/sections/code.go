package sections

import (
	"fmt"
	"image"
	"strings"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/markup"
	. "github.com/gascoigne-ai/uitoolkit/view"
	. "github.com/gascoigne-ai/uitoolkit/widget"
)

// demoFuncs is how many functions the demo source has, about a dozen lines
// each: a file long enough to stress scrolling.
const demoFuncs = 320

// The posed view: a folded block deep in the file, in the middle of the view.
const (
	posedFunc = 150
)

// Where lines are in the demo source.
const (
	headerLines = 2  // before the first function
	funcLines   = 11 // each function
	branchLine  = 3  // from a function's first line to its if block
)

// scripts put glyphs from several scripts into the comments, so scrolling
// asks the renderer for many different ones.
var scripts = []string{
	"quick brown fox jumps over the lazy dog 0123456789",
	"Zażółć gęślą jaźń, naïve café, Straße",
	"Ταχίστη αλώπηξ βαφής ψημένη γη",
	"Съешь же ещё этих мягких французских булок",
	"<= >= != == -> := && || {} [] () ~ ^ % @ # $ &",
}

// demoSource is the long file of the code view, the same on every call.
func demoSource() []Line {
	lines := []Line{
		codeLine(0, Span{Text: "package ", Style: core.AccentLine}, Span{Text: "demo", Style: core.Text}),
		codeLine(0),
	}
	for number := range demoFuncs {
		lines = append(lines, demoFunc(number)...)
	}
	return lines
}

func demoFunc(number int) []Line {
	name := fmt.Sprintf("step%03d", number)
	open := codeLine(0,
		Span{Text: "func ", Style: core.AccentLine}, Span{Text: name, Style: core.Info},
		Span{Text: "(values []int, limit int) int {", Style: core.Text})
	open.Foldable = true
	branch := codeLine(1, Span{Text: "if ", Style: core.AccentLine}, Span{Text: fmt.Sprintf("len(values) > %d {", number%17), Style: core.Text})
	branch.Foldable = true
	return []Line{
		codeLine(0, Span{Text: "// " + name + ": " + scripts[number%len(scripts)], Style: core.Muted}),
		open,
		codeLine(1, Span{Text: "total := ", Style: core.Text}, Span{Text: fmt.Sprint(number * 31 % 101), Style: core.OK}),
		branch,
		codeLine(2, Span{Text: "values = values[:", Style: core.Text}, Span{Text: fmt.Sprint(number % 17), Style: core.OK}, Span{Text: "]", Style: core.Text}),
		codeLine(1, Span{Text: "}", Style: core.Text}),
		codeLine(1, Span{Text: "for ", Style: core.AccentLine}, Span{Text: "_, value := range values { total = (total + value) % limit }", Style: core.Text}),
		codeLine(1, Span{Text: "log.Printf(", Style: core.Text}, Span{Text: `"` + strings.Repeat("a long line that scrolls sideways ", 1+number%3) + `%d"`, Style: core.OK}, Span{Text: ", total)", Style: core.Text}),
		codeLine(1, Span{Text: "return ", Style: core.AccentLine}, Span{Text: "total", Style: core.Text}),
		codeLine(0, Span{Text: "}", Style: core.Text}),
		codeLine(0),
	}
}

// codeLine is one line indented by depth steps. The indent is spaces, not
// tabs: CodeView finds the end of a foldable block by counting the leading
// spaces of a line's first span.
func codeLine(depth int, spans ...Span) Line {
	if depth > 0 {
		spans = append([]Span{{Text: strings.Repeat(" ", depth*core.TabStop), Style: core.Text}}, spans...)
	}
	return Line{Spans: spans, Depth: depth}
}

var codePoses = []Pose{{Apply: func(s *Showcase, _ *core.Ctx) image.Point {
	block := headerLines + posedFunc*funcLines + branchLine
	s.widgets.code.ToggleFold(block)
	s.widgets.code.RevealLine(block)
	return OffCanvas
}}}

func (s *Showcase) codeCount() string { return fmt.Sprintf("%d lines", len(s.widgets.code.Lines)) }

func codeGo(s *Showcase, _ *core.Ctx) Node {
	state := &s.widgets
	return Col(Grow(1),
		Toolbar(
			Text(s.codeCount(), Muted()),
			Spacer(),
			CheckRow(state.wrapCode, OnClick(flip(&state.wrapCode)), AlignItems(AlignCenter), Text("Wrap long lines")),
		),
		Code(&state.code, Grow(1), Bg(core.Sunken), When(state.wrapCode, Wrap())),
	)
}

func codeMarkup(s *Showcase, ctx *core.Ctx) Node {
	state := &s.widgets
	return fill(ctx, codeTemplate, markup.Scope{
		"code": &state.code, "count": s.codeCount(), "wrapCode": state.wrapCode, "flipWrap": flip(&state.wrapCode),
	})
}

var codeTemplate = markup.Must("code.vml", codeSource)

const codeSource = `
<col grow="1">
  <toolbar>
    <text muted>{{count}}</text>
    <spacer/>
    <check-row on="{{wrapCode}}" on-click="flipWrap" align-items="center"><text>Wrap long lines</text></check-row>
  </toolbar>
  <code state="code" grow="1" bg="sunken" wrap="{{wrapCode}}"/>
</col>`
