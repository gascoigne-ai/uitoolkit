package ebitenrender

import (
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// ebitenKey is the ebiten key of a render key; known is false when there is
// none, since ebiten's zero key is a real key.
type ebitenKey struct {
	key   ebiten.Key
	known bool
}

// keyTables maps the keys both ways.
type keyTables struct {
	fromEbiten [ebiten.KeyMax + 1]render.Key
	toEbiten   [render.KeyMax + 1]ebitenKey
}

// keys is built once, when the package starts, and never written after. The
// two key types are matched by the names their String methods return, because
// users store shortcuts by those names; the numbers may differ.
var keys = buildKeyTables()

func buildKeyTables() keyTables {
	named := map[string]render.Key{}
	for key := render.KeyUnknown + 1; key <= render.KeyMax; key++ {
		named[key.String()] = key
	}
	var tables keyTables
	for key := ebiten.Key(0); key <= ebiten.KeyMax; key++ {
		match, ok := named[key.String()]
		if !ok {
			continue
		}
		tables.fromEbiten[key] = match
		tables.toEbiten[match] = ebitenKey{key: key, known: true}
	}
	return tables
}

// KeyFromEbiten is the render key of the same name as key, and KeyUnknown for
// a value outside 0 to ebiten.KeyMax.
func KeyFromEbiten(key ebiten.Key) render.Key {
	if key < 0 || key > ebiten.KeyMax {
		return render.KeyUnknown
	}
	return keys.fromEbiten[key]
}

// KeyToEbiten is the ebiten key of the same name as key. It reports false for
// KeyUnknown and for a value above render.KeyMax.
func KeyToEbiten(key render.Key) (ebiten.Key, bool) {
	if key > render.KeyMax {
		return 0, false
	}
	found := keys.toEbiten[key]
	return found.key, found.known
}

// cursorShapes is the ebiten shape of each render cursor.
var cursorShapes = map[render.Cursor]ebiten.CursorShapeType{
	render.CursorDefault:    ebiten.CursorShapeDefault,
	render.CursorResizeH:    ebiten.CursorShapeEWResize,
	render.CursorResizeV:    ebiten.CursorShapeNSResize,
	render.CursorResizeNWSE: ebiten.CursorShapeNWSEResize,
	render.CursorResizeNESW: ebiten.CursorShapeNESWResize,
	render.CursorMove:       ebiten.CursorShapeMove,
}
