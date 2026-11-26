// Package site is the wiebkereister.de project: baseplate with the site's
// plugin and theme.
package site

import (
	"embed"
	"io/fs"

	"github.com/janmarkuslanger/baseplate"
	"github.com/janmarkuslanger/baseplate/fields"
	"github.com/janmarkuslanger/baseplate/plugins/access"
	"github.com/janmarkuslanger/baseplate/plugins/localbusiness"
	"github.com/janmarkuslanger/baseplate/plugins/navigation"
)

//go:embed theme
var themeFS embed.FS

// New returns the configured app; cfg.Theme is set by New.
func New(cfg baseplate.Config) *baseplate.App {
	theme, err := fs.Sub(themeFS, "theme")
	if err != nil {
		panic(err) // the embedded directory always exists
	}
	cfg.Theme = theme
	return baseplate.New(cfg).Use(
		fields.CoreBlocks{Only: []string{"heading", "richtext", "image"}},
		navigation.New(),
		access.New(),
		localbusiness.New(),
		Plugin{},
	)
}
