package site

import (
	"fmt"
	"strings"

	"github.com/janmarkuslanger/baseplate/content"
	"github.com/janmarkuslanger/baseplate/fields"
	"github.com/janmarkuslanger/baseplate/markdown"
)

// Blocks a page body offers. Section children are limited to the simple
// content blocks so the page structure stays one level deep.
var (
	bodyBlocks    = []string{"section", "faq", "cards", "map", "contact"}
	sectionBlocks = []string{"heading", "richtext", "image", "buttons", "logos"}
)

func opts(pairs ...string) []content.Option {
	out := make([]content.Option, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		out = append(out, content.Option{Value: pairs[i], Label: pairs[i+1]})
	}
	return out
}

func anchorField() content.Field {
	return content.Field{Name: "anchor", Label: "Anker (ID)", Kind: fields.Slug, Help: "Optional, z. B. „leistungen“ für Links auf /seite/#leistungen"}
}

func blocks() []content.Block {
	return []content.Block{
		{
			Name: "section", Label: "Abschnitt",
			Fields: []content.Field{
				{Name: "kicker", Label: "Dachzeile", Kind: fields.Text},
				{Name: "title", Label: "Überschrift", Kind: fields.Text, Required: true},
				{Name: "level", Label: "Überschrift-Ebene", Kind: fields.Select, Options: opts("h2", "H2", "h1", "H1 (Seitenkopf)", "h3", "H3")},
				{Name: "children", Label: "Inhalt", Kind: fields.Blocks, Blocks: sectionBlocks},
				{Name: "image", Label: "Bild", Kind: fields.Image},
				{Name: "alt", Label: "Alternativtext", Kind: fields.Text},
				{Name: "imageSide", Label: "Bildposition", Kind: fields.Select, Options: opts("right", "rechts", "left", "links")},
				{Name: "tone", Label: "Hintergrund", Kind: fields.Select, Options: opts("plain", "weiß", "muted", "grau", "cta", "Anfrage-Box (zentriert)")},
				anchorField(),
			},
			Markdown: func(d map[string]any) string {
				prefix := "## "
				if d["level"] == "h3" {
					prefix = "### "
				}
				md := prefix + markdown.EscapeText(str(d["title"]))
				if src := str(d["image"]); src != "" {
					md += "\n\n" + mdImage(str(d["alt"]), src)
				}
				return md
			},
		},
		{
			Name: "buttons", Label: "Buttons",
			Fields: []content.Field{
				{Name: "items", Label: "Buttons", Kind: fields.List, Required: true, Fields: []content.Field{
					{Name: "label", Label: "Beschriftung", Kind: fields.Text, Required: true},
					{Name: "href", Label: "Link", Kind: fields.URL, Required: true},
				}},
			},
			Markdown: func(d map[string]any) string {
				var links []string
				for _, it := range items(d) {
					links = append(links, mdLink(str(it["label"]), str(it["href"])))
				}
				return strings.Join(links, " · ")
			},
		},
		{
			Name: "logos", Label: "Logos",
			Fields: []content.Field{
				{Name: "items", Label: "Logos", Kind: fields.List, Required: true, Fields: []content.Field{
					{Name: "image", Label: "Bild", Kind: fields.Image, Required: true},
					{Name: "alt", Label: "Alternativtext", Kind: fields.Text, Required: true},
					{Name: "href", Label: "Link", Kind: fields.URL},
				}},
			},
		},
		{
			Name: "faq", Label: "Häufige Fragen",
			Fields: []content.Field{
				{Name: "title", Label: "Überschrift", Kind: fields.Text, Required: true},
				{Name: "items", Label: "Fragen", Kind: fields.List, Required: true, Fields: []content.Field{
					{Name: "question", Label: "Frage", Kind: fields.Text, Required: true},
					{Name: "answer", Label: "Antwort", Kind: fields.Textarea, Required: true},
				}},
				anchorField(),
			},
			Markdown: func(d map[string]any) string {
				md := "## " + markdown.EscapeText(str(d["title"]))
				for _, it := range items(d) {
					md += "\n\n### " + markdown.EscapeText(str(it["question"])) + "\n\n" + markdown.EscapeText(str(it["answer"]))
				}
				return md
			},
		},
		{
			Name: "cards", Label: "Karten",
			Fields: []content.Field{
				{Name: "items", Label: "Karten", Kind: fields.List, Required: true, Fields: []content.Field{
					{Name: "title", Label: "Titel", Kind: fields.Text, Required: true},
					{Name: "image", Label: "Bild", Kind: fields.Image},
					{Name: "alt", Label: "Alternativtext", Kind: fields.Text},
					{Name: "text", Label: "Text", Kind: fields.Textarea},
					{Name: "more", Label: "Weiterer Text (aufklappbar)", Kind: fields.Textarea},
					{Name: "linkLabel", Label: "Link-Beschriftung", Kind: fields.Text},
					{Name: "href", Label: "Link", Kind: fields.URL},
					anchorField(),
				}},
			},
			Markdown: func(d map[string]any) string {
				var parts []string
				for _, it := range items(d) {
					md := "### " + markdown.EscapeText(str(it["title"]))
					for _, k := range []string{"text", "more"} {
						if t := str(it[k]); t != "" {
							md += "\n\n" + markdown.EscapeText(t)
						}
					}
					if h := str(it["href"]); h != "" {
						md += "\n\n" + mdLink(str(it["linkLabel"]), h)
					}
					parts = append(parts, md)
				}
				return strings.Join(parts, "\n\n")
			},
		},
		{
			// map shows a Google Maps embed only after the visitor clicks
			// (no request to Google before consent).
			Name: "map", Label: "Karte (Zwei-Klick)",
			Fields: []content.Field{
				{Name: "title", Label: "Titel", Kind: fields.Text, Required: true},
				{Name: "address", Label: "Adresszeile", Kind: fields.Text},
				{Name: "query", Label: "Suchbegriff für Google Maps", Kind: fields.Text, Required: true, Help: "z. B. „Villa Adebar, Oststrasse 16, 51766 Engelskirchen“"},
			},
			Markdown: func(d map[string]any) string {
				md := "## " + markdown.EscapeText(str(d["title"]))
				if a := str(d["address"]); a != "" {
					md += "\n\n" + markdown.EscapeText(a)
				}
				return md + "\n\n" + mdLink("Route in Google Maps", mapsURL(str(d["query"])))
			},
		},
		{
			// contact renders the contact data from the site settings.
			Name: "contact", Label: "Kontaktdaten (aus Einstellungen)",
		},
	}
}

func str(v any) string { s, _ := v.(string); return s }

func items(d map[string]any) []map[string]any {
	list, _ := d["items"].([]any)
	out := make([]map[string]any, 0, len(list))
	for _, it := range list {
		if m, ok := it.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func mdLink(label, href string) string {
	return fmt.Sprintf("[%s](%s)", markdown.EscapeText(label), markdown.EscapeURL(href))
}

func mdImage(alt, src string) string {
	return fmt.Sprintf("![%s](%s)", markdown.EscapeText(alt), markdown.EscapeURL(src))
}
