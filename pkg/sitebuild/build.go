// Package sitebuild compiles a site's Handlebars templates and renders its pages.
//
// It is the library API behind the `hbc build` command. Applications embedding
// go-hbars should call Run directly instead of locating and executing an hbc
// binary.
package sitebuild

import "github.com/andriyg76/go-hbars/internal/buildcmd"

// Options configures template compilation and site generation.
type Options = buildcmd.Options

// Result contains the one-time compilation/build costs and render duration.
type Result = buildcmd.Result

// Run compiles the templates, builds a temporary renderer, generates the site,
// and removes the temporary build directory.
func Run(opts Options) (Result, error) {
	return buildcmd.Run(opts)
}
