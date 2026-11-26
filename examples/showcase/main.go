// Command showcase shows every widget of the toolkit, built in Go and from
// markup, with a choice of theme, contrast and scale. F3 shows the frame
// time and the draw counts. The renderer is chosen at run time:
//
//	go run ./examples/showcase -renderer=ebiten
//
// The user interface is package sections, which needs no display.
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/examples/showcase/sections"
)

func main() {
	log.SetFlags(0)
	name := flag.String("renderer", renderers[0].Name, "the renderer to draw with: ebiten or wgpu")
	flag.Parse()
	if err := run(*name); err != nil {
		log.Fatal(err)
	}
}

func run(renderer string) error {
	opened, err := openRenderer(renderer)
	if err != nil {
		return err
	}
	showcase := sections.New(time.Now)
	ctx, err := core.NewCtx(showcase.Theme(), opened.Device)
	if err != nil {
		return fmt.Errorf("start: %w", err)
	}
	ctx.Stats = &core.DrawStats{}
	ctx.Clipboard = core.SystemClipboard()
	return opened.Run(newHost(showcase, ctx, opened.ReadInput))
}
