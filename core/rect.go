// Package core is the base of the UI toolkit: theme tokens, a drawing context
// that works in logical pixels and draws through a render.Canvas, input and
// the clipboard. It knows nothing about the application; package view builds
// declarative views on top of it. While callers still hold ebiten images, one
// file, ebitenbridge.go, turns them into canvases.
package core

// Rect is an axis-aligned pixel rectangle.
type Rect struct{ X, Y, W, H int }

// Contains reports whether the point is inside r.
func (r Rect) Contains(posX, posY int) bool {
	return posX >= r.X && posX < r.X+r.W && posY >= r.Y && posY < r.Y+r.H
}
