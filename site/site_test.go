package site

import (
	"context"
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/janmarkuslanger/baseplate"
)

func setup(t *testing.T) (*baseplate.App, string) {
	t.Helper()
	for _, k := range []string{"BASEPLATE_DATA_DIR", "BASEPLATE_DIST_DIR"} {
		t.Setenv(k, "") // the config of the test, not the developer's shell
	}
	dir := t.TempDir()
	dist := filepath.Join(dir, "dist")
	app := New(baseplate.Config{DataDir: filepath.Join(dir, "data"), DistDir: dist})
	if err := app.Setup(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.Close() })
	return app, dist
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestSeedAndBuild imports the real seed content and builds the site.
func TestSeedAndBuild(t *testing.T) {
	ctx := context.Background()
	app, dist := setup(t)
	if err := Seed(ctx, app.Registry, os.DirFS("../seed")); err != nil {
		t.Fatal(err)
	}
	res, err := app.Builder.Build(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.Pages != 8 {
		t.Errorf("built %d pages, want 8", res.Pages)
	}

	home := read(t, filepath.Join(dist, "index.html"))
	for _, want := range []string{
		`<h1 class="section-title">Willkommen in der Privatpraxis für Physiotherapie</h1>`,
		`<title>Physiotherapie Engelskirchen – Privatpraxis Wiebke Reister</title>`,
		`href="tel:&#43;491632890824"`,                                  // footer, from the contact settings
		`<a href="/" class="active" aria-current="page">Startseite</a>`, // navigation marks the current page
		`src="/media/praxis-behandlung-`,                                // seeded image was uploaded
		`id="kontakt"`,
	} {
		if !strings.Contains(home, want) {
			t.Errorf("index.html lacks %s", want)
		}
	}

	for _, want := range []string{
		`<a href="/physiotherapie/#leistungen">Lymphdrainage</a>`, // nested list link to a page
		`<a href="https://villa-adebar.de/yoga/">Yoga in der Rückbildung</a>`,
		`<p class="submenu-title">Sprechstunde: Beckenboden / Rektusdiastase</p>`, // teaser from the settings
		`<a class="button header-cta" href="/kontakt/">Terminvereinbarung</a>`,
		`<a href="/datenschutz/">Datenschutz</a>`,
	} {
		if !strings.Contains(home, want) {
			t.Errorf("index.html lacks navigation link %s", want)
		}
	}

	physio := read(t, filepath.Join(dist, "physiotherapie", "index.html"))
	for _, want := range []string{`id="leistungen"`, `id="faq1"`, `<summary>Brauche ich für die Behandlung ein Rezept?</summary>`} {
		if !strings.Contains(physio, want) {
			t.Errorf("physiotherapie/index.html lacks %s", want)
		}
	}
	if !strings.Contains(physio, `<a href="/physiotherapie/" class="active" aria-current="page">Physiotherapie</a>`) {
		t.Error("physiotherapie page is not marked in the navigation")
	}

	kontakt := read(t, filepath.Join(dist, "kontakt", "index.html"))
	if strings.Contains(kontakt, "<iframe") {
		t.Error("the map must not load before the visitor clicks")
	}
	if !strings.Contains(kontakt, `data-map-src="https://maps.google.com/maps?output=embed&amp;q=Villa&#43;Adebar`) {
		t.Error("map block lacks its embed URL")
	}

	md := read(t, filepath.Join(dist, "physiotherapie.md"))
	if !strings.Contains(md, "### Was ist eine Rektusdiastase?") {
		t.Error("Markdown variant lacks the FAQ")
	}

	// A second run must not duplicate content.
	if err := Seed(ctx, app.Registry, os.DirFS("../seed")); !errors.Is(err, ErrNotEmpty) {
		t.Errorf("second seed: err = %v, want ErrNotEmpty", err)
	}
}

func TestSeedMissingMedia(t *testing.T) {
	app, _ := setup(t)
	dir := fstest.MapFS{"content.json": {Data: []byte(`{"settings":{"siteName":"x","baseURL":"https://example.org"},
		"pages":[{"title":"A","slug":"index","body":[{"_type":"section","title":"A","image":"media:missing.jpg"}]}]}`)}}
	err := Seed(context.Background(), app.Registry, dir)
	if err == nil || !strings.Contains(err.Error(), "missing.jpg") {
		t.Fatalf("err = %v, want an error naming missing.jpg", err)
	}
}

func TestSeedInvalidContent(t *testing.T) {
	app, _ := setup(t)
	dir := fstest.MapFS{"content.json": {Data: []byte(`{"settings":{"siteName":"x","baseURL":"https://example.org"},
		"pages":[{"title":"A","slug":"index","body":[{"_type":"section"}]}]}`)}}
	err := Seed(context.Background(), app.Registry, dir)
	if err == nil || !strings.Contains(err.Error(), `page "index"`) {
		t.Fatalf("err = %v, want a validation error for page index", err)
	}
}

func TestTelHref(t *testing.T) {
	for in, want := range map[string]template.URL{
		"0163 – 28 90 824":   "tel:+491632890824",
		"+49 163 2890824":    "tel:+491632890824",
		"0049 (163) 2890824": "tel:+491632890824",
		"":                   "",
		"keine Nummer":       "",
	} {
		if got := telHref(in); got != want {
			t.Errorf("telHref(%q) = %q, want %q", in, got, want)
		}
	}
	if got := telHref(nil); got != "" {
		t.Errorf("telHref(nil) = %q", got)
	}
}

func TestParagraphsAndLines(t *testing.T) {
	if got, want := paragraphs("a\nb\r\n\r\n\n c \n\n"), []string{"a\nb", "c"}; !slices.Equal(got, want) {
		t.Errorf("paragraphs = %q, want %q", got, want)
	}
	if got, want := lines(" a \n\n b"), []string{"a", "b"}; !slices.Equal(got, want) {
		t.Errorf("lines = %q, want %q", got, want)
	}
	if got := lines(nil); got != nil {
		t.Errorf("lines(nil) = %q", got)
	}
}

func TestSeedUnknownPage(t *testing.T) {
	app, _ := setup(t)
	dir := fstest.MapFS{"content.json": {Data: []byte(`{"settings":{"siteName":"x","baseURL":"https://example.org"},"pages":[],
		"navigations":[{"name":"N","key":"main","items":[{"_type":"navitem","title":"A","page":"page:nope"}]}]}`)}}
	err := Seed(context.Background(), app.Registry, dir)
	if err == nil || !strings.Contains(err.Error(), `"nope"`) {
		t.Fatalf("err = %v, want an error naming the unknown page", err)
	}
}

// The structured data of baseplate v0.10: WebSite and the practice (from
// plugins/localbusiness) on the home page, breadcrumbs on the others.
func TestStructuredData(t *testing.T) {
	ctx := context.Background()
	app, dist := setup(t)
	if err := Seed(ctx, app.Registry, os.DirFS("../seed")); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Builder.Build(ctx); err != nil {
		t.Fatal(err)
	}
	home := read(t, filepath.Join(dist, "index.html"))
	for _, want := range []string{
		`"@type":"WebSite"`,
		`"@type":"Physiotherapy"`,
		`"telephone":"+491632890824"`,
		`"addressLocality":"Engelskirchen"`,
		`"openingHours":["Mo-Fr 08:00-20:00"]`,
	} {
		if !strings.Contains(home, want) {
			t.Errorf("index.html lacks %s", want)
		}
	}
	if strings.Contains(home, `"@type":"Organization"`) {
		t.Error("the organization is the Physiotherapy node")
	}
	if page := read(t, filepath.Join(dist, "physiotherapie", "index.html")); !strings.Contains(page, `"@type":"BreadcrumbList"`) {
		t.Error("physiotherapie/index.html lacks the breadcrumbs")
	}
}
