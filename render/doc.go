// Package render is the contract between the toolkit and a renderer: the
// surface the toolkit draws on, the images and fonts a renderer provides, the
// keys and cursor shapes a window reports and takes, and the program a window
// runs. It imports only the standard library, so that neither side depends on
// a GPU library through it.
package render
