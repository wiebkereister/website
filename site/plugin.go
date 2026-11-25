package site

import (
	"html/template"
	"net/url"
	"strings"
	"time"

	"github.com/janmarkuslanger/baseplate/content"
	"github.com/janmarkuslanger/baseplate/fields"
	"github.com/janmarkuslanger/baseplate/plugin"
)

// Plugin adds the site's content blocks, the contact and menu teaser
// settings and the template functions the theme uses.
type Plugin struct{}

func (Plugin) Name() string { return "wiebkereister" }

func (Plugin) Register(r *plugin.Registry) error {
	for _, b := range blocks() {
		if err := r.Block(b); err != nil {
			return err
		}
	}
	if err := r.AllowBlocks(fields.PageType, "body", bodyBlocks...); err != nil {
		return err
	}
	if err := r.Extend(fields.SettingsType, content.Field{
		Name: "contact", Label: "Kontaktdaten", Kind: fields.Group, Tab: "Kontakt",
		Fields: []content.Field{
			{Name: "locationName", Label: "Praxisort", Kind: fields.Text},
			{Name: "locationURL", Label: "Link zum Praxisort", Kind: fields.URL},
			{Name: "street", Label: "Straße", Kind: fields.Text},
			{Name: "city", Label: "PLZ und Ort", Kind: fields.Text},
			{Name: "phone", Label: "Telefon", Kind: fields.Text, Help: "Anzeige, z. B. „0163 – 28 90 824“; der Link wird daraus erzeugt"},
			{Name: "email", Label: "E-Mail", Kind: fields.Email},
			{Name: "hours", Label: "Zeiten", Kind: fields.Textarea, Help: "Eine Angabe pro Zeile"},
			{Name: "footerText", Label: "Text im Footer", Kind: fields.Textarea},
		},
	}, content.Field{
		Name: "menuTeaser", Label: "Teaser im Mega-Menü", Kind: fields.Group, Tab: "Navigation",
		Help: "Erscheint im Untermenü, dessen Unterpunkte in Spalten gruppiert sind",
		Fields: []content.Field{
			{Name: "title", Label: "Titel", Kind: fields.Text},
			{Name: "text", Label: "Text", Kind: fields.Textarea},
			{Name: "label", Label: "Button-Beschriftung", Kind: fields.Text},
			{Name: "url", Label: "Button-Link", Kind: fields.URL},
		},
	}); err != nil {
		return err
	}
	return r.Funcs(plugin.TemplateFuncs{
		"telHref":      telHref,
		"mapsURL":      mapsURL,
		"mapsEmbedURL": mapsEmbedURL,
		"lines":        lines,
		"paragraphs":   paragraphs,
		"year":         func() int { return time.Now().Year() },
	})
}

// telHref turns a displayed German phone number into a tel: link
// ("0163 – 28 90 824" → "tel:+491632890824"). Empty input gives "".
// The result holds only "tel:", "+" and digits, so it is marked safe for
// html/template, which otherwise rejects the tel: scheme.
func telHref(phone any) template.URL {
	var digits strings.Builder
	for _, c := range str(phone) {
		if c >= '0' && c <= '9' || c == '+' && digits.Len() == 0 {
			digits.WriteRune(c)
		}
	}
	n := digits.String()
	switch {
	case n == "":
		return ""
	case strings.HasPrefix(n, "00"):
		n = "+" + n[2:]
	case strings.HasPrefix(n, "0"):
		n = "+49" + n[1:]
	}
	return template.URL("tel:" + n)
}

// mapsURL opens a Google Maps search for query in a new tab.
func mapsURL(query string) string {
	return "https://www.google.com/maps/search/?" + url.Values{"api": {"1"}, "query": {query}}.Encode()
}

// mapsEmbedURL is the iframe source of the two-click map.
func mapsEmbedURL(query string) string {
	return "https://maps.google.com/maps?" + url.Values{"q": {query}, "output": {"embed"}}.Encode()
}

// lines splits a textarea value into its non-empty lines.
func lines(v any) []string {
	var out []string
	for _, l := range strings.Split(str(v), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

// paragraphs splits a textarea value at blank lines.
func paragraphs(v any) []string {
	var out []string
	for _, p := range strings.Split(strings.ReplaceAll(str(v), "\r\n", "\n"), "\n\n") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
