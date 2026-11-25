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

## Deployment

Two environments run on one server (Host Europe VPS), each with its own
data directory, port and systemd service, behind Caddy (automatic HTTPS):

| | dev | prod |
|---|---|---|
| Trigger | push/merge to `main` | published GitHub release |
| URL | `https://dev.wiebkereister.de` | `https://wiebkereister.de` |
| Service | `wiebkereister@dev` | `wiebkereister@prod` |
| Data | `/var/lib/wiebkereister/dev` | `/var/lib/wiebkereister/prod` |

Content is edited in each environment's admin (`/admin/`), so the data
directories are **not** touched by deployments; only the binary is replaced.
The first deployment seeds an empty data directory. A failing health check
rolls back to the previous release (the last 5 are kept in
`/opt/wiebkereister/<env>/releases`).

### Setup (once)

1. Create a deploy key: `ssh-keygen -t ed25519 -f deploy_key -N "" -C github-deploy`
2. On the server as root: `bash deploy/bootstrap.sh "$(cat deploy_key.pub)"`
3. In GitHub, create the environments `dev` and `prod` (add required
   reviewers on `prod` if wanted) and set in each the secrets
   `DEPLOY_HOST` (IP/hostname), `DEPLOY_SSH_KEY` (private key) and
   `DEPLOY_KNOWN_HOSTS` (`ssh-keyscan -t ed25519 <host>`).
4. Merge to `main` → dev is live. Create a release → prod is live.
5. Create the admin users, per environment (the password is read from stdin):

   ```sh
   sudo -u deploy bash -c 'cd /opt/wiebkereister/dev/current && set -a && . /etc/wiebkereister/dev.env && set +a && ./wiebkereister user add you@example.com'
   ```
