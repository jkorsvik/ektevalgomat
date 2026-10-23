// Package web embeds ektevalgomat's templates and static assets into the
// Go binary at build time: the final image is FROM scratch with nothing
// on disk but one binary,
// no runtime file I/O, no "inside" to get a shell in.
package web

import "embed"

//go:embed templates
var Templates embed.FS

//go:embed static
var Static embed.FS
