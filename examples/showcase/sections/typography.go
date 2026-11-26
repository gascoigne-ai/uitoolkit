package sections

import (
	"fmt"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/markup"
	. "github.com/gascoigne-ai/uitoolkit/view"
	. "github.com/gascoigne-ai/uitoolkit/widget"
)

// Texts and sizes of the typography section.
const (
	faceLabelW = 110
	proseMaxW  = 480 // a reading width: the page's whole width makes lines too long to follow
	sample     = "Sphinx of black quartz, judge my vow. 0O 1lI {}[] ->"
	prose      = "A paragraph is muted small text that breaks into lines at the width its parent gives it, " +
		"so a description sits beside a control without a column count guessed in advance. " +
		"Text is measured by the face that draws it, so the lines break in the same places on every renderer."
	ruleW     = 3  // the gap between two hairline posts
	postH     = 24 // the height of a hairline post
	postCount = 40
)

// dpiLines say what the display's scale does to the sizes.
func dpiLines(ctx *core.Ctx) []string {
	return []string{
		fmt.Sprintf("Device pixels per logical pixel: %.2f", ctx.Scale),
		fmt.Sprintf("Line heights in logical pixels: interface %d, small %d, mono %d",
			ctx.LineH(ctx.F.UI), ctx.LineH(ctx.F.Small), ctx.LineH(ctx.F.Mono)),
		"The rules and posts below stay whole pixels at every scale:",
	}
}

// proseNode wraps at a reading width, narrower than the page.
func proseNode() Node { return Paragraph(prose, MaxW(proseMaxW)) }

func typographyGo(s *Showcase, ctx *core.Ctx) Node {
	face := func(label string, opts ...Option) Node {
		return Row(GapOf(core.GutterOf), AlignItems(AlignBaseline),
			Text(label, W(faceLabelW), Small(), Muted()),
			Text(sample, opts...),
		)
	}
	posts := make([]int, postCount)
	return Scroll(&s.widgets.typographyPage, PadOf(core.GutterOf), GapOf(core.GutterOf),
		Caption("Faces"),
		face("Interface"),
		face("Small", Small()),
		face("Mono", Mono()),
		face("Bold", Bold()),
		face("Small bold", Small(), Bold()),
		Caption("Wrapped text"),
		proseNode(),
		Caption("DPI"),
		For(dpiLines(ctx), func(_ int, line string) Node { return Text(line) }),
		HRule(),
		HRule(),
		Row(H(postH), Gap(ruleW),
			For(posts, func(int, int) Node { return VRule(Bg(core.Text)) }),
		),
	)
}

func typographyMarkup(s *Showcase, ctx *core.Ctx) Node {
	return fill(ctx, typographyTemplate, markup.Scope{
		"page": &s.widgets.typographyPage, "labelW": faceLabelW, "sample": sample, "paragraph": proseNode(),
		"lines": dpiLines(ctx), "posts": make([]int, postCount), "postH": postH, "postGap": ruleW,
	})
}

var typographyTemplate = markup.Must("typography.vml", typographySource)

const typographySource = `
<scroll state="page" pad="{{gutter}}" gap="{{gutter}}">
  <caption>Faces</caption>
  <row gap="{{gutter}}" align-items="baseline"><text w="{{labelW}}" small muted>Interface</text><text>{{sample}}</text></row>
  <row gap="{{gutter}}" align-items="baseline"><text w="{{labelW}}" small muted>Small</text><text small>{{sample}}</text></row>
  <row gap="{{gutter}}" align-items="baseline"><text w="{{labelW}}" small muted>Mono</text><text mono>{{sample}}</text></row>
  <row gap="{{gutter}}" align-items="baseline"><text w="{{labelW}}" small muted>Bold</text><text bold>{{sample}}</text></row>
  <row gap="{{gutter}}" align-items="baseline"><text w="{{labelW}}" small muted>Small bold</text><text small bold>{{sample}}</text></row>
  <caption>Wrapped text</caption>
  <slot name="paragraph"/>
  <caption>DPI</caption>
  <text for="line in lines">{{line}}</text>
  <hrule/>
  <hrule/>
  <row h="{{postH}}" gap="{{postGap}}">
    <vrule for="post in posts" bg="text"/>
  </row>
</scroll>`
