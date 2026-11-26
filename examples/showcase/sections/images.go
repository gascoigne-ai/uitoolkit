package sections

import (
	"image"
	"image/color"
	"slices"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/markup"
	"github.com/gascoigne-ai/uitoolkit/render"
	. "github.com/gascoigne-ai/uitoolkit/view"
	. "github.com/gascoigne-ai/uitoolkit/widget"
)

// Sizes of the images section.
const (
	pictureW, pictureH = 480, 240 // the picture's own pixels
	largeW, largeH     = 320, 160
	mediumW, mediumH   = 160, 80
	smallW, smallH     = 80, 40
	iconsPerRow        = 7
	iconCellW          = 104
	iconSide           = 24
	checkerSide        = 24
)

// drawPicture paints the demo picture: a gradient under a checkerboard, with
// a disc, so scaling and filtering show.
func drawPicture() *image.RGBA {
	picture := image.NewRGBA(image.Rect(0, 0, pictureW, pictureH))
	centerX, centerY, radius := pictureW*3/4, pictureH/2, pictureH/3
	for y := range pictureH {
		for x := range pictureW {
			paint := color.RGBA{R: uint8(40 + 200*x/pictureW), G: uint8(40 + 160*y/pictureH), B: 160, A: 255}
			if (x/checkerSide+y/checkerSide)%2 == 0 {
				paint.B = 100
			}
			if (x-centerX)*(x-centerX)+(y-centerY)*(y-centerY) < radius*radius {
				paint = color.RGBA{R: 250, G: 214, B: 110, A: 255}
			}
			picture.SetRGBA(x, y, paint)
		}
	}
	return picture
}

// picture is the demo picture as ctx's device draws it, made once for a
// device.
func (s *Showcase) picture(ctx *core.Ctx) render.Image {
	demo := &s.widgets.demo
	if demo.source == nil {
		demo.source = drawPicture()
	}
	if demo.picture == nil || demo.device != ctx.Device {
		demo.device, demo.picture = ctx.Device, ctx.Device.NewImage(demo.source)
	}
	return demo.picture
}

// iconRows are the names of every registered icon, cut into grid rows.
func iconRows() [][]string { return slices.Collect(slices.Chunk(IconNames(), iconsPerRow)) }

func imagesGo(s *Showcase, ctx *core.Ctx) Node {
	picture := s.picture(ctx)
	return Scroll(&s.widgets.imagesPage, PadOf(core.GutterOf), GapOf(core.GutterOf),
		Caption("One picture, scaled to fit"),
		Row(H(largeH), GapOf(core.GutterOf), AlignItems(AlignStart),
			Image(picture, W(largeW), H(largeH)),
			Image(picture, W(mediumW), H(mediumH)),
			Image(picture, W(smallW), H(smallH)),
		),
		Caption("Icons"),
		For(iconRows(), func(_ int, names []string) Node {
			return Row(GapOf(core.LooseGapOf),
				For(names, func(_ int, name string) Node {
					return Col(W(iconCellW), GapOf(core.NearGapOf), AlignItems(AlignCenter),
						Icon(name, W(iconSide), H(iconSide), Color(core.Text)),
						Text(name, Small(), Muted()),
					)
				}),
			)
		}),
	)
}

func imagesMarkup(s *Showcase, ctx *core.Ctx) Node {
	return fill(ctx, imagesTemplate, markup.Scope{
		"page": &s.widgets.imagesPage, "picture": s.picture(ctx), "rows": iconRows(),
		"largeW": largeW, "largeH": largeH, "mediumW": mediumW, "mediumH": mediumH, "smallW": smallW, "smallH": smallH,
		"cellW": iconCellW, "iconSide": iconSide,
	})
}

var imagesTemplate = markup.Must("images.vml", imagesSource)

const imagesSource = `
<scroll state="page" pad="{{gutter}}" gap="{{gutter}}">
  <caption>One picture, scaled to fit</caption>
  <row h="{{largeH}}" gap="{{gutter}}" align-items="start">
    <image img="picture" w="{{largeW}}" h="{{largeH}}"/>
    <image img="picture" w="{{mediumW}}" h="{{mediumH}}"/>
    <image img="picture" w="{{smallW}}" h="{{smallH}}"/>
  </row>
  <caption>Icons</caption>
  <row for="names in rows" gap="{{looseGap}}">
    <col for="name in names" w="{{cellW}}" gap="{{nearGap}}" align-items="center">
      <icon name="{{name}}" w="{{iconSide}}" h="{{iconSide}}" color="text"/>
      <text small muted>{{name}}</text>
    </col>
  </row>
</scroll>`
