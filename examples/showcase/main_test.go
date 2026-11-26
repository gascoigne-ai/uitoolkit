package main

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// testGame runs the tests inside the first Update, because Ebitengine reads
// pixels only inside its run loop, so no test here may call t.Parallel. It
// is the showcase's own copy of the renderer's, which is not exported.
type testGame struct {
	suite *testing.M
	code  int
}

func (g *testGame) Update() error {
	g.code = g.suite.Run()
	return ebiten.Termination
}

func (*testGame) Draw(*ebiten.Image) {}

func (*testGame) Layout(int, int) (int, int) { return 320, 240 }

// sandbox points HOME and the XDG directories into root, so no test reads or
// writes the files of the user who runs it.
func sandbox(root string) error {
	for name, dir := range map[string]string{
		"HOME":            "home",
		"XDG_CONFIG_HOME": "config",
		"XDG_DATA_HOME":   "data",
		"XDG_CACHE_HOME":  "cache",
		"XDG_STATE_HOME":  "state",
	} {
		path := filepath.Join(root, dir)
		if err := os.MkdirAll(path, 0o700); err != nil {
			return err
		}
		if err := os.Setenv(name, path); err != nil {
			return err
		}
	}
	return nil
}

func TestMain(suite *testing.M) {
	root, err := os.MkdirTemp("", "showcase-test-")
	if err != nil {
		log.Fatal(err)
	}
	if err := sandbox(root); err != nil {
		log.Fatal(err)
	}
	loop := &testGame{suite: suite, code: 1}
	err = ebiten.RunGame(loop)
	// os.Exit skips deferred calls, so the sandbox is removed here.
	if removeErr := os.RemoveAll(root); removeErr != nil {
		log.Print(removeErr)
	}
	if err != nil {
		panic(err)
	}
	if loop.code != 0 {
		os.Exit(loop.code)
	}
}
