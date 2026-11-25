// Command wiebkereister runs the CMS of wiebkereister.de (see README).
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/janmarkuslanger/baseplate"

	"wiebkereister.de/site"
)

func main() {
	app := site.New(baseplate.Config{})
	if len(os.Args) > 1 && os.Args[1] == "seed" {
		if err := seed(app); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Println("seeded initial content from ./seed")
		return
	}
	app.Main()
}

// seed imports ./seed into an empty data directory.
func seed(app *baseplate.App) error {
	ctx := context.Background()
	if err := app.Setup(ctx); err != nil {
		return err
	}
	defer app.Close()
	return site.Seed(ctx, app.Registry, os.DirFS("seed"))
}
