package site

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/janmarkuslanger/baseplate/entries"
	"github.com/janmarkuslanger/baseplate/fields"
	"github.com/janmarkuslanger/baseplate/plugin"
	"github.com/janmarkuslanger/baseplate/plugins/navigation"
	"github.com/janmarkuslanger/baseplate/storage"
)

// Prefixes of references in content.json: "media:<file>" names a file in
// media/, which Seed uploads and replaces with its /media/ URL; "page:<slug>"
// names a seeded page and becomes its ID (for relations).
const (
	mediaPrefix = "media:"
	pagePrefix  = "page:"
)

// ErrNotEmpty is returned when the site already has pages.
var ErrNotEmpty = errors.New("seed: the site already has pages; seed only an empty data directory")

type seedFile struct {
	Settings    map[string]any   `json:"settings"`
	Pages       []map[string]any `json:"pages"`
	Navigations []map[string]any `json:"navigations"`
}

// Seed imports the initial content of the site from dir (content.json and
// the images in media/) into an empty site: the settings, all pages
// (published) and the navigations, whose items point to the pages. It
// refuses to run when a page exists, so it never duplicates or overwrites
// content editors made.
func Seed(ctx context.Context, reg *plugin.Registry, dir fs.FS) error {
	if _, n, err := reg.Store.ListEntries(ctx, storage.ListQuery{Type: fields.PageType, Version: storage.Draft, Limit: 1}); err != nil {
		return err
	} else if n > 0 {
		return ErrNotEmpty
	}
	raw, err := fs.ReadFile(dir, "content.json")
	if err != nil {
		return err
	}
	var sf seedFile
	if err := json.Unmarshal(raw, &sf); err != nil {
		return fmt.Errorf("seed: content.json: %w", err)
	}
	r := resolver{reg: reg, dir: dir, urls: map[string]string{}, pages: map[string]int64{}}
	if _, err := r.save(ctx, fields.SettingsType, sf.Settings); err != nil {
		return fmt.Errorf("seed: settings: %w", err)
	}
	for _, p := range sf.Pages {
		id, err := r.save(ctx, fields.PageType, p)
		if err != nil {
			return fmt.Errorf("seed: page %q: %w", p["slug"], err)
		}
		r.pages[str(p["slug"])] = id
	}
	for _, n := range sf.Navigations {
		if _, err := r.save(ctx, navigation.TypeName, n); err != nil {
			return fmt.Errorf("seed: navigation %q: %w", n["key"], err)
		}
	}
	return nil
}

// resolver turns the references in seed data into stored values.
type resolver struct {
	reg   *plugin.Registry
	dir   fs.FS
	urls  map[string]string // file name → uploaded URL
	pages map[string]int64  // slug → ID of the seeded page
}

// save resolves the references in data and saves it published.
func (u *resolver) save(ctx context.Context, typ string, data map[string]any) (int64, error) {
	if err := u.resolve(data); err != nil {
		return 0, err
	}
	res, err := u.reg.Entries.Save(ctx, entries.Save{Type: typ, Data: data, Publish: true})
	if err != nil {
		return 0, err
	}
	return res.Entry.ID, nil
}

// resolve replaces "media:" and "page:" references in v (recursively);
// each file is uploaded once.
func (u *resolver) resolve(v map[string]any) error {
	for k, val := range v {
		nv, err := u.value(val)
		if err != nil {
			return err
		}
		v[k] = nv
	}
	return nil
}

func (u *resolver) value(v any) (any, error) {
	switch x := v.(type) {
	case string:
		if name, ok := strings.CutPrefix(x, mediaPrefix); ok {
			return u.upload(name)
		}
		if slug, ok := strings.CutPrefix(x, pagePrefix); ok {
			id, ok := u.pages[slug]
			if !ok {
				return nil, fmt.Errorf("seed: unknown page %q (pages are seeded before the navigations)", slug)
			}
			return id, nil
		}
	case map[string]any:
		return x, u.resolve(x)
	case []any:
		for i, it := range x {
			nv, err := u.value(it)
			if err != nil {
				return nil, err
			}
			x[i] = nv
		}
	}
	return v, nil
}

func (u *resolver) upload(name string) (string, error) {
	if url, ok := u.urls[name]; ok {
		return url, nil
	}
	f, err := u.dir.Open(path.Join("media", name))
	if err != nil {
		return "", fmt.Errorf("seed: %w", err)
	}
	defer f.Close()
	url, err := u.reg.Media.Save(f, name)
	if err != nil {
		return "", fmt.Errorf("seed: upload %s: %w", name, err)
	}
	u.urls[name] = url
	return url, nil
}
