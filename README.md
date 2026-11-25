# wiebkereister.de

Website of the private physiotherapy practice Wiebke Reister, built on
[baseplate](https://github.com/janmarkuslanger/baseplate): one Go binary with
the admin, and a static build of the site into `dist/`.

## Quickstart

Requires Go ≥ 1.26.

```sh
go run . seed                        # import the initial content (empty data dir only)
go run . user add you@example.com    # create an admin user (password from stdin)
go run . serve                       # admin at http://localhost:8080/admin/
go run . build                       # static site in ./dist
```

All baseplate commands and environment variables apply (see the baseplate
README), e.g. `BASEPLATE_DATA_DIR`, `BASEPLATE_DIST_DIR`,
`BASEPLATE_BUILD_ON_PUBLISH=1`.

## Layout

| Path | Contents |
|---|---|
| `main.go` | CLI: baseplate's commands plus `seed` |
| `site/site.go` | app setup: core blocks (heading, richtext, image), navigation, access protection, site plugin, theme |
| `site/blocks.go` | content blocks: `section`, `buttons`, `logos`, `faq`, `cards`, `map`, `contact` |
| `site/plugin.go` | block and type registration, contact fields in the settings (tab "Kontakt"), template functions |
| `site/seed.go` | one-time import of `seed/` |
| `site/theme/` | layout (header, navigation, footer), block templates, CSS/JS, self-hosted fonts |
| `seed/content.json` | settings, the eight pages and the navigations, taken over from the previous WordPress site |
| `seed/media/` | images referenced as `media:<file>` in `content.json` |

## Content model

- Pages are baseplate's core `page` type. Their body offers `section`, `faq`,
  `cards`, `map` and `contact`; a `section` holds `heading`, `richtext`,
  `image`, `buttons` and `logos`.
- **Navigations** come from baseplate's `plugins/navigation` (admin:
  *Navigationen* and *Links*; a link has a URL or a page and child links):
  `main` (header and footer page list), `cta` (header/footer button) and
  `legal` (footer). A `main` link whose children have children renders as
  mega menu, with those children as column headings; its teaser is set in
  **Einstellungen → Navigation**.
- Contact data (address, phone, e-mail, hours, footer text) lives in
  **Einstellungen → Kontakt** and feeds the footer and the `contact` block.
- **Password protection** comes from baseplate's `plugins/access` (admin:
  *Zugriffsschutz*): one username and password for the site and `/media/`,
  checked with Basic Auth while `serve` runs. The admin stays open. It does
  not apply to a static host in front of `dist/`.
- The `map` block loads Google Maps only after a click. The site sets no
  cookies and makes no third-party requests before that click.

## Seeding

`seed` refuses to run when a page exists, so it never duplicates or
overwrites edited content. If it fails part-way, delete the data directory
and run it again.

## Tests

```sh
go test ./...
```
