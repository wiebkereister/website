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
data directory, port and systemd service, behind Caddy:

| | dev | prod |
|---|---|---|
| Trigger | push/merge to `main` | published GitHub release |
| URL (now) | `http://<IP>:8081` | `http://<IP>` |
| URL (later) | `https://dev.wiebkereister.de` | `https://wiebkereister.de` |
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
5. Create the admin users, per environment:
   `sudo -u deploy env $(cat /etc/wiebkereister/dev.env | xargs) /opt/wiebkereister/dev/current/wiebkereister user add you@example.com`

### Switching to the domains

The Caddyfile already contains the domains, so the only step is DNS: point the
A records of `dev`, then `wiebkereister.de` and `www` to the server's IP.
Caddy fetches the Let's Encrypt certificates on its own as soon as a record
resolves. Leave e-mail records (MX, SPF, DKIM) untouched, and remove any old
AAAA record or point it at the server's IPv6 address.

Afterwards set `BASEPLATE_SECURE_COOKIES=1` in `/etc/wiebkereister/<env>.env`
(`sudo systemctl restart wiebkereister@<env>`); with it on, the admin only
works over HTTPS, no longer via the IP. Then close port 8081 (`ufw delete
allow 8081/tcp`). Re-run `bootstrap.sh` to roll out a changed Caddyfile.

### DNS-Umstellung (Anweisung für die bisherige Agentur)

Bitte nur diese Einträge der Zone `wiebkereister.de` ändern:

| Name | Typ | Neuer Wert | Bisher |
|---|---|---|---|
| `dev.wiebkereister.de` | A | `92.205.56.194` | `185.30.32.67` |
| `wiebkereister.de` | A | `92.205.56.194` | `185.30.32.67` |
| `www.wiebkereister.de` | A | `92.205.56.194` | `185.30.32.67` |

Optional (IPv6), jeweils für dieselben drei Namen: Typ `AAAA`, Wert
`2a00:1169:111:c8c0::`. Aktuell gibt es keine AAAA-Einträge, es reicht also,
die drei A-Einträge zu ändern.

**Bitte unverändert lassen:**

- alle `MX`-, `SPF`-/`TXT`- und `DKIM`-Einträge (die E-Mail-Postfächer laufen
  weiter beim bisherigen Hoster),
- die Nameserver (`ns1`–`ns4.webgo.de`); die Domain bleibt dort, es ändern sich
  nur die drei Einträge oben.

**Ablauf:**

1. Etwa 24 Stunden vorher die TTL der drei Einträge auf 300 Sekunden senken.
2. Zuerst nur `dev.wiebkereister.de` umstellen. Wir prüfen, ob die Seite
   unter `https://dev.wiebkereister.de` mit gültigem Zertifikat läuft.
3. Nach unserer Freigabe `wiebkereister.de` und `www` umstellen. Das
   HTTPS-Zertifikat holt der Server selbst, auf eurer Seite ist dafür nichts zu tun.
4. Rückgängig machen geht jederzeit, indem die Einträge wieder auf
   `185.30.32.67` gesetzt werden.
